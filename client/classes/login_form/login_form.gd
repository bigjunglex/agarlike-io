extends VBoxContainer

@onready var _username: LineEdit = $Username
@onready var _password: LineEdit = $Password
@onready var _login_btb: Button = $HBoxContainer/LoginButton
@onready var _hiscores: Button = $HBoxContainer/Hiscores

signal form_submitted(username: String, password: String) 

func _ready() -> void:
	_login_btb.pressed.connect(_on_login)
	_hiscores.pressed.connect(_on_hiscores)

func _on_login() -> void: 
	form_submitted.emit(_username.text, _password.text)


func _on_hiscores() -> void:
	GameManager.set_state(GameManager.State.BROWSE_SCORES)
