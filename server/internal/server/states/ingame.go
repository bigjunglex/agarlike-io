package states

import (
	"agar-server/internal/server"
	"agar-server/internal/server/db"
	"agar-server/internal/server/objects"
	"agar-server/pkg/packets"
	"context"
	"fmt"
	"log"
	"math"
	"math/rand/v2"
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

	g.player.Speed = 150.0
	g.player.Radius = 20.0
	g.player.X, g.player.Y = objects.GetSpawnCoords(
		g.player.Radius,
		g.client.SharedGameObjects().Players,
		g.client.SharedGameObjects().Spores,
	)

	g.client.SocketSend(packets.NewPlayer(g.client.Id(), g.player))

	//cringe with batches
	go func() {
		const batchSize = 20
		sporesBatch := make(map[uint64]*objects.Spore, batchSize)

		g.client.SharedGameObjects().Spores.ForEach(func(spore_id uint64, spore *objects.Spore) {
			sporesBatch[spore_id] = spore
			if len(sporesBatch) >= batchSize {
				g.client.SocketSend(packets.NewSporesBatch(sporesBatch))
				sporesBatch = make(map[uint64]*objects.Spore, batchSize)
				time.Sleep(50 * time.Millisecond)
			}
		})

		if len(sporesBatch) > 0 {
			g.client.SocketSend(packets.NewSporesBatch(sporesBatch))
		}
	}()
}

func (g *InGame) HandleMessage(senderId uint64, msg packets.Msg) {
	switch msg := msg.(type) {
	case *packets.Packet_Player:
		g.handlePlayer(senderId, msg)
	case *packets.Packet_PlayerDirection:
		g.handleDirection(senderId, msg)
	case *packets.Packet_Chat:
		g.handleChat(senderId, msg)
	case *packets.Packet_SporeConsumed:
		g.handleSporeConsumed(senderId, msg)
	case *packets.Packet_PlayerConsumed:
		g.handlePlayerConsumed(senderId, msg)
	case *packets.Packet_Spore:
		g.handleSpore(senderId, msg)
	}
}

func (g *InGame) handlePlayerConsumed(senderId uint64, msg *packets.Packet_PlayerConsumed) {
	if senderId != g.client.Id() {
		g.client.SocketSendAs(msg, senderId)

		if msg.PlayerConsumed.PlayerId == g.client.Id() {
			g.logger.Println("Player consumed, respawning... ")
			g.client.SetState(&InGame{
				player: &objects.Player{
					Name: g.player.Name,
				},
			})
		}

		return
	}

	errMsg := "Could not verify player consumption"

	consumedId := msg.PlayerConsumed.PlayerId
	consumed, err := g.getPlayer(consumedId)
	if err != nil {
		g.logger.Println(errMsg + err.Error())
		return
	}

	aMass := radToMass(g.player.Radius)
	cMass := radToMass(consumed.Radius)

	if aMass <= cMass*1.5 {
		g.logger.Println(
			errMsg+"not big enough aR: [%f], cR: [%f]",
			g.player.Radius,
			consumed.Radius,
		)
		return
	}

	err = g.validatePlayerProximityToTarget(
		consumed.X,
		consumed.Y,
		consumed.Radius,
		15,
	)
	if err != nil {
		g.logger.Println(errMsg + err.Error())
		return
	}

	g.player.Radius = g.nextRadius(cMass)
	go g.client.SharedGameObjects().Players.Remove(consumedId)

	g.client.Broadcast(msg)
	go g.syncPlayerBestScore()
}

