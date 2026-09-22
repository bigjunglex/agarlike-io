extends Node2D

const packets := preload("res://packets.gd")
@onready var _hiscores: Hiscores = $UI/MarginContainer/VBoxContainer/Hiscores
@onready var _back: Button = $UI/MarginContainer/VBoxContainer/HBoxContainer/Back
@onready var _search_btn: Button = $UI/MarginContainer/VBoxContainer/HBoxContainer/Search
@onready var _line_edit: LineEdit = $UI/MarginContainer/VBoxContainer/HBoxContainer/LineEdit
@onready var _log: Log = $UI/MarginContainer/VBoxContainer/Log


func _ready() -> void:
	WS.packet_received.connect(_on_ws_packet_recived)
	_back.button_down.connect(_on_back_pressed)
	_line_edit.text_submitted.connect(_on_line_edit_submit)
	_search_btn.button_down.connect(_on_search_pressed)
	
	var packet := packets.Packet.new()
	packet.new_hiscore_board_request()
	WS.send(packet)
	
func _on_ws_packet_recived(packet: packets.Packet) -> void:
	if packet.has_hiscore_board():
		_handle_hiscore_board_message(packet.get_hiscore_board())
	elif packet.has_deny_response():
		_handle_deny_response(packet.get_deny_response())
 

func _handle_hiscore_board_message(msg: packets.HiscoreBoardMessage) -> void:
	_hiscores.clear_scores()
	for s: packets.HiscoreMessage in msg.get_hiscores():
		@warning_ignore("shadowed_variable_base_class")
		var name := "%d. %s" % [s.get_rank(), s.get_name()]
		var score := s.get_score()
		var highlight := s.get_name().to_lower() == _line_edit.text.to_lower()
		_hiscores.set_hiscore(name, score, highlight)
		
func _handle_deny_response(msg: packets.DenyResponseMessage) -> void:
	_log.error(msg.get_reason())
		
func _on_back_pressed() -> void:
	var packet := packets.Packet.new()
	packet.new_menu_request()
	WS.send(packet)
	GameManager.set_state(GameManager.State.CONNECTED)
	
func _on_line_edit_submit(_str: String) -> void:
	_on_search_pressed()
	

func _on_search_pressed() -> void:
	var packet := packets.Packet.new()
	var msg := packet.new_search_history()
	msg.set_name(_line_edit.text)
	WS.send(packet)
	
