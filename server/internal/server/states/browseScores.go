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
	client  server.ClientInterfacer
	logger  *log.Logger
	queries *db.Queries
	dbCtx   context.Context
}

func (b *BrowsingScores) Name() string {
	return "BrowsingScores"
}

func (b *BrowsingScores) SetClient(client server.ClientInterfacer) {
	b.client = client
	logPrefix := fmt.Sprintf("[CLIENT] [%s]id -(%d): ", b.Name(), client.Id())
	b.logger = log.New(log.Writer(), logPrefix, log.LstdFlags)
	b.queries = client.DbTx().Queries
	b.dbCtx = client.DbTx().Ctx
}

func (b *BrowsingScores) OnEnter() {
	b.sendTopScores(10, 0)
}

func (b *BrowsingScores) HandleMessage(senderId uint64, msg packets.Msg) {
	switch msg := msg.(type) {
	case *packets.Packet_MenuRequest:
		b.hanldeMenuRequest(senderId, msg)
	case *packets.Packet_SearchHistory:
		b.handleSearchScore(senderId, msg)
	}
}

func (b *BrowsingScores) hanldeMenuRequest(_ uint64, _ *packets.Packet_MenuRequest) {
	b.client.SetState(&Connected{})
}

func (b *BrowsingScores) handleSearchScore(_ uint64, msg *packets.Packet_SearchHistory) {
	p, err := b.queries.GetPlayerByName(b.dbCtx, msg.SearchHistory.Name)
	if err != nil {
		b.logger.Printf("Error getting player %s: %v", p.Name, err)
		b.client.SocketSend(packets.NewDenyResponse("No player found with that name"))
		return
	}

	pRank, err := b.queries.GetPlayerRank(b.dbCtx, p.ID)
	if err != nil {
		b.logger.Printf("Error getting rank for player %s: %v", p.Name, err)
		b.client.SocketSend(packets.NewDenyResponse("Player is unranked"))
		return
	}

	const limit int64 = 10
	offset := pRank - limit/2
	b.sendTopScores(limit, max(0, offset))
}

func (b *BrowsingScores) sendTopScores(limit int64, offset int64) {
	topScores, err := b.queries.GetTopScores(b.dbCtx, db.GetTopScoresParams{
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

func (b *BrowsingScores) OnExit() {

}
