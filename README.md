# fsm [![PkgGoDev](https://pkg.go.dev/badge/github.com/cocoonspace/fsm)](https://pkg.go.dev/github.com/cocoonspace/fsm) [![Build Status](https://app.travis-ci.com/cocoonspace/fsm.svg?branch=master)](https://app.travis-ci.com/cocoonspace/fsm) [![Coverage Status](https://coveralls.io/repos/github/cocoonspace/fsm/badge.svg?branch=master)](https://coveralls.io/github/cocoonspace/fsm?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/cocoonspace/fsm)](https://goreportcard.com/report/github.com/cocoonspace/fsm)

Package fsm allows you to add [finite-state machines](https://en.wikipedia.org/wiki/Finite-state_machine) to your Go code.

States and Events are defined as int consts:

```go
const (
    StateFoo fsm.State = iota
    StateBar
)

const (
    EventFoo fsm.Event = iota
    EventBar
)

f := fsm.New(StateFoo)
f.Transition(
    fsm.On(EventFoo), fsm.Src(StateFoo),
    fsm.Dst(StateBar),
)
```

You can have custom checks or actions:

```go
f.Transition(
    fsm.Src(StateFoo), fsm.Check(func () bool {
        // check something
    }),
    fsm.Call(func () {
        // do something
    }),
)
```

Transitions can be triggered the second time an event occurs:

```go
f.Transition(
    fsm.On(EventFoo), fsm.Src(StateFoo), fsm.Times(2),
    fsm.Dst(StateBar),
)
```

Functions can be called when entering or leaving a state:

```go
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
```

## Performance

This package is much faster and does a lot less allocations than github.com/looplab/fsm:

```
BenchmarkCocoonSpaceFSM-12    	29371851	        40.32 ns/op	       0 B/op	       0 allocs/op
BenchmarkLooplabFSM-12        	 2438946	       487.8 ns/op	     320 B/op	       4 allocs/op
```

(benchmark data is for two executed transitions)

Benchmark information on https://github.com/cocoonspace/fsm-bench

## Installation

go get github.com/cocoonspace/fsm

## Contribution guidelines

Contributions are welcome, as long as:

* unit tests & comments are included,
* no external package is used.

## License

MIT - See LICENSE
