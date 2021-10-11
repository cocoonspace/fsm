package fsm_test

import (
	"fmt"
	"testing"

	"github.com/cocoonspace/fsm"
)

const (
	StateFoo fsm.State = iota
	StateBar
)

const (
	EventFoo fsm.Event = iota
)

func TestFSM(t *testing.T) {
	f := fsm.New(StateFoo)
	f.Transition(
		fsm.On(EventFoo), fsm.Src(StateFoo),
		fsm.Dst(StateBar),
	)
	res := f.Event(EventFoo)
	if !res {
		t.Error("Event returned false")
	}
	if f.Current() != StateBar {
		t.Error("Bad destination state")
	}
}

func ExampleCheck() {
	f := fsm.New(StateFoo)
	f.Transition(
		fsm.On(EventFoo), fsm.Src(StateFoo), fsm.Check(func() bool {
			return true
		}),
		fsm.Dst(StateBar),
	)
}

func ExampleCall() {
	f := fsm.New(StateFoo)
	f.Transition(
		fsm.On(EventFoo), fsm.Src(StateFoo),
		fsm.Dst(StateBar), fsm.Call(func() {
			fmt.Println("Call called")
		}),
	)
}
