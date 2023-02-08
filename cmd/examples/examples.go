package examples

import (
	"github.com/cocoonspace/fsm"
)

const (
	StateFoo fsm.State = iota
	StateBar
)

const (
	EventFoo fsm.Event = iota
	EventBar
)

func fsm1() {
	f := fsm.New(StateFoo)
	f.Transition(
		fsm.On(EventFoo), fsm.Src(StateFoo),
		fsm.Dst(StateBar),
	)
	f.Transition(
		fsm.On(EventBar), fsm.Src(StateBar),
		fsm.Dst(StateFoo),
	)
}

var (
	f = fsm.New(StateBar)
)

func fsm2() {
	f.Transition(
		fsm.On(EventBar), fsm.Times(2), fsm.Src(StateBar),
		fsm.Dst(StateFoo),
	)
}

type sf struct {
	fsm *fsm.FSM
}

var (
	fs = sf{
		fsm: fsm.New(StateFoo),
	}
)

func fsm3() {
	fs.fsm.Transition(
		fsm.On(EventBar), fsm.Src(StateBar),
		fsm.Dst(StateFoo),
	)
}