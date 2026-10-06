extends Area2D

const packets := preload("res://packets.gd")
const Scene := preload("res://objects/actor/actor.tscn")
const Actor := preload("res://objects/actor/actor.gd")

var _target_zoom := 2.0
var _max_zoom := _target_zoom

@onready var _collision_shape: CollisionShape2D = $CollisionShape2D
@onready var _nameplate: Label = $Label
@onready var _camera: Camera2D = $Camera2D

var actor_id: int
var actor_name: String
var start_x: float
var start_y: float
var start_radius: float
var direction: float
var speed: float
var is_player: bool
var color: Color
var velocity: Vector2
var radius: float:
	set(r):
		radius = r
		_collision_shape.shape.radius = radius
		_update_zoom()
		queue_redraw()

static func instanciate(
		actor_id: int,
		actor_name: String,
		x: float,
		y: float,
		radius: float,
		direction: float,
		speed: float,
		is_player: bool,
		color: Color
	) -> Actor:
	var actor := Scene.instantiate()
	actor.actor_id = actor_id
	actor.actor_name = actor_name
	actor.start_x = x
	actor.start_y = y
	actor.start_radius = radius
	actor.direction = direction
	actor.speed = speed
	actor.is_player = is_player
	actor.color = color
	
	return actor

func _ready() -> void:
	position.x = start_x
	position.y = start_y
	velocity = Vector2.RIGHT * speed
	radius = start_radius
	
	_nameplate.text = actor_name
	_collision_shape.shape.radius = radius

func _process(_delta: float) -> void:
	if not is_equal_approx(_camera.zoom.x, _target_zoom):
		_camera.zoom -= Vector2(1,1) * (_camera.zoom.x - _target_zoom) * 0.05

func _physics_process(delta: float) -> void:
	position += velocity * delta
	if not is_player:
		return
	var mouse_pos := get_global_mouse_position()
	var input_vec = position.direction_to(mouse_pos).normalized()
	if abs(velocity.angle_to(input_vec)) > TAU / 15:
		velocity = input_vec * speed
		var packet := packets.Packet.new()
		var player_dir_message := packet.new_player_direction()
		player_dir_message.set_direction(velocity.angle())
		WS.send(packet)
	
func _draw() -> void:
	draw_circle(Vector2.ZERO, _collision_shape.shape.radius, color)
	
func _input(event: InputEvent) -> void:
	if is_player and event is InputEventMouseButton and event.is_pressed():
		match event.button_index:
			MOUSE_BUTTON_WHEEL_UP:
				_target_zoom = min(4, _target_zoom + 0.1)
			MOUSE_BUTTON_WHEEL_DOWN:
				_target_zoom = max(_max_zoom, _target_zoom - 0.1)
		_camera.zoom.y = _camera.zoom.x
		
func _update_zoom() -> void:
	if is_node_ready():
		_nameplate.add_theme_font_size_override("font_size", max(16, radius / 2))
	if not is_player:
		return
		
	var newMax := 2 * start_radius / radius
	if is_equal_approx(_target_zoom, _max_zoom):
		_target_zoom = newMax
	_max_zoom = newMax
	
	
	
	
	
	
