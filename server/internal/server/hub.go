package server

import (
	"agar-server/internal/server/db"
	"agar-server/internal/server/objects"
	"agar-server/pkg/packets"
	"context"
	"database/sql"
	_ "embed"
	"log"
	"net/http"

	_ "modernc.org/sqlite"
)

//go:embed db/config/schema.sql
var schemaGenSql string

type DbTx struct {
	Ctx     context.Context
	Queries *db.Queries
}

type ClientStateHandler interface {
	Name() string

	OnEnter()

	SetClient(client ClientInterfacer)
	HandleMessage(senderId uint64, msg packets.Msg)

	OnExit()
}

type ClientInterfacer interface {
	Initialize(id uint64)
	Id() uint64

	SetState(newState ClientStateHandler)

	ProcessMessage(sender_id uint64, msg packets.Msg)

	//message from this client to write pump
	SocketSend(message packets.Msg)

	//message from another client to write pump
	SocketSendAs(message packets.Msg, senderId uint64)

	//forward message to client
	PassToPeer(message packets.Msg, peerId uint64)
	Broadcast(message packets.Msg)

	// data from socket to client
	ReadPump()

	// data from client to socket
	WritePump()

	DbTx() *DbTx

	SharedGameObjects() *SharedGameObjects

	Close(reason string)
}

type Hub struct {
	Clients           *objects.SharedCollection[ClientInterfacer]
	BroadcastChan     chan *packets.Packet
	RegisterChan      chan ClientInterfacer
	UnregisterChan    chan ClientInterfacer
	dbPool            *sql.DB
	SharedGameObjects *SharedGameObjects
}

func (h *Hub) NewDbTx() *DbTx {
	return &DbTx{
		Ctx:     context.Background(),
		Queries: db.New(h.dbPool),
	}
}

type SharedGameObjects struct {
	//client id = player id
	Players *objects.SharedCollection[*objects.Player]
}

func NewHub() *Hub {
	dbPool, err := sql.Open("sqlite", "db.sqlite")
	if err != nil {
		log.Fatalf("Error opening database %v", err)
	}
	return &Hub{
		Clients:        objects.NewSharedCollection[ClientInterfacer](),
		BroadcastChan:  make(chan *packets.Packet),
		RegisterChan:   make(chan ClientInterfacer),
		UnregisterChan: make(chan ClientInterfacer),
		dbPool:         dbPool,
		SharedGameObjects: &SharedGameObjects{
			Players: objects.NewSharedCollection[*objects.Player](),
		},
	}
}

func (h *Hub) Run() {
	log.Println("Initializing database ... ")
	_, err := h.dbPool.ExecContext(context.Background(), schemaGenSql)

	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}

	log.Println("Client registrations ... ")

	for {
		select {
		case client := <-h.RegisterChan:
			client.Initialize(h.Clients.Add(client))
		case client := <-h.UnregisterChan:
			h.Clients.Remove(client.Id())
		case packet := <-h.BroadcastChan:
			h.Clients.ForEach(func(id uint64, client ClientInterfacer) {
				if id != packet.SenderId {
					client.ProcessMessage(packet.SenderId, packet.Msg)
				}
			})
		}
	}
}

func (h *Hub) Serve(
	getNewClient func(*Hub, http.ResponseWriter, *http.Request) (ClientInterfacer, error),
	writer http.ResponseWriter,
	req *http.Request,
) {
	log.Println("[HUB]: New Client Connection :", req.RemoteAddr)
	client, err := getNewClient(h, writer, req)

	if err != nil {
		log.Printf("[HUB]: Error with proccesing client %v", err)
		return
	}

	if client == nil {
		log.Printf("[HUB]: client is nil")
		return
	}

	h.RegisterChan <- client
	go client.WritePump()
	go client.ReadPump()
}