func (g *InGame) getPlayer(id uint64) (*objects.Player, error) {
	p, exists := g.client.SharedGameObjects().Players.Get(id)
	if !exists {
		return nil, fmt.Errorf("player with [ID]: [%d] does not exists", id)
	}

	return p, nil
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

func (g *InGame) handleSporeConsumed(senderId uint64, msg *packets.Packet_SporeConsumed) {
	if senderId != g.client.Id() {
		g.client.SocketSendAs(msg, senderId)
		return
	}

	errMsg := "[SPORE]: conumptions not verified"
	sporeId := msg.SporeConsumed.Id
	spore, err := g.getSpore(sporeId)
	if err != nil {
		g.logger.Println(errMsg + err.Error())
		return
	}

	err = g.validatePlayerProximityToTarget(spore.X, spore.Y, spore.Radius, 15)
	if err != nil {
		g.logger.Println(errMsg + err.Error())
		return
	}

	sporeMass := radToMass(spore.Radius)
	g.player.Radius = g.nextRadius(sporeMass)

	go g.client.SharedGameObjects().Spores.Remove(sporeId)
	g.client.Broadcast(msg)
	go g.syncPlayerBestScore()
}

func (g *InGame) handleSpore(senderId uint64, msg *packets.Packet_Spore) {
	g.client.SocketSendAs(msg, senderId)
}

func (g *InGame) syncPlayer(dt float64) {
	newX := g.player.X + g.player.Speed*math.Cos(g.player.Direction)*dt
	newY := g.player.Y + g.player.Speed*math.Sin(g.player.Direction)*dt
	bound := 4000.0

	if newX > bound {
		newX = bound
	}

	if newY > bound {
		newY = bound
	}

	g.player.X = newX
	g.player.Y = newY

	prob := g.player.Radius / float64(server.MaxSpores*5)
	if rand.Float64() < prob && g.player.Radius > 10 {
		spore := &objects.Spore{
			X:      g.player.X,
			Y:      g.player.Y,
			Radius: min(5+g.player.Radius/50, 15),
		}
		sporeId := g.client.SharedGameObjects().Spores.Add(spore)
		packet := packets.NewSpore(sporeId, spore)
		g.client.Broadcast(packet)
		go g.client.SocketSend(packet)
		g.player.Radius = g.nextRadius(-radToMass(spore.Radius))
	}

	updatePacket := packets.NewPlayer(g.client.Id(), g.player)
	g.client.Broadcast(updatePacket)
	go g.client.SocketSend(updatePacket)
}

func (g *InGame) OnExit() {
	if g.cancelPlayerUpdateLoop != nil {
		g.cancelPlayerUpdateLoop()
	}
	g.client.SharedGameObjects().Players.Remove(g.client.Id())
	go g.syncPlayerBestScore()
}

func (g *InGame) getSpore(sporeId uint64) (*objects.Spore, error) {
	spore, exists := g.client.SharedGameObjects().Spores.Get(sporeId)
	if !exists {
		return nil, fmt.Errorf("spore with ID %d does not exists", sporeId)
	}
	return spore, nil
}

func (g *InGame) validatePlayerProximityToTarget(tX, tY, tRadius, buffer float64) error {
	dx := g.player.X - tX
	dy := g.player.Y - tY
	dSq := dx*dx + dy*dy
	dThreshold := g.player.Radius + buffer + tRadius
	dSqThreshold := dThreshold * dThreshold

	if dSq > dSqThreshold {
		return fmt.Errorf(
			"player is to far from the object to interact (disSq: %f, threshold: %f)",
			dSq,
			dSqThreshold,
		)
	}

	return nil
}

func (g *InGame) syncPlayerBestScore() {
	currScore := int64(math.Round(radToMass(g.player.Radius)))
	if currScore > g.player.BestScore {
		g.player.BestScore = currScore
		err := g.client.DbTx().Queries.UpdatePlayerBestScore(
			g.client.DbTx().Ctx,
			db.UpdatePlayerBestScoreParams{
				ID:        g.player.DbId,
				BestScore: g.player.BestScore,
			},
		)
		if err != nil {
			g.logger.Printf("Error updating player best score: %v", err)
		}
	}
}

func (g *InGame) nextRadius(dMass float64) float64 {
	prevMass := radToMass(g.player.Radius)
	newMass := prevMass + dMass
	return massToRad(newMass)
}

func radToMass(r float64) float64 {
	return math.Pi * r * r
}

func massToRad(m float64) float64 {
	return math.Sqrt(m / math.Pi)
}
