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

You can have custom checks or actions :

	f.Transition(
		fsm.Src(StateFoo), fsm.Check(func() bool {
			// check something
		}),
		fsm.Call(func() {
			// do something
		}),
	)

*/
package fsm
