package states

import (
	"agar-server/internal/server"
	"agar-server/internal/server/db"
	"agar-server/internal/server/objects"
	"agar-server/pkg/packets"
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type Connected struct {
	client  server.ClientInterfacer
	logger  *log.Logger
	queries *db.Queries
	dbCtx   context.Context
}

func (c *Connected) Name() string {
	return "Connected"
}
func (c *Connected) SetClient(client server.ClientInterfacer) {
	c.client = client
	logPrefix := fmt.Sprintf("[Client]: %d [%s]: ", client.Id(), c.Name())
	c.logger = log.New(log.Writer(), logPrefix, log.LstdFlags)
	c.queries = client.DbTx().Queries
	c.dbCtx = client.DbTx().Ctx
}

func (c *Connected) OnEnter() {
	c.client.SocketSend(packets.NewId(c.client.Id()))

}

func (c *Connected) HandleMessage(senderId uint64, msg packets.Msg) {
	switch msg := msg.(type) {
	case *packets.Packet_LoginRequest:
		c.handleLoginRequest(senderId, msg)
	case *packets.Packet_RegisterRequest:
		c.handleRegisterRequest(senderId, msg)
	// case *packets.Packet_Chat:
	// 	c.handleChatMessage(senderId, msg)
	}
}

// func (c *Connected) handleChatMessage(senderId uint64, msg *packets.Packet_Chat) {
// 	c.logger.Printf("chat message from %d: %v", senderId, msg.Chat.Msg)
// 	if senderId == c.client.Id() {
// 		c.client.Broadcast(msg)
// 	} else {
// 		c.client.SocketSendAs(msg, senderId)
// 	}
// }

func (c *Connected) handleLoginRequest(senderId uint64, msg *packets.Packet_LoginRequest) {
	if senderId != c.client.Id() {
		c.logger.Printf("Invalid login request id [ID]: %d", senderId)
		return
	}

	username := msg.LoginRequest.Username
	failMsg := packets.NewDenyResponse("Incorrect username or password")

	user, err := c.queries.GetUserByUsername(c.dbCtx, strings.ToLower(username))
	if err != nil {
		c.logger.Printf("Error during logging")
		c.client.SocketSend(failMsg)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(msg.LoginRequest.Password))
	if err != nil {
		c.logger.Printf("Incorrect password")
		c.client.SocketSend(failMsg)
		return
	}

	c.logger.Printf("[LOGIN]: %s logged in", username)
	c.client.SocketSend(packets.NewOkResponse())

	c.client.SetState(&InGame{
		player: &objects.Player{
			Name: username,
		},
	})
}

func (c *Connected) handleRegisterRequest(senderId uint64, msg *packets.Packet_RegisterRequest) {
	if senderId != c.client.Id() {
		c.logger.Printf("Invalid login request id [ID]: %d", senderId)
		return
	}

	username := msg.RegisterRequest.Username
	err := validateUsername(username)

	if err != nil {
		reason := fmt.Sprintf("Invalid username: %v", err)
		c.logger.Println(reason)
		c.client.SocketSend(packets.NewDenyResponse(reason))
	}

	if _, err := c.queries.GetUserByUsername(c.dbCtx, username); err == nil {
		c.logger.Printf("[REGISTRATION]: %s user already exists", username)
		c.client.SocketSend(packets.NewDenyResponse("User already exists"))
		return
	}

	failMsg := packets.NewDenyResponse("Failed to register user (internal server error)")

	passHash, err := bcrypt.GenerateFromPassword([]byte(msg.RegisterRequest.Password), bcrypt.DefaultCost)
	if err != nil {
		c.logger.Printf("[BCRYPT]: failed to hash password: %v", err)
		c.client.SocketSend(failMsg)
		return
	}

	_, err = c.queries.CreateUser(c.dbCtx, db.CreateUserParams{
		Username:     strings.ToLower(username),
		PasswordHash: string(passHash),
	})

	if err != nil {
		c.logger.Printf("[REGISTRATION]: failed to create user: %v", err)
		c.client.SocketSend(failMsg)
		return
	}

	c.logger.Printf("User %s registered", username)
	c.client.SocketSend(packets.NewOkResponse())
}

func (c *Connected) OnExit() {}

func validateUsername(username string) error {
	if len(username) <= 0 {
		return errors.New("empty")
	}

	if len(username) > 20 {
		return errors.New("too long")
	}

	if username != strings.TrimSpace(username) {
		return errors.New("leading or trailing white space")
	}

	return nil
}
