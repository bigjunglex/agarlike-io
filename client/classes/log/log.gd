class_name Log
extends RichTextLabel

func _message(msg: String, color: Color = Color.WHITE) -> void:
	append_text("[color=#%s]%s\n" % [color.to_html(false), msg])

func info(msg: String) -> void: 
	_message(msg, Color.WHITE)

func warning(msg: String) -> void: 
	_message(msg, Color.ORANGE)

func error(msg: String) -> void: 
	_message(msg, Color.DARK_RED)

func success(msg: String) -> void: 
	_message(msg, Color.SPRING_GREEN)
	
func chat(sender_name: String, msg: String) -> void: 
	_message("[color=#%s]%s:[/color] [i]%s[/i]" % [Color.CORNFLOWER_BLUE.to_html(false), sender_name, msg])
