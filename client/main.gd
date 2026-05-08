extends Node

var packets := preload("res://packets.gd")

func _ready() -> void:
	WS.connected_to_server.connect(_on_ws_connected_to_server)
	WS.connection_closed.connect(_on_ws_connection_closed)
	WS.packet_received.connect(_on_ws_message)
	
	
func  _on_ws_connected_to_server() -> void:
	var packet := packets.Packet.new()
	var msg := packet.new_chat()
	msg.set_msg("Hello from godot!")
	var err := WS.send(packet)
	if err:
		print("[ERROR]: sendig %s", packet)

func _on_ws_connection_closed() -> void:
	print("Disconnected from the server")

func _on_ws_message() -> void:
	
