package server

import (
	"agar-server/pkg/packets"
	"log"
	"net/http"
)

type ClientInterfacer interface {
	Initialize(id uint64)
	Id() uint64
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
	Clients        map[uint64]ClientInterfacer
	BroadcastChan  chan *packets.Packet
	RegisterChan   chan ClientInterfacer
	UnregisterChan chan ClientInterfacer
}

func NewHub() *Hub {
	return &Hub{
		Clients:        make(map[uint64]ClientInterfacer),
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
			client.Initialize(uint64(len(h.Clients)))
		case client := <-h.UnregisterChan:
			h.Clients[client.Id()] = nil
		case packet := <-h.BroadcastChan:
			for id, client := range h.Clients {
				if id != packet.SenderId {
					client.ProcessMessage(packet.SenderId, packet.Msg)
				}
			}
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
