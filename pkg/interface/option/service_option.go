package option

import (
	"fmt"
	"go/ast"
	"go/constant"
	stdtypes "go/types"
	"path"
	stdstrings "strings"

	"github.com/iancoleman/strcase"
	"golang.org/x/tools/go/types/typeutil"

	"github.com/swipe-io/swipe/pkg/domain/model"
	"github.com/swipe-io/swipe/pkg/errors"
	"github.com/swipe-io/swipe/pkg/openapi"
	"github.com/swipe-io/swipe/pkg/parser"
	"github.com/swipe-io/swipe/pkg/strings"
	"github.com/swipe-io/swipe/pkg/types"
	"github.com/swipe-io/swipe/pkg/usecase/option"
)

type ErrorData struct {
	Named *stdtypes.Named
	Code  int64
}

type serviceOption struct {
	info model.GenerateInfo
}

func (g *serviceOption) Parse(option *parser.Option) (interface{}, error) {
	o := model.ServiceOption{}

	serviceOpt := parser.MustOption(option.At("iface"))
	ifacePtr, ok := serviceOpt.Value.Type().(*stdtypes.Pointer)
	if !ok {
		return nil, errors.NotePosition(serviceOpt.Position,
			fmt.Errorf("the Interface option is required must be a pointer to an interface type; found %s", stdtypes.TypeString(serviceOpt.Value.Type(), nil)))
	}
	iface, ok := ifacePtr.Elem().Underlying().(*stdtypes.Interface)
	if !ok {
		return nil, errors.NotePosition(serviceOpt.Position,
			fmt.Errorf("the Interface option is required must be a pointer to an interface type; found %s", stdtypes.TypeString(serviceOpt.Value.Type(), nil)))
	}

	o.ID = strcase.ToCamel(stdtypes.TypeString(ifacePtr.Elem(), func(p *stdtypes.Package) string {
		return p.Name()
	}))

	if transportOpt, ok := option.At("Transport"); ok {
		transportOption, err := g.loadTransport(transportOpt)
		if err != nil {
			return nil, err
		}
		o.Transport = transportOption
	}

	o.Type = ifacePtr.Elem()
	o.Interface = iface

	errorStatusMethod := "StatusCode"
	if o.Transport.JsonRPC.Enable {
		errorStatusMethod = "ErrorCode"
	}

	hasher := typeutil.MakeHasher()

	var collectErrorReturn func(b *model.BlockStmt) []model.ErrorHTTPTransportOption

	collectErrorReturn = func(b *model.BlockStmt) (result []model.ErrorHTTPTransportOption) {
		for _, block := range b.Blocks {
			result = append(result, collectErrorReturn(block)...)
		}
		for _, stmt := range b.Returns {
			for _, e := range stmt.Results {
				switch v := e.(type) {
				case *model.ValueResult:
				case *model.CallResult:
					if v.IsIface {
						for _, m := range g.info.MapTypes {
							if decl, ok := m.DeclStmt[v.FnID]; ok {
								result = append(result, collectErrorReturn(decl.Block)...)
							}
						}
					}
				}
				if v, ok := e.(*model.ValueResult); ok {
					if m, ok := g.info.MapTypes[v.ID]; ok {
						for _, decl := range m.DeclStmt {
							if decl.Name == errorStatusMethod {
								for _, returnStmt := range decl.Block.Returns {
									for _, i2 := range returnStmt.Results {
										if v, ok := i2.(*model.ValueResult); ok && v.Value != nil {
											code, _ := constant.Int64Val(v.Value)
											var t stdtypes.Type
											if p, ok := m.Type.(*stdtypes.Pointer); ok {
												t = p.Elem()
											}
											path.Join()
											if named, ok := t.(*stdtypes.Named); ok {
												result = append(result, model.ErrorHTTPTransportOption{
													Named: named,
													Code:  code,
												})
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
		return
	}

	for i := 0; i < iface.NumMethods(); i++ {
		m := iface.Method(i)

		sig := m.Type().(*stdtypes.Signature)
		comments, _ := g.info.CommentMap.At(m.Type()).([]string)

		sm := model.ServiceMethod{
			Type:     m,
			T:        m.Type(),
			Name:     m.Name(),
			LcName:   strings.LcFirst(m.Name()),
			Comments: comments,
		}

		for _, decls := range g.info.MapTypes {
			if decl := decls.DeclStmt[types.HashObject(stdtypes.Object(m), hasher)]; decl != nil {
				errorExists := map[*stdtypes.Named]struct{}{}
				for _, e := range collectErrorReturn(decl.Block) {
					if _, ok := errorExists[e.Named]; !ok {
						sm.Errors = append(sm.Errors, e)
						errorExists[e.Named] = struct{}{}
					}
				}
				for _, e := range sm.Errors {
					o.Transport.MapCodeErrors[e.Named.Obj().Name()] = e
				}
			}
		}

		var (
			resultOffset, paramOffset int
		)
		if types.ContainsContext(sig.Params()) {
			sm.ParamCtx = sig.Params().At(0)
			paramOffset = 1
		}
		if types.ContainsError(sig.Results()) {
			sm.ReturnErr = sig.Results().At(sig.Results().Len() - 1)
			resultOffset = 1
		}
		if types.IsNamed(sig.Results()) {
			sm.ResultsNamed = true
		}
		if !sm.ResultsNamed && sig.Results().Len()-resultOffset > 1 {
			return nil, errors.NotePosition(serviceOpt.Position,
				fmt.Errorf("interface method with unnamed results cannot be greater than 1"))
		}
		for j := paramOffset; j < sig.Params().Len(); j++ {
			sm.Params = append(sm.Params, sig.Params().At(j))
		}
		for j := 0; j < sig.Results().Len()-resultOffset; j++ {
			sm.Results = append(sm.Results, sig.Results().At(j))
		}
		o.Methods = append(o.Methods, sm)
	}

	if _, ok := option.At("Logging"); ok {
		o.Logging = true
	}

	if instrumentingOpt, ok := option.At("Instrumenting"); ok {
		o.Instrumenting.Enable = true
		if namespace, ok := instrumentingOpt.At("namespace"); ok {
			o.Instrumenting.Namespace = namespace.Value.String()
		}
		if subsystem, ok := instrumentingOpt.At("subsystem"); ok {
			o.Instrumenting.Subsystem = subsystem.Value.String()
		}
	}

	return o, nil
}

func (g *serviceOption) loadTransport(opt *parser.Option) (option model.TransportOption, err error) {
	_, fastHTTP := opt.At("FastEnable")
	option = model.TransportOption{
		Protocol:      parser.MustOption(opt.At("protocol")).Value.String(),
		FastHTTP:      fastHTTP,
		MethodOptions: map[string]model.MethodHTTPTransportOption{},
		MapCodeErrors: map[string]model.ErrorHTTPTransportOption{},
		Openapi: model.OpenapiHTTPTransportOption{
			Methods: map[string]*model.OpenapiMethodOption{},
		},
	}
	if _, ok := opt.At("ClientEnable"); ok {
		option.Client.Enable = true
	}
	if _, ok := opt.At("ServerDisabled"); ok {
		option.ServerDisabled = true
	}
	if openapiDocOpt, ok := opt.At("Openapi"); ok {
		option.Openapi.Enable = true
		if v, ok := openapiDocOpt.At("OpenapiOutput"); ok {
			option.Openapi.Output = v.Value.String()
		}
		if v, ok := openapiDocOpt.At("OpenapiInfo"); ok {
			option.Openapi.Info = openapi.Info{
				Title:       parser.MustOption(v.At("title")).Value.String(),
				Description: parser.MustOption(v.At("description")).Value.String(),
				Version:     parser.MustOption(v.At("version")).Value.String(),
			}
		}
		if v, ok := openapiDocOpt.At("OpenapiContact"); ok {
			option.Openapi.Info.Contact = &openapi.Contact{
				Name:  parser.MustOption(v.At("name")).Value.String(),
				Email: parser.MustOption(v.At("email")).Value.String(),
				URL:   parser.MustOption(v.At("url")).Value.String(),
			}
		}
		if v, ok := openapiDocOpt.At("OpenapiLicence"); ok {
			option.Openapi.Info.License = &openapi.License{
				Name: parser.MustOption(v.At("name")).Value.String(),
				URL:  parser.MustOption(v.At("url")).Value.String(),
			}
		}
		if s, ok := openapiDocOpt.Slice("OpenapiServer"); ok {
			for _, v := range s {
				option.Openapi.Servers = append(option.Openapi.Servers, openapi.Server{
					Description: parser.MustOption(v.At("description")).Value.String(),
					URL:         parser.MustOption(v.At("url")).Value.String(),
				})
			}
		}
		if openapiTags, ok := openapiDocOpt.Slice("OpenapiTags"); ok {
			for _, openapiTagsOpt := range openapiTags {
				var methods []string
				if methodsOpt, ok := openapiTagsOpt.At("methods"); ok {
					for _, expr := range methodsOpt.Value.ExprSlice() {
						fnSel, ok := expr.(*ast.SelectorExpr)
						if !ok {
							return option, errors.NotePosition(methodsOpt.Position, fmt.Errorf("the %s value must be func selector", methodsOpt.Name))
						}
						methods = append(methods, fnSel.Sel.Name)
						if _, ok := option.Openapi.Methods[fnSel.Sel.Name]; !ok {
							option.Openapi.Methods[fnSel.Sel.Name] = &model.OpenapiMethodOption{}
						}
					}
				}
				if tagsOpt, ok := openapiTagsOpt.At("tags"); ok {
					if len(methods) > 0 {
						for _, method := range methods {
							option.Openapi.Methods[method].Tags = append(option.Openapi.Methods[method].Tags, tagsOpt.Value.StringSlice()...)
						}
					} else {
						option.Openapi.DefaultMethod.Tags = append(option.Openapi.DefaultMethod.Tags, tagsOpt.Value.StringSlice()...)
					}
				}
			}
		}
		if option.Openapi.Output == "" {
			option.Openapi.Output = "./"
		}
	}
	if jsonRpcOpt, ok := opt.At("JSONRPC"); ok {
		option.JsonRPC.Enable = true
		if path, ok := jsonRpcOpt.At("JSONRPCPath"); ok {
			option.JsonRPC.Path = path.Value.String()
		}
	}
	if methodDefaultOpt, ok := opt.At("MethodDefaultOptions"); ok {
		defaultMethodOptions, err := getMethodOptions(methodDefaultOpt, model.MethodHTTPTransportOption{})
		if err != nil {
			return option, err
		}
		option.DefaultMethodOptions = defaultMethodOptions
	}

	if methods, ok := opt.Slice("MethodOptions"); ok {
		for _, methodOpt := range methods {
			signOpt := parser.MustOption(methodOpt.At("signature"))
			fnSel, ok := signOpt.Value.Expr().(*ast.SelectorExpr)
			if !ok {
				return option, errors.NotePosition(signOpt.Position, fmt.Errorf("the Signature value must be func selector"))
			}
			baseMethodOpts := option.MethodOptions[fnSel.Sel.Name]
			mopt, err := getMethodOptions(methodOpt, baseMethodOpts)
			if err != nil {
				return option, err
			}
			option.MethodOptions[fnSel.Sel.Name] = mopt
		}
	}

	option.Prefix = "REST"
	if option.JsonRPC.Enable {
		option.Prefix = "JSONRPC"
	}

	return
}

func NewServiceOption(info model.GenerateInfo) option.Option {
	return &serviceOption{
		info: info,
	}
}

func getMethodOptions(methodOpt *parser.Option, baseMethodOpts model.MethodHTTPTransportOption) (model.MethodHTTPTransportOption, error) {
	if wrapResponseOpt, ok := methodOpt.At("WrapResponse"); ok {
		baseMethodOpts.WrapResponse.Enable = true
		baseMethodOpts.WrapResponse.Name = wrapResponseOpt.Value.String()
	}
	if httpMethodOpt, ok := methodOpt.At("Method"); ok {
		baseMethodOpts.MethodName = httpMethodOpt.Value.String()
		baseMethodOpts.Expr = httpMethodOpt.Value.Expr()
	}
	if path, ok := methodOpt.At("Path"); ok {
		baseMethodOpts.Path = path.Value.String()

		idxs, err := httpBraceIndices(baseMethodOpts.Path)
		if err != nil {
			return baseMethodOpts, err
		}
		if len(idxs) > 0 {
			baseMethodOpts.PathVars = make(map[string]string, len(idxs))

			var end int
			for i := 0; i < len(idxs); i += 2 {
				end = idxs[i+1]
				parts := stdstrings.SplitN(baseMethodOpts.Path[idxs[i]+1:end-1], ":", 2)

				name := parts[0]
				regexp := ""

				if len(parts) == 2 {
					regexp = parts[1]
				}
				baseMethodOpts.PathVars[name] = regexp
			}
		}
	}
	if serverRequestFunc, ok := methodOpt.At("ServerDecodeRequestFunc"); ok {
		baseMethodOpts.ServerRequestFunc.Type = serverRequestFunc.Value.Type()
		baseMethodOpts.ServerRequestFunc.Expr = serverRequestFunc.Value.Expr()
	}
	if serverResponseFunc, ok := methodOpt.At("ServerEncodeResponseFunc"); ok {
		baseMethodOpts.ServerResponseFunc.Type = serverResponseFunc.Value.Type()
		baseMethodOpts.ServerResponseFunc.Expr = serverResponseFunc.Value.Expr()
	}
	if clientRequestFunc, ok := methodOpt.At("ClientEncodeRequestFunc"); ok {
		baseMethodOpts.ClientRequestFunc.Type = clientRequestFunc.Value.Type()
		baseMethodOpts.ClientRequestFunc.Expr = clientRequestFunc.Value.Expr()
	}
	if clientResponseFunc, ok := methodOpt.At("ClientDecodeResponseFunc"); ok {
		baseMethodOpts.ClientResponseFunc.Type = clientResponseFunc.Value.Type()
		baseMethodOpts.ClientResponseFunc.Expr = clientResponseFunc.Value.Expr()
	}
	if queryVars, ok := methodOpt.At("QueryVars"); ok {
		baseMethodOpts.QueryVars = map[string]string{}
		values := queryVars.Value.StringSlice()
		for i := 0; i < len(values); i += 2 {
			baseMethodOpts.QueryVars[values[i]] = values[i+1]
		}
	}
	if headerVars, ok := methodOpt.At("HeaderVars"); ok {
		baseMethodOpts.HeaderVars = map[string]string{}
		values := headerVars.Value.StringSlice()
		for i := 0; i < len(values); i += 2 {
			baseMethodOpts.HeaderVars[values[i]] = values[i+1]
		}
	}
	return baseMethodOpts, nil
}

func httpBraceIndices(s string) ([]int, error) {
	var level, idx int
	var idxs []int
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '{':
			if level++; level == 1 {
				idx = i
			}
		case '}':
			if level--; level == 0 {
				idxs = append(idxs, idx, i+1)
			} else if level < 0 {
				return nil, fmt.Errorf("mux: unbalanced braces in %q", s)
			}
		}
	}
	if level != 0 {
		return nil, fmt.Errorf("mux: unbalanced braces in %q", s)
	}
	return idxs, nil
}
