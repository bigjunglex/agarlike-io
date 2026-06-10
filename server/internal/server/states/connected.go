package states

import (
	"agar-server/internal/server"
	"agar-server/pkg/packets"
	"fmt"
	"log"
)

type Connected struct {
	client server.ClientInterfacer
	logger *log.Logger
}

func (c *Connected) Name() string {
	return "Connected"
}
func (c *Connected) SetClient(client server.ClientInterfacer) {
	c.client = client
	logPrefix := fmt.Sprintf("[Client]: %d [%s]: ", client.Id(), c.Name())
	c.logger = log.New(log.Writer(), logPrefix, log.LstdFlags)
}

func (c *Connected) OnEnter() {
	c.client.SocketSend(packets.NewId(c.client.Id()))
}

func (c *Connected) HandleMessage(senderId uint64, msg packets.Msg) {
	if senderId == c.client.Id() {
		c.client.Broadcast(msg)
	} else {
		c.client.SocketSendAs(msg, senderId)
	}
}

func (c *Connected) OnExit() {}