extends Node

const packets := preload("res://packets.gd")
const Actor := preload("res://objects/actor/actor.gd")
const Spore := preload("res://objects/spore/spore.gd")

var _players: Dictionary[int, Actor]
var _spores: Dictionary[int, Spore]

@onready var _log: Log = $UI/Log
@onready var _line_edit: LineEdit = $UI/LineEdit
@onready var _world: Node2D = $World

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
	elif packet.has_player():
		_handle_player_packet(sender_id, packet.get_player())
	elif packet.has_spore():
		_handle_spore_packet(sender_id, packet.get_spore())
	
	
func _handle_chat_packet(sender_id: int, chat: packets.ChatMessage) -> void:
	var username := _players[sender_id].actor_name
	_log.chat("[%s]" % username, chat.get_msg())
	
	
func _handle_player_packet(_sender_id: int, player: packets.PlayerMessage) -> void:
	var actor_id := player.get_id()
	var actor_name := player.get_name()
	var x := player.get_x()
	var y := player.get_y()
	var radius := player.get_radius()
	var speed := player.get_speed()
	var direction := player.get_direction()
	var is_player := actor_id == GameManager.client_id
	
	if actor_id not in _players:
		var actor := Actor.instanciate(
			actor_id,
			actor_name,
			x,
			y,
			radius,
			direction,
			speed,
			is_player
		)
		_world.add_child(actor)
		_players[actor_id] = actor
	else:
		var actor := _players[actor_id]
		actor.position.x = x
		actor.position.y = y
		actor.velocity = speed * Vector2.from_angle(direction)
		

func _handle_spore_packet(sender_id: int, packet: packets.SporeMessage) -> void:
	var spore_id = packet.get_id()
	var x = packet.get_x()
	var y = packet.get_y()
	var radius = packet.get_radius()
	
	if spore_id not in _spores:
		var spore := Spore.instanciate(spore_id, x, y, radius)
		_world.add_child(spore)
		_spores[spore_id] = spore

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
