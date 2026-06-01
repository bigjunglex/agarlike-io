extends Node

const packets := preload("res://packets.gd")
@onready var _log := $Log as Log

func _ready() -> void:
	WS.connected_to_server.connect(_on_ws_connected_to_server)
	WS.connection_closed.connect(_on_ws_connection_closed)
	WS.packet_received.connect(_on_ws_packet_recived)
	
	_log.info("Connecting to server...")
	WS.connect_to_url("ws://localhost:8075/ws")
	
func  _on_ws_connected_to_server() -> void:
	var packet := packets.Packet.new()
	var msg := packet.new_chat()
	msg.set_msg("Hello from godot!")
	var err := WS.send(packet)
	if err:
		_log.error("[ERROR]: sendig %s" % packet)
	else:
		_log.success("[OK]: %s send successfully" % packet)
		

func _on_ws_connection_closed() -> void:
	_log.warning("Disconnected from the server")

func _on_ws_packet_recived(packet: packets.Packet) -> void:
	_log.info("Recieved pakcet: %s" % packet )
