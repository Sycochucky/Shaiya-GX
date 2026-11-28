//go:build ignore

// AST inventory for generate_server_reference.py; run from the repository root.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Record struct {
	Path      string   `json:"path"`
	Package   string   `json:"package"`
	Kind      string   `json:"kind"`
	Name      string   `json:"name"`
	Line      int      `json:"line"`
	Doc       string   `json:"doc,omitempty"`
	Source    string   `json:"source,omitempty"`
	Signature string   `json:"signature,omitempty"`
	Imports   []string `json:"imports,omitempty"`
	Calls     []string `json:"calls,omitempty"`
	Value     string   `json:"value,omitempty"`
	Ops       []string `json:"ops,omitempty"`
}

func main() {
	fset := token.NewFileSet()
	var records []Record
	for _, root := range []string{"server/ai", "server/cmd", "server/internal", "internal/proto", "internal/game", "internal/session", "internal/authwire", "internal/transport", "internal/itemstack", "internal/netx", "internal/textwire"} {
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() || !strings.HasSuffix(path, ".go") {
				return nil
			}
			file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
			if err != nil {
				return err
			}
			raw, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			nodeText := func(n ast.Node) string {
				if n == nil {
					return ""
				}
				return strings.ReplaceAll(string(raw[fset.Position(n.Pos()).Offset:fset.Position(n.End()).Offset]), "\r\n", "\n")
			}
			formatted := func(n ast.Node) string { var b bytes.Buffer; _ = printer.Fprint(&b, fset, n); return b.String() }
			var imports []string
			aliases := map[string]string{}
			for _, im := range file.Imports {
				v, _ := strconv.Unquote(im.Path.Value)
				imports = append(imports, v)
				name := filepath.Base(v)
				if im.Name != nil {
					name = im.Name.Name
				}
				aliases[name] = v
			}
			qualify := func(e ast.Expr) string {
				if s, ok := e.(*ast.SelectorExpr); ok {
					if id, ok := s.X.(*ast.Ident); ok {
						return aliases[id.Name] + "." + s.Sel.Name
					}
				}
				if id, ok := e.(*ast.Ident); ok {
					return filepath.ToSlash(filepath.Dir(path)) + "." + id.Name
				}
				return nodeText(e)
			}
			base := Record{Path: filepath.ToSlash(path), Package: file.Name.Name, Kind: "file", Line: 1, Imports: imports}
			if file.Doc != nil {
				base.Doc = file.Doc.Text()
			}
			records = append(records, base)
			for _, decl := range file.Decls {
				switch n := decl.(type) {
				case *ast.GenDecl:
					typ := ""
					for _, spec := range n.Specs {
						switch s := spec.(type) {
						case *ast.TypeSpec:
							r := base
							r.Kind = "type"
							r.Name = s.Name.Name
							r.Doc = ""
							r.Line = fset.Position(s.Pos()).Line
							r.Source = nodeText(s)
							r.Imports = nil
							if n.Doc != nil {
								r.Doc = n.Doc.Text()
							}
							if s.Doc != nil {
								r.Doc = s.Doc.Text()
							}
							records = append(records, r)
						case *ast.ValueSpec:
							if s.Type != nil {
								typ = nodeText(s.Type)
							} else if len(s.Values) > 0 {
								typ = ""
							}
							if n.Tok == token.CONST || n.Tok == token.VAR {
								for i, name := range s.Names {
									r := base
									r.Kind = n.Tok.String()
									r.Name = name.Name
									if i < len(s.Values) {
										r.Value = nodeText(s.Values[i])
									}
									r.Source = nodeText(s)
									r.Signature = typ
									r.Line = fset.Position(s.Pos()).Line
									r.Imports = nil
									r.Doc = ""
									if s.Doc != nil {
										r.Doc = s.Doc.Text()
									} else if n.Doc != nil {
										r.Doc = n.Doc.Text()
									}
									if s.Comment != nil {
										r.Doc += s.Comment.Text()
									}
									records = append(records, r)
								}
							}
						}
					}
				case *ast.FuncDecl:
					if n.Body == nil {
						continue
					}
					r := base
					r.Kind = "func"
					r.Name = n.Name.Name
					r.Line = fset.Position(n.Pos()).Line
					r.Source = nodeText(n)
					r.Imports = nil
					r.Doc = ""
					r.Signature = strings.TrimSpace(string(raw[fset.Position(n.Pos()).Offset:fset.Position(n.Body.Pos()).Offset]))
					if n.Doc != nil {
						r.Doc = n.Doc.Text()
					}
					ast.Inspect(n.Body, func(v ast.Node) bool {
						if c, ok := v.(*ast.CallExpr); ok {
							ref := qualify(c.Fun)
							r.Calls = append(r.Calls, ref)
							call := Record{Path: base.Path, Package: base.Package, Kind: "call", Name: n.Name.Name, Line: fset.Position(c.Pos()).Line, Value: ref}
							for _, arg := range c.Args {
								call.Ops = append(call.Ops, qualify(arg))
							}
							records = append(records, call)
							if sel, ok := c.Fun.(*ast.SelectorExpr); ok {
								if (sel.Sel.Name == "New" || sel.Sel.Name == "NewWriter") && len(c.Args) > 0 {
									if strings.Contains(ref, "/packet.New") || strings.Contains(ref, "/proto.NewWriter") {
										r.Ops = append(r.Ops, qualify(c.Args[0]))
									}
								}
							}
							if id, ok := c.Fun.(*ast.Ident); ok && id.Name == "NewWriter" && filepath.ToSlash(filepath.Dir(path)) == "internal/proto" && len(c.Args) > 0 {
								r.Ops = append(r.Ops, qualify(c.Args[0]))
							}
						}
						return true
					})
					records = append(records, r)
					ast.Inspect(n.Body, func(v ast.Node) bool {
						sw, ok := v.(*ast.SwitchStmt)
						if !ok {
							return true
						}
						for _, clause := range sw.Body.List {
							c := clause.(*ast.CaseClause)
							if len(c.List) == 0 {
								continue
							}
							cr := base
							cr.Kind = "case"
							cr.Name = n.Name.Name
							cr.Line = fset.Position(c.Pos()).Line
							cr.Source = nodeText(c)
							cr.Signature = formatted(sw.Tag)
							cr.Imports = nil
							cr.Doc = ""
							for _, expr := range c.List {
								cr.Ops = append(cr.Ops, qualify(expr))
							}
							ast.Inspect(c, func(v ast.Node) bool {
								if call, ok := v.(*ast.CallExpr); ok {
									cr.Calls = append(cr.Calls, qualify(call.Fun))
								}
								return true
							})
							records = append(records, cr)
						}
						return true
					})
				}
			}
			return nil
		})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	if err := json.NewEncoder(os.Stdout).Encode(records); err != nil {
		panic(err)
	}
}
