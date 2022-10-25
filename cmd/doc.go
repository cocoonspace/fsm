//go:generate stringer -type=State,Event --output=doc_fsm_string.go
//go:generate go run -tags doc doc.go doc_fsm_string.go ../README.md

package main

import (
	"fmt"
	"os"

	"github.com/cocoonspace/fsm"
)

type State fsm.State
type Event fsm.Event

func (s State) State() fsm.State {
	return fsm.State(s)
}

func (e Event) Event() fsm.Event {
	return fsm.Event(e)
}

var _ fsm.ExtendedState = (*State)(nil)
var _ fsm.ExtendedEvent = (*Event)(nil)

const (
	StateFoo State = iota
	StateBar
)

const (
	EventFoo Event = iota
)

func example() *fsm.FSM {
	f := fsm.New(StateFoo.State())
	f.Transition(
		fsm.On(EventFoo),
		fsm.Src(StateFoo),
		fsm.Dst(StateBar),
	)

	f.Transition(
		fsm.On(EventFoo),
		fsm.Times(2),
		fsm.Src(StateFoo),
		fsm.Dst(StateBar),
	)

	return f
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("use go generate")
		return
	}
	f := example()
	f.GenerateDoc("Visual example", os.Args[1])
}
