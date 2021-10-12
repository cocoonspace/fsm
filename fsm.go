package fsm

// Event is the event type.
// You can define your own values as
//	const (
//		EventFoo fsm.Event = iota
//		EventBar
//	)
type Event int

// State is the state type.
// You can define your own values as
//	const (
//		StateFoo fsm.State = iota
//		StateBar
//	)
type State int

type transition struct {
	conditions []optionCondition
	actions    []optionAction
}

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
func New(initial State) *FSM {
	return &FSM{
		enterState: map[State]func(){},
		exitState:  map[State]func(){},
		current:    initial,
		initial:    initial,
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

// Src defines the source States for a Transition.
func Src(s ...State) Option {
	return func(t *transition) {
		t.conditions = append(t.conditions, func(e Event, times int, fsm *FSM) result {
			for _, src := range s {
				if fsm.current == src {
					return resultOK
				}
			}
			return resultNOK
		})
	}
}

// On defines the Event that triggers a Transition.
func On(e Event) Option {
	return func(t *transition) {
		t.conditions = append(t.conditions, func(evt Event, times int, fsm *FSM) result {
			if e == evt {
				return resultOK
			}
			return resultNOK
		})
	}
}

// Dst defines the new State the machine switches to after a Transition.
func Dst(s State) Option {
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
			fsm.current = s
			if fn, ok := fsm.enterState[fsm.current]; ok {
				fn()
			}
			if fsm.enter != nil {
				fsm.enter(fsm.current)
			}
		})
	}
}

// NotCheck is an external condition that allows a Transition only if fn returns false.
func NotCheck(fn func() bool) Option {
	return func(t *transition) {
		t.conditions = append(t.conditions, func(e Event, times int, fsm *FSM) result {
			if !fn() {
				return resultOK
			}
			return resultNOK
		})
	}
}

// Check is an external condition that allows a Transition only if fn returns true.
func Check(fn func() bool) Option {
	return func(t *transition) {
		t.conditions = append(t.conditions, func(e Event, times int, fsm *FSM) result {
			if fn() {
				return resultOK
			}
			return resultNOK
		})
	}
}

// Call defines a function that is called when a Transition occurs.
func Call(fn func()) Option {
	return func(t *transition) {
		t.actions = append(t.actions, func(fsm *FSM) {
			fn()
		})
	}
}

// Times defines the number of consecutive times conditions must be valid before a Transition occurs.
// Times will not work if multiple Transitions are possible at the same time.
func Times(n int) Option {
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

// EnterState sets a func that will be called when entering a specific state.
func (f *FSM) EnterState(state State, fn func()) {
	f.enterState[state] = fn
}

// ExitState sets a func that will be called when exiting a specific state.
func (f *FSM) ExitState(state State, fn func()) {
	f.exitState[state] = fn
}

// Event send an Event to a machine, applying at most one transition.
// true is returned if a transition has been applied, false otherwise.
func (f *FSM) Event(e Event) bool {
	for i := range f.transitions {
		times := f.times
		if i != f.previous {
			times = 0
		}
		if res := f.transitions[i].match(e, times+1, f); res != resultNOK {
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
