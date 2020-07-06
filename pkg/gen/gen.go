package gen

import (
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	stdtypes "go/types"
	"os"
	"path/filepath"
	"sort"
	stdstrings "strings"
	"sync"

	"github.com/swipe-io/swipe/pkg/domain/model"
	"github.com/swipe-io/swipe/pkg/file"
	"github.com/swipe-io/swipe/pkg/importer"
	"github.com/swipe-io/swipe/pkg/parser"
	"github.com/swipe-io/swipe/pkg/registry"
	"github.com/swipe-io/swipe/pkg/types"
	"github.com/swipe-io/swipe/pkg/usecase/processor"
	"github.com/swipe-io/swipe/pkg/value"

	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/types/typeutil"
)

type importerer interface {
	SetImporter(*importer.Importer)
}

type Result struct {
	PkgPath    string
	OutputPath string
	Content    []byte
	Errs       []error
}

type Swipe struct {
	ctx         context.Context
	version     string
	wd          string
	env         []string
	patterns    []string
	commentMaps *typeutil.Map
	pkgs        []*packages.Package
	returnTypes map[uint32][]interface{}
	mapTypes    map[uint32]*model.DeclType
	allPkgs     []*packages.Package
}

func (s *Swipe) Generate() ([]Result, []error) {
	r := registry.NewRegistry()

	errs := s.loadPackages()
	if len(errs) > 0 {
		return nil, errs
	}

	var result []Result
	files := make(map[string]*file.File)
	basePaths := map[string]struct{}{}

	for _, pkg := range s.pkgs {
		importerFactory := processor.NewImporterFactory(pkg)

		basePath, err := s.detectBasePath(pkg.GoFiles)
		if err != nil {
			return nil, []error{err}
		}

		basePaths[basePath] = struct{}{}

		for _, f := range pkg.Syntax {
			for _, decl := range f.Decls {
				switch decl := decl.(type) {
				case *ast.FuncDecl:
					call, err := s.findInjector(pkg.TypesInfo, decl)
					if err != nil {
						return nil, []error{err}
					}
					if call != nil {
						opt, err := parser.NewParser(pkg).Parse(call.Args[0])
						if err != nil {
							return nil, []error{err}
						}

						info := model.GenerateInfo{
							Pkg:         pkg,
							Pkgs:        s.pkgs,
							BasePath:    basePath,
							Version:     s.version,
							CommentMap:  s.commentMaps,
							ReturnTypes: s.returnTypes,
							MapTypes:    s.mapTypes,
						}

						option := r.Option(opt.Name, info)
						if option == nil {
							return nil, []error{errors.New("unknown option:" + opt.Name)}
						}

						o, err := option.Parse(opt)
						if err != nil {
							return nil, []error{err}
						}
						processor, err := r.Processor(opt.Name, info)
						if err != nil {
							return nil, []error{err}
						}
						if !processor.SetOption(o) {
							return nil, []error{errors.New("option not suitable for processor: " + opt.Name)}
						}
						for _, g := range processor.Generators() {
							if err := g.Prepare(s.ctx); err != nil {
								return nil, []error{err}
							}
							outputDir := g.OutputDir()
							if outputDir == "" {
								outputDir = basePath
							}
							filename := g.Filename()
							if filename == "" {
								filename = "swipe_gen.go"
							}

							fileKey := outputDir + filename

							i := importerFactory.Instance(fileKey)
							if is, ok := g.(importerer); ok {
								is.SetImporter(i)
							}

							if err := g.Process(s.ctx); err != nil {
								return nil, []error{err}
							}

							f, ok := files[fileKey]
							if !ok {
								f = &file.File{
									PkgName:   pkg.Name,
									PkgPath:   pkg.PkgPath,
									OutputDir: outputDir,
									Filename:  filename,
									Version:   s.version,
									Importer:  i,
								}
								files[fileKey] = f
							}

							b := g.Bytes()
							if len(b) > 0 {
								_, _ = f.Write(b)
							}
						}

						continue
					}
				case *ast.GenDecl:
					if decl.Tok == token.IMPORT {
						continue
					}
				}
			}
		}
	}

	for path := range basePaths {
		files, err := filepath.Glob(filepath.Join(path, "*_gen.*"))
		if err != nil {
			panic(err)
		}
		for _, f := range files {
			if err := os.Remove(f); err != nil {
				return nil, []error{err}
			}
		}
	}

	for _, f := range files {
		if len(f.Bytes()) > 0 {
			goSrc, err := f.Frame()
			if err != nil {
				f.Errs = append(f.Errs, err)
			}
			result = append(result, Result{
				PkgPath:    f.PkgPath,
				OutputPath: filepath.Join(f.OutputDir, f.Filename),
				Content:    goSrc,
				Errs:       f.Errs,
			})
		}
	}

	return result, nil
}

