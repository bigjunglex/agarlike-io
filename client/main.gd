extends Node

const packets := preload("res://packets.gd")

func _ready() -> void:
	var packet := packets.Packet.new()
	packet.set_sender_id(777)
	var chat_msg := packet.new_chat()
	print(" --- *** ---- ")
	print("\n")
	
	chat_msg.set_msg("Client Init")
	print(" --- Packet Created ---- ")
	print(packet)
	
	var data = packet.to_bytes()
	print(" --- Bytes Send ---- \n")
	print(data)
	print("\n")
	
	
	var recieved := packets.Packet.new()
	recieved.from_bytes(data)
	print(" --- Bytes Read ---- \n")
	print(recieved)
	print("\n")
	print(" --- *** ---- ")
