package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"strconv"
)

type transition struct {
	on    string
	src   []string
	dst   string
	times int
}

type flowchart struct {
	initial     string
	transitions []transition
}

func (fc *flowchart) render(w io.Writer) error {
	_, err := io.WriteString(w, "flowchart LR\n")
	if err != nil {
		return err
	}
	nodes := map[string]string{}
	if fc.initial != "" {
		_, err = io.WriteString(w, "id0[Start]\n")
		if err != nil {
			return err
		}
		nodes["Start"] = "id0"
		fc.transitions = append(fc.transitions, transition{src: []string{"Start"}, dst: fc.initial})
	}
	fn := func(ns ...string) error {
		for _, n := range ns {
			if nodes[n] != "" {
				return nil
			}
			id := "id" + strconv.Itoa(len(nodes))
			nodes[n] = id
			_, err := io.WriteString(w, id+"("+n+")\n")
			if err != nil {
				return err
			}
		}
		return nil
	}
	for _, t := range fc.transitions {
		err := fn(t.src...)
		if err != nil {
			return err
		}
		err = fn(t.dst)
		if err != nil {
			return err
		}
	}
	for _, t := range fc.transitions {
		on := t.on
		if t.times > 1 {
			on += " x" + strconv.Itoa(t.times)
		}
		if on != "" {
			on = "|" + on + "|"
		}
		for _, src := range t.src {
			_, err := io.WriteString(w, nodes[src]+"-->"+on+nodes[t.dst]+"\n")
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func main() {

	flag.Parse()
	file := flag.Arg(0)
	if file == "" {
		fmt.Println("Usage: doc [filename]")
		os.Exit(1)
	}
	fset := token.NewFileSet()
	parserMode := parser.ParseComments
	var fileAst *ast.File
	var err error

	fileAst, err = parser.ParseFile(fset, file, nil, parserMode)
	if err != nil {
		panic(err)
	}
	flowcharts := map[*ast.Object]*flowchart{}
	for _, d := range fileAst.Decls {
		switch decl := d.(type) {
		case *ast.FuncDecl:
			for _, stmt := range decl.Body.List {
				switch stmt := stmt.(type) {
				case *ast.AssignStmt:
					for i := range stmt.Lhs {
						fc := parseInit(stmt.Rhs[i])
						if fc == nil {
							continue
						}
						if o := obj(stmt.Lhs[i]); o != nil {
							flowcharts[o] = fc
						}
					}
				case *ast.ExprStmt:
					obj, t := parseTransition(stmt.X)
					if t == nil {
						continue
					}
					if obj == nil || flowcharts[obj] == nil {
						fmt.Fprintf(os.Stderr, "Error: Found transition but no matching FSM in %v\n", fset.Position(stmt.Pos()))
						continue
					}
					flowcharts[obj].transitions = append(flowcharts[obj].transitions, *t)
				}
			}
		case *ast.GenDecl:
			for _, spec := range decl.Specs {
				switch spec := spec.(type) {
				case *ast.ValueSpec:
					for i, id := range spec.Names {
						if len(spec.Values) > i {
							fc := parseInit(spec.Values[i])
							if fc != nil {
								flowcharts[id.Obj] = fc
							}
						}
					}
				}
			}
		}
	}
	for _, fc := range flowcharts {
		fc.render(os.Stdout)
	}
}

func parseInit(e ast.Expr) *flowchart {
	var call *ast.CallExpr
	switch v := e.(type) {
	case *ast.CallExpr:
		call = v
	case *ast.KeyValueExpr:
		return parseInit(v.Value)
	case *ast.CompositeLit:
		for _, e := range v.Elts {
			i := parseInit(e)
			if i != nil {
				return i
			}
		}
	}
	if call == nil || len(call.Args) != 1 {
		return nil
	}
	fc := flowchart{}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "New" {
		return nil
	}
	pkg, ok := sel.X.(*ast.Ident)
	if !ok || pkg.Name != "fsm" {
		return nil
	}
	if ident, ok := call.Args[0].(*ast.Ident); ok {
		fc.initial = ident.Name
	}
	return &fc
}

func parseTransition(e ast.Expr) (*ast.Object, *transition) {
	call, ok := e.(*ast.CallExpr)
	if !ok {
		return nil, nil
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "Transition" {
		return nil, nil
	}
	t := transition{}
	for _, arg := range call.Args {
		if call, ok := arg.(*ast.CallExpr); ok {
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				switch sel.Sel.Name {
				case "On":
					t.on = call.Args[0].(*ast.Ident).Name
				case "Src":
					for _, arg := range call.Args {
						if ident, ok := arg.(*ast.Ident); ok {
							t.src = append(t.src, ident.Name)
						}
					}
				case "Dst":
					t.dst = call.Args[0].(*ast.Ident).Name
				case "Times":
					if lit, ok := call.Args[0].(*ast.BasicLit); ok {
						t.times, _ = strconv.Atoi(lit.Value)
					}
				}
			}
		}
	}
	return obj(sel.X), &t
}

func obj(s ast.Expr) *ast.Object {
	switch v := s.(type) {
	case *ast.Ident:
		return v.Obj
	case *ast.SelectorExpr:
		if v.Sel.Obj != nil {
			return v.Sel.Obj
		}
		return obj(v.X)
	default:
		return nil
	}
}
