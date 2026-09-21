extends Area2D

const packets := preload("res://packets.gd")
const Scene := preload("res://objects/spore/spore.tscn")
const Spore := preload("res://objects/spore/spore.gd")
const Actor := preload("res://objects/actor/actor.gd")

var spore_id: int
var x: float
var y: float
var radius: float
var color: Color

var is_dropped: bool

@onready var _collision_shape: CircleShape2D = $CollisionShape2D.shape


static func instanciate(
	spore_id: int,
	x: float,
	y: float,
	radius: float,
	is_dropped: bool
	#color: Color
) -> Spore:
	var spore := Scene.instantiate()
	spore.spore_id = spore_id
	spore.x = x
	spore.y = y
	spore.radius = radius
	spore.is_dropped = is_dropped
	#spore.color = color
	return spore
	
func _ready() -> void:
	position.x = x
	position.y = y
	_collision_shape.radius = radius
	color = Color.from_hsv(randf(), 1, 1, 1)
	if is_dropped:
		area_exited.connect(_on_area_exit)
	
func _on_area_exit(area: Area2D) -> void:
	if area is Actor:
		is_dropped = false

func _draw() -> void:
	draw_circle(Vector2.ZERO, radius, color, true)
