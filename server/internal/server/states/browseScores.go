package states

import (
	"agar-server/internal/server"
	"agar-server/internal/server/db"
	"agar-server/pkg/packets"
	"context"
	"fmt"
	"log"
)

type BrowsingScores struct {
	client server.ClientInterfacer
	logger *log.Logger
	queris *db.Queries
	dbCtx  context.Context
}

func (b *BrowsingScores) Name() string {
	return "BrowsingScores"
}

func (b *BrowsingScores) SetClient(client server.ClientInterfacer) {
	b.client = client
	logPrefix := fmt.Sprintf("[CLIENT] [%s]id -(%d): ", b.Name(), client.Id())
	b.logger = log.New(log.Writer(), logPrefix, log.LstdFlags)
	b.queris = client.DbTx().Queries
	b.dbCtx = client.DbTx().Ctx
}

func (b *BrowsingScores) OnEnter() {
	const limit int64 = 10
	const offset int64 = 0

	topScores, err := b.queris.GetTopScores(b.dbCtx, db.GetTopScoresParams{
		Limit:  limit,
		Offset: offset,
	})

	if err != nil {
		b.logger.Printf("Error getting top %d scores from rank %d: %v", limit, offset, err)
		b.client.SocketSend(packets.NewDenyResponse("Failed to get top scores - please try again later"))
		return
	}
	hiscores := make([]*packets.HiscoreMessage, 0, limit)
	for rank, scoreRow := range topScores {
		hiscore_msg := &packets.HiscoreMessage{
			Rank:  uint64(rank) + uint64(offset) + 1,
			Name:  scoreRow.Name,
			Score: uint64(scoreRow.BestScore),
		}
		hiscores = append(hiscores, hiscore_msg)
	}

	b.client.SocketSend(packets.NewHiscoreBoard(hiscores))
}

func (b *BrowsingScores) HandleMessage(senderId uint64, msg packets.Msg) {
	switch msg := msg.(type) {
	case *packets.Packet_MenuRequest:
		b.hanldeMenuRequest(senderId, msg)
	}
}

func (b *BrowsingScores) hanldeMenuRequest(_ uint64, _ *packets.Packet_MenuRequest) {
	b.client.SetState(&Connected{})
}

func (b *BrowsingScores) OnExit() {

}
