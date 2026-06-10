extends Node

const packets := preload("res://packets.gd")

@onready var _log: Log = $UI/Log
@onready var _line_edit: LineEdit = $UI/LineEdit

func _ready() -> void:
	WS.connection_closed.connect(_on_ws_connection_closed)
	WS.packet_received.connect(_on_ws_packet_recived)
	
	_line_edit.text_submitted.connect(_on_line_edit_submit)

func _on_ws_connection_closed() -> void:
	_log.warning("Disconnected from the server")

func _on_ws_packet_recived(packet: packets.Packet) -> void:
	var sender_id := packet.get_sender_id()
	if packet.has_chat():
		_handle_chat_packet(sender_id, packet.get_chat())
	
	
func _handle_chat_packet(sender_id: int, chat: packets.ChatMessage) -> void:
	_log.chat("[%d]" % sender_id, chat.get_msg())
	
func _on_line_edit_submit(new_text: String) -> void:
	var packet := packets.Packet.new()
	var chat_msg := packet.new_chat()
	chat_msg.set_msg(new_text)
	
	var err := WS.send(packet)
	if err:
		_log.error("[ERROR]: failed to send chat message")
	else:
		_log.chat("You", new_text)
	_line_edit.clear()
