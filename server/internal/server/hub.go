package server

import (
	"agar-server/pkg/packets"
	"agar-server/internal/server/objects"
	"log"
	"net/http"
)

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

	Close(reason string)
}

type Hub struct {
	Clients        *objects.SharedCollection[ClientInterfacer]
	BroadcastChan  chan *packets.Packet
	RegisterChan   chan ClientInterfacer
	UnregisterChan chan ClientInterfacer
}

func NewHub() *Hub {
	return &Hub{
		Clients:        objects.NewSharedCollection[ClientInterfacer](),
		BroadcastChan:  make(chan *packets.Packet),
		RegisterChan:   make(chan ClientInterfacer),
		UnregisterChan: make(chan ClientInterfacer),
	}
}

func (h *Hub) Run() {
	log.Println("Client registrations ... ")

	for {
		select {
		case client := <-h.RegisterChan:
			client.Initialize(h.Clients.Add(client))
		case client := <-h.UnregisterChan:
			h.Clients.Remove(client.Id())
		case packet := <-h.BroadcastChan:
			h.Clients.ForEach(func(id uint64, client ClientInterfacer){
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
