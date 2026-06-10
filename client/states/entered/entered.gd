extends Node

const packets := preload("res://packets.gd")

@onready var _log: Log = $UI/Log 

func _ready() -> void:
	WS.connected_to_server.connect(_on_ws_connected_to_server)
	WS.connection_closed.connect(_on_ws_connection_closed)
	WS.packet_received.connect(_on_ws_packet_recived)
	
	_log.info("Connecting to server...")
	WS.connect_to_url("ws://localhost:8075/ws")
	
func  _on_ws_connected_to_server() -> void:
	_log.success("Connected to server")

func _on_ws_connection_closed() -> void:
	_log.warning("Disconnected from the server")


func _on_ws_packet_recived(packet: packets.Packet) -> void:
	var sender_id := packet.get_sender_id()
	if packet.has_id():
		_handle_id_packet(sender_id, packet.get_id())


func _handle_id_packet(sender_id: int, id_assign: packets.IdAssign) -> void:
	if sender_id != 0:
		_log.error("[SERVER]: undefined msg")

	GameManager.client_id = id_assign.get_id()
	GameManager.set_state(GameManager.State.INGAME)
	_log.success("[ID] assigned: %d" % GameManager.client_id)
