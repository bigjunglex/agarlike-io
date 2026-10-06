extends Node

const packets := preload("res://packets.gd")

var _action_on_ok_recieved: Callable 
@onready var _log: Log = $UI/MarginContainer/VBoxContainer/Log
@onready var _hiscores: Hiscores = $UI/MarginContainer/VBoxContainer/Hiscores
@onready var _login_form: VBoxContainer = $UI/MarginContainer/VBoxContainer/LoginForm
@onready var _register_form: RegisterForm = $UI/MarginContainer/VBoxContainer/RegisterForm
@onready var _register_prompt: RichTextLabel = $UI/MarginContainer/VBoxContainer/RegisterPrompt


func _ready() -> void:
	WS.packet_received.connect(_on_ws_packet_recieved)
	WS.connection_closed.connect(_on_ws_connection_closed)
	_login_form.form_submitted.connect(_on_login_pressed)
	_register_form.form_cancelled.connect(_on_reg_cancel)
	_register_form.form_submitted.connect(_on_register_pressed)
	_register_prompt.meta_clicked.connect(_on_register_meta)
	
func _on_ws_packet_recieved(packet: packets.Packet) -> void:
	var _sender_id := packet.get_sender_id()
	if packet.has_deny_response():
		var deny_response := packet.get_deny_response()
		_log.error(deny_response.get_reason())
	elif packet.has_ok_response():
		_action_on_ok_recieved.call()
	elif packet.has_hiscore_board():
		_handle_hiscore_board(packet.get_hiscore_board())

func _on_register_pressed(username: String, password: String, confirm_password: String, color: Color) -> void:
	if password != confirm_password:
		_log.error("Passwords do not match")
		return
		
	var packet := packets.Packet.new()
	var register_req := packet.new_register_request()
	register_req.set_username(username)
	register_req.set_password(password)
	register_req.set_color(color.to_rgba32())
	WS.send(packet)
	_action_on_ok_recieved = func(): _log.success("Registration succeed!")

func _on_reg_cancel() -> void:
	_register_form.hide()
	_login_form.show()
	_register_prompt.show()
	_hiscores.show()

func _on_login_pressed(username: String, password: String) -> void:
	var packet := packets.Packet.new()
	var login_req := packet.new_login_request()
	login_req.set_username(username)
	login_req.set_password(password)
	WS.send(packet)
	_action_on_ok_recieved = func(): GameManager.set_state(GameManager.State.INGAME)

func _on_hiscores_pressed() -> void:
	GameManager.set_state(GameManager.State.BROWSE_SCORES)

func _on_register_meta(meta) -> void:
	if meta is String and meta == "register":
		_login_form.hide()
		_register_form.show()
		_register_prompt.hide()
		_hiscores.hide()
		

func _on_ws_connection_closed() -> void:
	_log.warning("Disconnected from the server")
	

func _handle_hiscore_board(msg: packets.HiscoreBoardMessage) -> void:
	for m in msg.get_hiscores():
		var name := "%d. %s" % [m.get_rank(), m.get_name()]
		var score := m.get_score()
		_hiscores.set_hiscore(name, score)
