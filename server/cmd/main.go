package main

import (
	"fmt"
	"agar-server/pkg/packets"
	"google.golang.org/protobuf/proto"
)

func main() {
	fmt.Println("Server running ....")

	packet := &packets.Packet{
		SenderId: 67,
		Msg: packets.NewChat("-- Init Packet --"),
	}

	fmt.Println("----------")
	fmt.Println("Created packet:")
	fmt.Println(packet)

	data, err := proto.Marshal(packet)
	if err != nil {
		panic(err)
	}

	fmt.Println(" ")
	fmt.Println("Serialized into:")
	fmt.Println(data)
	fmt.Println("----------")
}