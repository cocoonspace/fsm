//go:build doc
// +build doc

package fsm

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
)

type transition struct {
	conditions []optionCondition
	actions    []optionAction

	srcs  []string
	on    string
	dst   string
	calls []string
	times int
}

// Src defines the source States for a Transition.
func Src(s ...ExtendedState) Option {
	return func(t *transition) {
		srcInternal(s...)(t)

		for _, src := range s {
			t.srcs = append(t.srcs, fmt.Sprintf("%v", src))
		}
	}
}

// On defines the Event that triggers a Transition.
func On(e ExtendedEvent) Option {
	return func(t *transition) {
		onInternal(e)(t)

		t.on = fmt.Sprintf("%v", e)
	}
}

// Dst defines the new State the machine switches to after a Transition.
func Dst(s ExtendedState) Option {
	return func(t *transition) {
		dstInternal(s)(t)

		t.dst = fmt.Sprintf("%v", s)
	}
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
	_, file, line, _ := runtime.Caller(1)

	return func(t *transition) {
		callInternal(fn)(t)

		t.calls = append(t.calls, fmt.Sprintf("%s:%d", file, line))
	}
}

// Times defines the number of consecutive times conditions must be valid before a Transition occurs.
// Times will not work if multiple Transitions are possible at the same time.
func Times(n int) Option {
	return func(t *transition) {
		timesInternal(n)

		t.times = n
	}
}

// EnterState sets a func that will be called when entering a specific state.
func (f *FSM) EnterState(state ExtendedState, fn func()) {
	f.enterStateInternal(state, fn)
}

// ExitState sets a func that will be called when exiting a specific state.
func (f *FSM) ExitState(state ExtendedState, fn func()) {
	f.exitStateInternal(state, fn)
}

// GenerateDoc will find if it exist the mermaid block in the markdown file with the right title
// and update it with the content describing this FSM. If it can not find the mermaid block, it will
// append it at the end of the file. The FSM package need to be compile with the tag `doc` for this
// to work.
func (f *FSM) GenerateDoc(title string, file string) error {
	lookupTitle := []byte("_" + title + "_:")

	generated, err := f.insertMermaidGraphInPlace(lookupTitle, file)
	if err == nil {
		os.Remove(file)
		return os.Rename(generated, file)
	}

	if !os.IsNotExist(err) {
		return err
	}

	c, err := os.Create(file)
	if err != nil {
		return err
	}
	defer c.Close()

	w := bufio.NewWriter(c)
	defer w.Flush()

	f.insertMermaidBlock(lookupTitle, w)

	return nil
}

var uniqueNameCounter int

func uniqueName(w *bufio.Writer, state string, uniqueNameMapping map[string]string) string {
	unique, ok := uniqueNameMapping[state]
	if ok {
		return unique
	}

	// Generate a unique ID for this state
	unique = fmt.Sprintf("id%d", uniqueNameCounter)
	uniqueNameCounter++
	w.WriteString("\t" + unique + "(" + state + ")\n")
	uniqueNameMapping[state] = unique

	return unique
}

var (
	lookupMermaid = []byte("```mermaid")
	lookupEnd     = []byte("```")
)

func (f *FSM) insertMermaidGraph(w *bufio.Writer) {
	uniqueNameMapping := make(map[string]string)

	w.WriteString("flowchart LR\n")

	for _, t := range f.transitions {
		dstID := uniqueName(w, t.dst, uniqueNameMapping)

		if t.on == "" {
			continue
		}
		on := t.on
		if t.times > 1 {
			on = fmt.Sprintf("%d x %s", t.times, on)
		}
		for _, call := range t.calls {
			prettyCall := call
			for _, path := range documentationPathToIgnore {
				if strings.HasPrefix(prettyCall, path) {
					prettyCall = prettyCall[len(path):]
					break
				}
			}
			on = on + "<br>" + prettyCall
		}

		for _, src := range t.srcs {
			srcID := uniqueName(w, src, uniqueNameMapping)

			w.WriteString("\t" + srcID + "--> |" + on + "| " + dstID + "\n")
		}
	}
}

func (f *FSM) insertMermaidBlock(lookupTitle []byte, w *bufio.Writer) {
	w.Write(lookupTitle)
	w.WriteString("\n")
	w.Write(lookupMermaid)
	w.WriteString("\n")

	f.insertMermaidGraph(w)

	w.Write(lookupEnd)
	w.WriteString("\n")
}

func (f *FSM) insertMermaidGraphInPlace(lookupTitle []byte, file string) (string, error) {
	out, err := ioutil.TempFile(".", "fsm-doc-")
	if err != nil {
		return "", err
	}
	defer out.Close()

	w := bufio.NewWriter(out)
	defer w.Flush()

	in, err := os.Open(file)
	if err != nil {
		os.Remove(out.Name())
		return "", err
	}
	defer in.Close()

	mermaidNext := false
	searchEnd := false
	found := false

	reader := bufio.NewReader(in)
	for {
		line, _, err := reader.ReadLine()

		if err == io.EOF {
			break
		} else if err != nil {
			return "", err
		}

		if bytes.Equal(line, lookupTitle) {
			mermaidNext = true
		} else if mermaidNext && bytes.Equal(line, lookupMermaid) {
			mermaidNext = false
			searchEnd = true
			found = true

			w.Write(lookupMermaid)
			w.WriteString("\n")
		} else if searchEnd && bytes.Equal(line, lookupEnd) {
			f.insertMermaidGraph(w)
			searchEnd = false
		} else {
			mermaidNext = false
		}

		if !searchEnd {
			w.Write(line)
			w.WriteString("\n")
		}
	}

	if !found {
		f.insertMermaidBlock(lookupTitle, w)
	}

	return out.Name(), nil
}
