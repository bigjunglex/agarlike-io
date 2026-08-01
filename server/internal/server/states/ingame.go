package states

import (
	"agar-server/internal/server"
	"agar-server/internal/server/objects"
	"agar-server/pkg/packets"
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"
)

type InGame struct {
	client                 server.ClientInterfacer
	player                 *objects.Player
	logger                 *log.Logger
	cancelPlayerUpdateLoop context.CancelFunc
}

func (g *InGame) Name() string {
	return "InGame"
}

func (g *InGame) SetClient(client server.ClientInterfacer) {
	g.client = client
	logPrefix := fmt.Sprintf("[CLIENT] [%s]id -(%d): ", g.Name(), client.Id())
	g.logger = log.New(log.Writer(), logPrefix, log.LstdFlags)
}

func (g *InGame) OnEnter() {
	g.logger.Printf("Player %s enetered game", g.player.Name)
	go g.client.SharedGameObjects().Players.Add(g.player, g.client.Id())

	g.player.X = rand.Float64() * 1000
	g.player.Y = rand.Float64() * 1000
	g.player.Speed = 150.0
	g.player.Radius = 20.0

	g.client.SocketSend(packets.NewPlayer(g.client.Id(), g.player))
}

func (g *InGame) HandleMessage(senderId uint64, msg packets.Msg) {
	switch msg := msg.(type) {
	case *packets.Packet_Player:
		g.handlePlayer(senderId, msg)
	case *packets.Packet_PlayerDirection:
		g.handleDirection(senderId, msg)
	case *packets.Packet_Chat:
		g.handleChat(senderId, msg)
	}
}

func (g *InGame) handleChat(senderId uint64, msg *packets.Packet_Chat) {
	if senderId == g.client.Id() {
		g.client.Broadcast(msg)
	} else {
		g.client.SocketSendAs(msg, senderId)
	}
}

func (g *InGame) handleDirection(senderId uint64, msg *packets.Packet_PlayerDirection) {
	if senderId == g.client.Id() {
		g.player.Direction = msg.PlayerDirection.Direction

		if g.cancelPlayerUpdateLoop == nil {
			ctx, cancel := context.WithCancel(context.Background())
			g.cancelPlayerUpdateLoop = cancel
			go g.updatePlayerLoop(ctx)
		}
	}
}

func (g *InGame) updatePlayerLoop(ctx context.Context) {
	const dt float64 = 0.05
	ticker := time.NewTicker(time.Duration(dt*1000) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			g.syncPlayer(dt)
		case <-ctx.Done():
			return
		}
	}
}

func (g *InGame) handlePlayer(senderId uint64, msg *packets.Packet_Player) {
	if senderId == g.client.Id() {
		g.logger.Printf("Recieved player message from our own cliend [%d], ignoring", senderId)
		return
	}

	g.client.SocketSendAs(msg, senderId)
}

func (g *InGame) syncPlayer(dt float64) {
	newX := g.player.X + g.player.Speed*math.Cos(g.player.Direction)*dt
	newY := g.player.Y + g.player.Speed*math.Sin(g.player.Direction)*dt
	g.player.X = newX
	g.player.Y = newY

	updatePacket := packets.NewPlayer(g.client.Id(), g.player)
	g.client.Broadcast(updatePacket)
	go g.client.SocketSend(updatePacket)
}

func (g *InGame) OnExit() {
	if g.cancelPlayerUpdateLoop != nil {
		g.cancelPlayerUpdateLoop()
	}
	g.client.SharedGameObjects().Players.Remove(g.client.Id())
}