func (s *Swipe) findInjector(info *stdtypes.Info, fn *ast.FuncDecl) (*ast.CallExpr, error) {
	if fn.Body == nil {
		return nil, nil
	}
	for _, stmt := range fn.Body.List {
		switch stmt := stmt.(type) {
		case *ast.ExprStmt:
			call, ok := stmt.X.(*ast.CallExpr)
			if !ok {
				continue
			}
			obj := value.QualifiedIdentObject(info, call.Fun)
			if obj == nil || obj.Pkg() == nil {
				continue
			}
			if obj.Name() != "Build" {
				continue
			}
			return call, nil
		case *ast.EmptyStmt:

			return nil, nil
		}
	}
	return nil, nil
}

func (s *Swipe) detectBasePath(paths []string) (string, error) {
	if len(paths) == 0 {
		return "", errors.New("no files to derive output directory from")
	}
	dir := filepath.Dir(paths[0])
	for _, p := range paths[1:] {
		if dir2 := filepath.Dir(p); dir2 != dir {
			return "", fmt.Errorf("found conflicting directories %q and %q", dir, dir2)
		}
	}
	return dir, nil
}

func (s *Swipe) loadPackages() []error {
	cfg := &packages.Config{
		Context:    s.ctx,
		Mode:       packages.NeedDeps | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedTypes | packages.NeedTypesSizes | packages.NeedImports | packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles,
		Dir:        s.wd,
		Env:        s.env,
		BuildFlags: []string{"-tags=swipe"},
	}

	var (
		errs []error
		err  error
	)

	escaped := make([]string, len(s.patterns))
	for i := range s.patterns {
		escaped[i] = "pattern=" + s.patterns[i]
	}
	s.pkgs, err = packages.Load(cfg, escaped...)
	if err != nil {
		return []error{err}
	}

	seen := new(sync.Map)

	var visit func(pkg *packages.Package)
	visit = func(pkg *packages.Package) {
		if _, ok := seen.Load(pkg); !ok {
			seen.Store(pkg, true)
			var importPaths []string
			for path := range pkg.Imports {
				importPaths = append(importPaths, path)
			}
			sort.Strings(importPaths)
			for _, path := range importPaths {
				visit(pkg.Imports[path])
			}
			s.allPkgs = append(s.allPkgs, pkg)
		}
	}
	for _, pkg := range s.pkgs {
		visit(pkg)
	}

	hasher := typeutil.MakeHasher()

	visitReturnStmt := func(p *packages.Package, ret *ast.ReturnStmt) *model.ReturnStmt {
		returnStmt := &model.ReturnStmt{}
		for _, result := range ret.Results {
			switch v := result.(type) {
			case *ast.CallExpr:
				if selExpr, ok := v.Fun.(*ast.SelectorExpr); ok {
					if xSelExpr, ok := selExpr.X.(*ast.SelectorExpr); ok {
						id := hasher.Hash(p.TypesInfo.ObjectOf(xSelExpr.Sel).Type())
						fnID := types.HashObject(p.TypesInfo.ObjectOf(selExpr.Sel), hasher)
						_, isIface := p.TypesInfo.ObjectOf(xSelExpr.Sel).Type().Underlying().(*stdtypes.Interface)
						returnStmt.Results = append(returnStmt.Results, &model.CallResult{
							ID:      id,
							FnID:    fnID,
							IsIface: isIface,
						})
					}
				}
			case *ast.SelectorExpr:
			case *ast.Ident:
			}
			if tv, ok := p.TypesInfo.Types[result]; ok {
				returnStmt.Results = append(returnStmt.Results, &model.ValueResult{
					ID:    hasher.Hash(tv.Type),
					Type:  tv.Type,
					Value: tv.Value,
				})
			}
		}
		return returnStmt
	}

	var visitBlockStmt func(p *packages.Package, block *ast.BlockStmt) *model.BlockStmt
	visitBlockStmt = func(p *packages.Package, block *ast.BlockStmt) *model.BlockStmt {
		bs := &model.BlockStmt{}
		for _, stmt := range block.List {
			switch v := stmt.(type) {
			case *ast.SelectStmt:
				bs.Blocks = append(bs.Blocks, visitBlockStmt(p, v.Body))
			case *ast.RangeStmt:
				bs.Blocks = append(bs.Blocks, visitBlockStmt(p, v.Body))
			case *ast.ForStmt:
				bs.Blocks = append(bs.Blocks, visitBlockStmt(p, v.Body))
			case *ast.TypeSwitchStmt:
				bs.Blocks = append(bs.Blocks, visitBlockStmt(p, v.Body))
			case *ast.SwitchStmt:
				bs.Blocks = append(bs.Blocks, visitBlockStmt(p, v.Body))
			case *ast.IfStmt:
				bs.Blocks = append(bs.Blocks, visitBlockStmt(p, v.Body))
			case *ast.BlockStmt:
				bs.Blocks = append(bs.Blocks, visitBlockStmt(p, v))
			case *ast.ReturnStmt:
				bs.Returns = append(bs.Returns, visitReturnStmt(p, v))
			}
		}
		return bs
	}

	types.Inspect(s.pkgs, func(p *packages.Package, n ast.Node) bool {
		if declStmt, ok := n.(*ast.FuncDecl); ok {
			if declStmt.Recv != nil {

				declType := p.TypesInfo.TypeOf(declStmt.Recv.List[0].Type)

				id := hasher.Hash(declType)

				if _, ok := s.mapTypes[id]; !ok {
					s.mapTypes[id] = &model.DeclType{DeclStmt: map[uint32]*model.DeclStmt{}}
				}

				fnID := types.HashObject(p.TypesInfo.ObjectOf(declStmt.Name), hasher)

				s.mapTypes[id].Type = declType
				s.mapTypes[id].DeclStmt[fnID] = &model.DeclStmt{Name: declStmt.Name.Name, Block: visitBlockStmt(p, declStmt.Body)}
			}
		}
		return true
	})

	types.Inspect(s.pkgs, func(p *packages.Package, n ast.Node) bool {

		//if declStmt, ok := n.(*ast.FuncDecl); ok {
		//	if declStmt.Recv != nil {
		//		declObj := p.TypesInfo.ObjectOf(declStmt.Name)
		//
		//		var nodes []interface{}
		//
		//		ast.Inspect(declStmt, func(nn ast.Node) bool {
		//			if retStmt, ok := nn.(*ast.ReturnStmt); ok {
		//				for _, expr := range retStmt.Results {
		//					if callExpr, ok := expr.(*ast.CallExpr); ok {
		//						obj := p.TypesInfo.ObjectOf(callExpr.Fun.(*ast.SelectorExpr).Sel)
		//						nodes = append(nodes, types.HashObject(obj, hasher))
		//					} else {
		//						nodes = append(nodes, model.ReturnType{
		//							Expr: expr,
		//							Type: p.TypesInfo.TypeOf(expr),
		//						})
		//					}
		//				}
		//			}
		//			return true
		//		})
		//		s.returnTypes[types.HashObject(declObj, hasher)] = nodes
		//	}
		//}

		if spec, ok := n.(*ast.Field); ok {
			t := p.TypesInfo.TypeOf(spec.Type)
			if t != nil {
				var comments []string
				if spec.Doc != nil {
					for _, comment := range spec.Doc.List {
						comments = append(comments, stdstrings.TrimLeft(comment.Text, "/"))
					}
				}
				if spec.Comment != nil {
					for _, comment := range spec.Comment.List {
						comments = append(comments, stdstrings.TrimLeft(comment.Text, "/"))
					}
				}
				if len(comments) > 0 {
					s.commentMaps.Set(t, comments)
				}
			}
		}
		return true
	})

	for _, p := range s.pkgs {
		for _, e := range p.Errors {
			errs = append(errs, e)
		}
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

func NewSwipe(ctx context.Context, version, wd string, env []string, patterns []string) *Swipe {
	return &Swipe{
		ctx:         ctx,
		version:     version,
		wd:          wd,
		env:         env,
		patterns:    patterns,
		commentMaps: new(typeutil.Map),
		returnTypes: map[uint32][]interface{}{},
		mapTypes:    map[uint32]*model.DeclType{},
	}
}
