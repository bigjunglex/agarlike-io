extends Node2D

const packets := preload("res://packets.gd")
@onready var _hiscores: Hiscores = $UI/VBoxContainer/Hiscores
@onready var _back: Button = $UI/VBoxContainer/Back


func _ready() -> void:
	WS.packet_received.connect(_on_ws_packet_recived)
	_back.button_down.connect(_on_back_pressed)
	
	var packet := packets.Packet.new()
	packet.new_hiscore_board_request()
	WS.send(packet)
	
func _on_ws_packet_recived(packet: packets.Packet) -> void:
	if packet.has_hiscore_board():
		_handle_hiscore_board_message(packet.get_hiscore_board())
 

func _handle_hiscore_board_message(msg: packets.HiscoreBoardMessage) -> void:
	for s: packets.HiscoreMessage in msg.get_hiscores():
		@warning_ignore("shadowed_variable_base_class")
		var name := "%d. %s" % [s.get_rank(), s.get_name()]
		var score := s.get_score()
		_hiscores.set_hiscore(name, score)
		
		
func _on_back_pressed() -> void:
	var packet := packets.Packet.new()
	packet.new_menu_request()
	WS.send(packet)
	GameManager.set_state(GameManager.State.CONNECTED)
