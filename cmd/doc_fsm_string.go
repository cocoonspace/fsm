// Code generated by "stringer -type=State,Event --output=doc_fsm_string.go"; DO NOT EDIT.

package main

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[StateFoo-0]
	_ = x[StateBar-1]
}

const _State_name = "StateFooStateBar"

var _State_index = [...]uint8{0, 8, 16}

func (i State) String() string {
	if i < 0 || i >= State(len(_State_index)-1) {
		return "State(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _State_name[_State_index[i]:_State_index[i+1]]
}
func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[EventFoo-0]
}

const _Event_name = "EventFoo"

var _Event_index = [...]uint8{0, 8}

func (i Event) String() string {
	if i < 0 || i >= Event(len(_Event_index)-1) {
		return "Event(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Event_name[_Event_index[i]:_Event_index[i+1]]
}