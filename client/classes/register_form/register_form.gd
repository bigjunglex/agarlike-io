class_name RegisterForm
extends VBoxContainer

@onready var _username: LineEdit = $Username
@onready var _password: LineEdit = $Password
@onready var _confirm_password: LineEdit = $ConfirmPassword
@onready var _register: Button = $HBoxContainer/Register
@onready var _cancel: Button = $HBoxContainer/Cancel
@onready var _color_picker: ColorPicker = $ColorPicker

signal form_submitted(
	username: String,
	password: String,
	confirm_password: String,
	color: Color
	)
signal form_cancelled()

func _ready() -> void:
	_register.pressed.connect(_on_confirm_button_pressed)
	_cancel.pressed.connect(_on_cancel_button_pressed)

func _on_confirm_button_pressed() -> void:
		form_submitted.emit(
			_username.text,
			_password.text,
			_confirm_password.text,
			_color_picker.color
			)

func _on_cancel_button_pressed() -> void:
	form_cancelled.emit()
