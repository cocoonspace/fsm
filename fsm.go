package fsm

import (
	"fmt"
	"strconv"
)

// Event is the event type.
// You can define your own values as
//
//	const (
//		EventFoo fsm.Event = iota
//		EventBar
//	)
type Event int

// State is the state type.
// You can define your own values as
//
//	const (
//		StateFoo fsm.State = iota
//		StateBar
//	)
type State int

// NamedState allow for pretty printing the FSM state by providing a String() interface
type NamedState interface {
	State() State
	fmt.Stringer
}

// NamedEvent allow for pretty printing the FSM event by providing a String() interface
type NamedEvent interface {
	Event() Event
	fmt.Stringer
}

var _ NamedState = (*State)(nil)
var _ NamedEvent = (*Event)(nil)

func (t *transition) match(e Event, times int, fsm *FSM) result {
	var res result
	for _, fn := range t.conditions {
		cres := fn(e, times, fsm)
		if cres == resultNOK {
			return resultNOK
		}
		if cres > res {
			res = cres
		}
	}
	return res
}

func (t *transition) apply(fsm *FSM) {
	for _, fn := range t.actions {
		fn(fsm)
	}
}

// FSM is a finite state machine.
type FSM struct {
	transitions []transition
	enterState  map[State]func()
	exitState   map[State]func()
	enter       func(State)
	exit        func(State)
	current     State
	initial     State
	previous    int
	times       int
}

// New creates a new finite state machine having the specified initial state.
func New(initial NamedState) *FSM {
	return &FSM{
		enterState: map[State]func(){},
		exitState:  map[State]func(){},
		current:    initial.State(),
		initial:    initial.State(),
	}
}

// Option defines a transition option.
type Option func(*transition)

type result int

const (
	resultNOK result = iota
	resultOK
	resultNoAction
)

type optionCondition func(e Event, times int, fsm *FSM) result

type optionAction func(*FSM)

// Transition creates a new transition, usually having trigger On an Event, from a Src State, to a Dst State.
func (f *FSM) Transition(opts ...Option) {
	t := transition{}
	for _, opt := range opts {
		opt(&t)
	}
	f.transitions = append(f.transitions, t)
}

func srcInternal(s ...NamedState) Option {
	return func(t *transition) {
		t.conditions = append(t.conditions, func(e Event, times int, fsm *FSM) result {
			for _, src := range s {
				if fsm.current == src.State() {
					return resultOK
				}
			}
			return resultNOK
		})
	}
}

func onInternal(e NamedEvent) Option {
	return func(t *transition) {
		t.conditions = append(t.conditions, func(evt Event, times int, fsm *FSM) result {
			if e.Event() == evt {
				return resultOK
			}
			return resultNOK
		})
	}
}

func dstInternal(s NamedState) Option {
	return func(t *transition) {
		t.actions = append(t.actions, func(fsm *FSM) {
			if fsm.current == s {
				return
			}
			if fn, ok := fsm.exitState[fsm.current]; ok {
				fn()
			}
			if fsm.exit != nil {
				fsm.exit(fsm.current)
			}
			fsm.current = s.State()
			if fn, ok := fsm.enterState[fsm.current]; ok {
				fn()
			}
			if fsm.enter != nil {
				fsm.enter(fsm.current)
			}
		})
	}
}

func notCheckInternal(fn func() bool) Option {
	return func(t *transition) {
		t.conditions = append(t.conditions, func(e Event, times int, fsm *FSM) result {
			if !fn() {
				return resultOK
			}
			return resultNOK
		})
	}
}

func checkInternal(fn func() bool) Option {
	return func(t *transition) {
		t.conditions = append(t.conditions, func(e Event, times int, fsm *FSM) result {
			if fn() {
				return resultOK
			}
			return resultNOK
		})
	}
}

func callInternal(fn func()) Option {
	return func(t *transition) {
		t.actions = append(t.actions, func(fsm *FSM) {
			fn()
		})
	}
}

func timesInternal(n int) Option {
	return func(t *transition) {
		t.conditions = append(t.conditions, func(e Event, times int, fsm *FSM) result {
			if times == n {
				return resultOK
			}
			if times < n {
				return resultNoAction
			}
			return resultNOK
		})
	}
}

// Reset resets the machine to its initial state.
func (f *FSM) Reset() {
	f.current = f.initial
}

// Current returns the current state.
func (f *FSM) Current() State {
	return f.current
}

// Enter sets a func that will be called when entering any state.
func (f *FSM) Enter(fn func(state State)) {
	f.enter = fn
}

// Exit sets a func that will be called when exiting any state.
func (f *FSM) Exit(fn func(state State)) {
	f.exit = fn
}

func (f *FSM) enterStateInternal(state NamedState, fn func()) {
	f.enterState[state.State()] = fn
}

func (f *FSM) exitStateInternal(state NamedState, fn func()) {
	f.exitState[state.State()] = fn
}

// Event send an Event to a machine, applying at most one transition.
// true is returned if a transition has been applied, false otherwise.
func (f *FSM) Event(e NamedEvent) bool {
	for i := range f.transitions {
		times := f.times
		if i != f.previous {
			times = 0
		}
		if res := f.transitions[i].match(e.Event(), times+1, f); res != resultNOK {
			if res == resultOK {
				f.transitions[i].apply(f)
			}
			if i == f.previous {
				f.times++
			} else {
				f.previous = i
				f.times = 1
			}
			return res == resultOK
		}
	}
	return false
}

var documentationPathToIgnore []string

// AddDocumentationPathToIgnore add base path to ignore when displaying file path in generated documentation
func AddDocumentationPathToIgnore(path string) {
	documentationPathToIgnore = append(documentationPathToIgnore, path)
}

// State return the value of a State to be compliant with NamedState
func (f State) State() State {
	return f
}

// String return the value as a string of a State to be compliant with NamedState
func (f State) String() string {
	return strconv.Itoa(int(f))
}

// Event return the value of an Event to be compliant with NamedEvent
func (e Event) Event() Event {
	return e
}

// String return the value as a string of an Event to be compliant with NamedEvent
func (e Event) String() string {
	return strconv.Itoa(int(e))
}
