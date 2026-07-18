extends Node

const packets := preload("res://packets.gd")

var _action_on_ok_recieved: Callable 

@onready var _username_field: LineEdit = $UI/VBoxContainer/Username
@onready var _password_field: LineEdit = $UI/VBoxContainer/Password
@onready var _login_button: Button = $UI/VBoxContainer/HBoxContainer/LoginButton
@onready var _register_button: Button = $UI/VBoxContainer/HBoxContainer/RegisterButton
@onready var _log: Log = $UI/VBoxContainer/Log

func _ready() -> void:
	WS.packet_received.connect(_on_ws_packet_recieved)
	WS.connection_closed.connect(_on_ws_connection_closed)
	_login_button.pressed.connect(_on_login_pressed)
	_register_button.pressed.connect(_on_register_pressed)
	
func _on_ws_packet_recieved(packet: packets.Packet) -> void:
	var _sender_id := packet.get_sender_id()
	if packet.has_deny_response():
		var deny_response := packet.get_deny_response()
		_log.error(deny_response.get_reason())
	elif packet.has_ok_response():
		_action_on_ok_recieved.call()

func _on_register_pressed() -> void:
	var packet := packets.Packet.new()
	var register_req := packet.new_register_request()
	register_req.set_username(_username_field.text)
	register_req.set_password(_password_field.text)
	WS.send(packet)
	_action_on_ok_recieved = func(): _log.success("Registration succeed!")

func _on_login_pressed() -> void:
	var packet := packets.Packet.new()
	var login_req := packet.new_login_request()
	login_req.set_username(_username_field.text)
	login_req.set_password(_password_field.text)
	WS.send(packet)
	_action_on_ok_recieved = func(): GameManager.set_state(GameManager.State.INGAME)

func _on_ws_connection_closed() -> void:
	_log.warning("Disconnected from the server")
