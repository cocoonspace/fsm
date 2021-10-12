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

func TestCheck(t *testing.T) {
	check := false
	f := fsm.New(StateFoo)
	f.Transition(
		fsm.On(EventFoo), fsm.Src(StateFoo), fsm.Check(func() bool {
			return check
		}),
		fsm.Dst(StateBar),
	)
	res := f.Event(EventFoo)
	if res || f.Current() == StateBar {
		t.Error("Transition should not happen because of Check")
	}
	check = true
	res = f.Event(EventFoo)
	if !res && f.Current() != StateBar {
		t.Error("Transition should happen because of Check")
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

func TestCall(t *testing.T) {
	call := false
	f := fsm.New(StateFoo)
	f.Transition(
		fsm.On(EventFoo), fsm.Src(StateFoo),
		fsm.Call(func() {
			call = true
		}),
	)
	_ = f.Event(EventFoo)
	if !call {
		t.Error("Call should have been called")
	}
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

func TestTimes(t *testing.T) {
	f := fsm.New(StateFoo)
	f.Transition(
		fsm.On(EventFoo), fsm.Src(StateFoo), fsm.Times(2),
		fsm.Dst(StateBar),
	)

	res := f.Event(EventFoo)
	if res || f.Current() == StateBar {
		t.Error("Transition should not happen the first time")
	}
	res = f.Event(EventFoo) // transition to StateBar
	if !res || f.Current() != StateBar {
		t.Error("Transition should happen the second time")
	}
}

func ExampleTimes() {
	f := fsm.New(StateFoo)
	f.Transition(
		fsm.On(EventFoo), fsm.Src(StateFoo), fsm.Times(2),
		fsm.Dst(StateBar),
	)

	_ = f.Event(EventFoo) // no transition
	_ = f.Event(EventFoo) // transition to StateBar
}
