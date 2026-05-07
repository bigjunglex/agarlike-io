package packets

type Msg = isPacket_Msg

func NewChat(msg string) Msg {
	return &Packet_Chat{
		Chat: &ChatMessage{
			Msg: msg,
		},
	}
}

func NewId(id uint64) Msg {
	return  &Packet_Id{
		Id: &IdAssign{
			Id: id,
		},
	}
}

func NewDir(direction float32) Msg {
	return &Packet_Dir{
		Dir: &DirectionUpdate{
			Direction: direction,
		},
	}
}