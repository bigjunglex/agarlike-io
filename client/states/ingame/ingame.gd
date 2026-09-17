extends Node

const packets := preload("res://packets.gd")
const Actor := preload("res://objects/actor/actor.gd")
const Spore := preload("res://objects/spore/spore.gd")

var _players: Dictionary[int, Actor]
var _spores: Dictionary[int, Spore]

@onready var _log: Log =$UI/VBoxContainer/Log
@onready var _line_edit: LineEdit = $UI/VBoxContainer/LineEdit
@onready var _hiscores: Hiscores = $UI/VBoxContainer/Hiscores

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
	elif packet.has_spores_batch():
		_handle_spore_batch_packet(sender_id, packet.get_spores_batch())
	elif packet.has_spore_consumed():
		_handle_spore_conmsumed_packet(sender_id, packet.get_spore_consumed())
	
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
		_add_actor(
			actor_id,
			actor_name,
			x,
			y,
			radius,
			direction,
			speed,
			is_player
		)
	else:
		_update_actor(actor_id, x, y, radius, direction, speed, is_player)

func _handle_spore_packet(_sender_id: int, packet: packets.SporeMessage) -> void:
	var spore_id = packet.get_id()
	var x = packet.get_x()
	var y = packet.get_y()
	var radius = packet.get_radius()
	
	if spore_id not in _spores:
		var spore := Spore.instanciate(spore_id, x, y, radius)
		_world.add_child(spore)
		_spores[spore_id] = spore

func _handle_spore_batch_packet(sender_id: int, batch: packets.SporesBatchMessage) -> void:
	for spore_msg in batch.get_spores():
		_handle_spore_packet(sender_id, spore_msg)
	

func _handle_spore_conmsumed_packet(sender_id: int, packet: packets.SporeConsumedMessage) -> void: 
	if sender_id in _players:
		var actor := _players[sender_id]
		var actor_mass := _rad_to_mass(actor.radius)
		var spore_id := packet.get_id() 
		if spore_id in _spores:
			var spore := _spores[spore_id]
			var spore_mass := _rad_to_mass(spore.radius)
			var new_actor_mass := actor_mass + spore_mass
			_set_actor_mass(actor, new_actor_mass)
			_remove_spore(spore)

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


func _add_actor(
		actor_id: int,
		actor_name: String,
		x: float,
		y: float,
		radius: float,
		direction: float,
		speed: float,
		is_player: bool
	) -> void:
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
	_hiscores.set_hiscore(actor_name, _rad_to_mass(radius))
	_players[actor_id] = actor
	if is_player:
		actor.area_entered.connect(_on_player_area_entered)

	
func _update_actor(
	actor_id: int,
	x: float,
	y: float,
	radius: float,
	direction: float,
	speed: float,
	is_player: bool 
	) -> void:
	var actor := _players[actor_id]
	actor.radius = radius
	_hiscores.set_hiscore(actor.actor_name, _rad_to_mass(radius))
	if actor.position.distance_squared_to(Vector2(x, y)) > 100:
		actor.position.x = x
		actor.position.y = y
	
	if not is_player:
		actor.velocity = speed * Vector2.from_angle(direction)

func _on_player_area_entered(area: Area2D) -> void:
	if area is Spore: 
		_consume_spore(area as Spore)
	elif area is Actor:
		_collide_actor(area as Actor)
		
func _collide_actor(a: Actor) -> void: 
	var player := _players[GameManager.client_id]
	var p_mass := _rad_to_mass(player.radius)
	var a_mass := _rad_to_mass(a.radius)
	if p_mass > a_mass * 1.5:
		_set_actor_mass(player, p_mass + a_mass)
		var packet := packets.Packet.new()
		var player_consume_msg := packet.new_player_consumed()
		player_consume_msg.set_player_id(a.actor_id)
		WS.send(packet)
		_remove_actor(a)
		

func _consume_spore(spore: Spore) -> void:
	var player  := _players[GameManager.client_id]
	var p_mass  := _rad_to_mass(player.radius)
	var s_mass  := _rad_to_mass(spore.radius) 
	_set_actor_mass(player, p_mass + s_mass)
	
	var packet := packets.Packet.new()
	var spore_consume_message := packet.new_spore_consumed()
	spore_consume_message.set_id(spore.spore_id)
	WS.send(packet)
	_remove_spore(spore)
	
func _remove_spore(s: Spore) -> void:
	_spores.erase(s.spore_id)
	s.queue_free()

func _remove_actor(a: Actor) -> void: 
	_players.erase(a.actor_id)
	_hiscores.remove_hiscore(a.actor_name)
	a.queue_free()

func _rad_to_mass(r: float) -> float:
	return r * r * PI

func _set_actor_mass(a: Actor, m: float) -> void:
	a.radius = sqrt(m / PI)
	_hiscores.set_hiscore(a.actor_name, roundi(m))
