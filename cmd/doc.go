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
	"strings"
)

type transition struct {
	on     string
	src    []string
	dst    string
	times  int
	checks []string
	calls  []string
}

type flowchart struct {
	name        string
	source      string
	initial     string
	transitions []transition
}

func (fc *flowchart) render(w io.Writer) error {
	r := "\n```mermaid\n---\ntitle: " +
		fc.name + " " +
		fc.source +
		"\n---\nflowchart LR\n"
	nodes := map[string]string{}
	if fc.initial != "" {
		r += "id0[Start]\n"
		nodes["Start"] = "id0"
		fc.transitions = append(fc.transitions, transition{src: []string{"Start"}, dst: fc.initial})
	}
	newnode := func(id, label string, style byte) {
		if id == "" || nodes[id] != "" {
			return
		}
		nid := "id" + strconv.Itoa(len(nodes))
		nodes[id] = nid
		switch style {
		case '(':
			r += nid + "(" + label + ")\n"
		case '[':
			r += nid + "[[" + label + "]]\n"
		}
		return
	}
	callnode := func(t *transition) string {
		return strings.Join(t.calls, "|") + "||" + t.dst
	}
	for _, t := range fc.transitions {
		for _, src := range t.src {
			newnode(src, src, '(')
		}
		newnode(t.dst, t.dst, '(')
		if len(t.calls) > 0 {
			newnode(callnode(&t), strings.Join(t.calls, " ,"), '[')
		}
	}
	for _, t := range fc.transitions {
		on := t.on
		if len(t.checks) > 0 {
			if on != "" {
				on += " + "
			}
			on += strings.Join(t.checks, "? + ") + "?"
		}
		if t.times > 1 {
			on += " x" + strconv.Itoa(t.times)
		}
		if on != "" {
			on = "|" + on + "|"
		}
		tdst := t.dst
		if len(t.calls) > 0 {
			tdst = callnode(&t)
			r += nodes[tdst] + "-->" + nodes[t.dst] + "\n"
		}
		for _, src := range t.src {
			dst := tdst
			if dst == "" {
				dst = src
			}
			r += nodes[src] + "-->" + on + nodes[dst] + "\n"
		}
	}
	_, err := io.WriteString(w, r+"```\n")
	return err
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
						if i >= len(stmt.Rhs) {
							continue
						}
						fc := parseInit(stmt.Rhs[i],
							fileAst.Name.String()+"."+decl.Name.String()+"()",
							fset.Position(stmt.Pos()).String(),
						)
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
							fc := parseInit(spec.Values[i],
								fileAst.Name.String()+"."+id.String(),
								fset.Position(spec.Pos()).String(),
							)
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

func parseInit(e ast.Expr, name, source string) *flowchart {
	var call *ast.CallExpr
	switch v := e.(type) {
	case *ast.CallExpr:
		call = v
	case *ast.KeyValueExpr:
		return parseInit(v.Value, name, source)
	case *ast.CompositeLit:
		for _, e := range v.Elts {
			i := parseInit(e, name, source)
			if i != nil {
				return i
			}
		}
	}
	if call == nil || len(call.Args) != 1 {
		return nil
	}
	fc := flowchart{
		name:   name,
		source: source,
	}
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
				case "Check":
					chk := funcname(call.Args[0], "Check")
					if chk != "" {
						t.checks = append(t.checks, chk)
					}
				case "NotCheck":
					chk := funcname(call.Args[0], "Check")
					if chk != "" {
						t.checks = append(t.checks, "!"+chk)
					}
				case "Call":
					cll := funcname(call.Args[0], "Call")
					if cll != "" {
						t.calls = append(t.calls, cll)
					}
				}
			}
		}
	}
	return obj(sel.X), &t
}

func funcname(s ast.Expr, def string) string {
	switch check := s.(type) {
	case *ast.FuncLit:
		fmt.Printf("funclit: %#v", check)
		return def
	case *ast.SelectorExpr:
		return check.Sel.String()
		fmt.Printf("sel: %#v", check)
	default:
		fmt.Printf("other: %#v", check)
	}
	return ""
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
