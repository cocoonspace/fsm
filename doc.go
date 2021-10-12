/*
Package fsm allows you to add Finite State Machines to your code.

	const (
		StateFoo fsm.State = iota
		StateBar
	)

	const (
		EventFoo fsm.Event = iota
	)

	f := fsm.New(StateFoo)
	f.Transition(
		fsm.On(EventFoo), fsm.Src(StateFoo),
		fsm.Dst(StateBar),
	)

You can have custom checks or actions:

	f.Transition(
		fsm.Src(StateFoo), fsm.Check(func() bool {
			// check something
		}),
		fsm.Call(func() {
			// do something
		}),
	)


Transitions can be triggered the second time an event occurs:

	f.Transition(
		fsm.On(EventFoo), fsm.Src(StateFoo), fsm.Times(2),
		fsm.Dst(StateBar),
	)

Functions can be called when entering or leaving a state:

	f.EnterState(StateFoo, func() {
		// do something
	})
	f.Enter(func(state fsm.State) {
		// do something
	})
	f.ExitState(StateFoo, func() {
		// do something
	})
	f.Exit(func(state fsm.State) {
		// do something
	})

*/
package fsm
