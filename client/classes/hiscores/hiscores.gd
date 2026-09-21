class_name Hiscores
extends ScrollContainer

var _scores: Array[int]

@onready var _entry_container: VBoxContainer = $VBoxContainer
@onready var _entry_template: HBoxContainer = $VBoxContainer/HBoxContainer

@warning_ignore("shadowed_variable_base_class")
func _add_hiscore(name: String, score: int, highlight: bool) -> void:
	_scores.append(score)
	_scores.sort()
	
	var pos := len(_scores) - _scores.find(score) - 1
	var entry: HBoxContainer = _entry_template.duplicate()
	var name_label: Label = entry.get_child(0)
	var score_label: Label =  entry.get_child(1)
	
	_entry_container.add_child(entry)
	_entry_container.move_child(entry, pos)
	name_label.text = name
	score_label.text = str(score)
	
	if highlight:
		name_label.add_theme_color_override("font_color", Color.YELLOW)
	
	entry.show()
	
@warning_ignore("shadowed_variable_base_class")
func set_hiscore(name: String, score: int, highlight: bool = false) -> void:
	remove_hiscore(name)
	_add_hiscore(name, score, highlight)
	
@warning_ignore("shadowed_variable_base_class")
func remove_hiscore(name: String) -> void:
	for i in range(len(_scores)):
		var entry: HBoxContainer = _entry_container.get_child(i)
		var pos := len(_scores) - i - 1
		var name_label: Label = entry.get_child(0)
		if name_label.text == name:
			_scores.remove_at(pos)
			entry.free()
			return
			
			
			 
func clear_scores() -> void:
	_scores.clear()
	for e in _entry_container.get_children():
		if e != _entry_template:
			e.free()
	
	
	
