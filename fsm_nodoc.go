//go:build !doc
// +build !doc

package fsm

import "errors"

type transition struct {
	conditions []optionCondition
	actions    []optionAction
}

// Src defines the source States for a Transition.
func Src(s ...NamedState) Option {
	return srcInternal(s...)
}

// On defines the Event that triggers a Transition.
func On(e NamedEvent) Option {
	return onInternal(e)
}

// Dst defines the new State the machine switches to after a Transition.
func Dst(s NamedState) Option {
	return dstInternal(s)
}

// NotCheck is an external condition that allows a Transition only if fn returns false.
func NotCheck(fn func() bool) Option {
	return notCheckInternal(fn)
}

// Check is an external condition that allows a Transition only if fn returns true.
func Check(fn func() bool) Option {
	return checkInternal(fn)
}

// Call defines a function that is called when a Transition occurs.
func Call(fn func()) Option {
	return callInternal(fn)
}

// Times defines the number of consecutive times conditions must be valid before a Transition occurs.
// Times will not work if multiple Transitions are possible at the same time.
func Times(n int) Option {
	return timesInternal(n)
}

// EnterState sets a func that will be called when entering a specific state.
func (f *FSM) EnterState(state NamedState, fn func()) {
	f.enterStateInternal(state, fn)
}

// ExitState sets a func that will be called when exiting a specific state.
func (f *FSM) ExitState(state NamedState, fn func()) {
	f.exitStateInternal(state, fn)
}

// GenerateDoc will find if it exist the mermaid block in the markdown file with the right title
// and update it with the content describing this FSM. If it can not find the mermaid block, it will
// append it at the end of the file. The FSM package need to be compile with the tag `doc` for this
// to work.
func (f *FSM) GenerateDoc(_ string, _ string) error {
	return errors.New("not compiled with doc tag")
}
