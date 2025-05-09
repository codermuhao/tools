// Package generate generate
package generate

import (
	"fmt"
	"sort"
	"strings"

	"github.com/codermuhao/tools/cmd/protoc-gen-gin-bff/internal/util"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"

	options2 "github.com/codermuhao/tools/cmd/protoc-gen-gin-bff/api"
)

type gen struct {
	g         *protogen.Plugin
	pkgs      []pkgImport
	pkgExists map[string]struct{}
}

type pkgImport struct {
	url   string
	alias string
}

type router struct {
	method        string
	url           string
	fun           string
	group         string
	input, output string
}

var SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)

var (
	header, vars, service string
	noValidate            = map[string]bool{
		"google.protobuf.Empty": true,
	}
	message2config = map[string]struct {
		goPackage, goImport string
	}{}
)

// NewGen create a gen instance
func NewGen(g *protogen.Plugin) *gen {
	return &gen{g: g, pkgs: []pkgImport{
		{url: "github.com/gin-gonic/gin"},
	}, pkgExists: map[string]struct{}{"github.com/gin-gonic/gin": {}}}
}

// File generate codes by proto file
func (g *gen) File(file *protogen.File, release string) *protogen.GeneratedFile {
	filename := file.GeneratedFilenamePrefix + ".pb.gin.bff.go"
	gf := g.g.NewGeneratedFile(filename, file.GoImportPath)
	g.g.SupportedFeatures = SupportedFeatures
	header += fmt.Sprintf("// Code generated by protoc-gen-gin-bff. DO NOT EDIT.\n")
	header += fmt.Sprintf("// source: %s\n", *file.Proto.Name)
	header += fmt.Sprintf("// version:  %s\n\n", release)
	header += fmt.Sprintf("package %s\n", file.GoPackageName)
	vars += fmt.Sprintf("var _ = gin.New\n\n")
	for i, imps := 0, file.Desc.Imports(); i < imps.Len(); i++ {
		impFile := g.g.FilesByPath[imps.Get(i).Path()]
		if impFile.GoImportPath == file.GoImportPath {
			continue
		}
		for _, v := range impFile.Messages {
			message2config[fmt.Sprintf("%s.%s", impFile.Proto.GetPackage(), v.Desc.Name())] = struct {
				goPackage, goImport string
			}{goPackage: string(impFile.GoPackageName), goImport: impFile.GoImportPath.String()}
		}
	}
	g.genServices(file.Proto.GetService())
	gf.P(header)
	g.genImports(gf)
	gf.P(vars)
	gf.P(service)
	return gf
}

func (g *gen) genServices(srvs []*descriptorpb.ServiceDescriptorProto) {
	for _, v := range srvs {
		g.genSrvInterface(v)
		g.genServiceConstruct(v)
		g.genErrorAndSuccessFunc(v)
		g.genMiddlewares(v)
		groups := make(map[string][]*router)
		for _, vv := range v.GetMethod() {
			r := &router{fun: vv.GetName(), input: vv.GetInputType(), output: vv.GetOutputType()}
			ext := proto.GetExtension(vv.GetOptions(), annotations.E_Http)
			opts, ok := ext.(*annotations.HttpRule)
			if !ok {
				panic("extension not match http rule")
			}
			switch pattern := opts.GetPattern().(type) {
			case *annotations.HttpRule_Post:
				r.method, r.url = "POST", pattern.Post
			default:
				panic("not support http pattern")
			}
			if strings.HasPrefix(r.url, "/") == false {
				r.url = "/" + r.url
			}
			ext2 := proto.GetExtension(vv.GetOptions(), options2.E_Router)
			opts2, ok := ext2.(*options2.RouterRule)
			if !ok {
				panic("extension not match group rule")
			}
			r.group = opts2.GetGroup()
			groups[r.group] = append(groups[r.group], r)
		}
		indexes := sortGroupKey(groups)
		service += fmt.Sprintf("func (b *%sBFF) Init(router *gin.Engine) {\n", util.FirstLower(v.GetName()))
		for _, idx := range indexes {
			service += fmt.Sprintf("%sGroup := router.Group(%#v, b.middlewares[%#v]...)\n",
				util.VariableName(idx), idx, idx)
			service += fmt.Sprintf("{\n")
			for _, item := range groups[idx] {
				g.genMethod(item)
			}
			service += fmt.Sprintf("}\n")
		}
		service += fmt.Sprintf("}\n")
	}
}

func (g *gen) genSrvInterface(srv *descriptorpb.ServiceDescriptorProto) {
	service += fmt.Sprintf("type %sHandler interface {\n", srv.GetName())
	for _, v := range srv.GetMethod() {
		service += fmt.Sprintf("%s (ctx *gin.Context, req *%s) (rsp *%s, err error)\n", v.GetName(),
			g.formatType(v.GetInputType()), g.formatType(v.GetOutputType()))
	}
	service += fmt.Sprintf("}\n\n")
}

func (g *gen) formatType(from string) string {
	parts := strings.Split(from, ".")
	last := parts[len(parts)-1]
	if item, ok := message2config[strings.TrimLeft(from, ".")]; ok {
		if _, ok := g.pkgExists[strings.Trim(item.goImport, "\"")]; !ok {
			g.pkgs = append(g.pkgs, pkgImport{url: strings.Trim(item.goImport, "\"")})
			g.pkgExists[strings.Trim(item.goImport, "\"")] = struct{}{}
		}
		return item.goPackage + "." + last
	}
	return last
}

func (g *gen) genServiceConstruct(srv *descriptorpb.ServiceDescriptorProto) {
	firstLowerName := util.FirstLower(srv.GetName())
	service += fmt.Sprintf("type %sBFF struct {\n", firstLowerName)
	service += fmt.Sprintf("h %sHandler\n", srv.GetName())
	service += fmt.Sprintf("errorFunc func(*gin.Context, error)\n")
	service += fmt.Sprintf("rspFunc func(*gin.Context, interface{})\n")
	service += fmt.Sprintf("middlewares map[string][]gin.HandlerFunc\n")
	service += fmt.Sprintf("routerMiddlewares map[string][]gin.HandlerFunc\n")
	service += fmt.Sprintf("}\n\n")

	service += fmt.Sprintf("type %sBFFOptions func(*%sBFF)\n\n", firstLowerName, firstLowerName)
	service += fmt.Sprintf("func New%sBFF(h %sHandler, opts ...%sBFFOptions) *%sBFF {\n",
		srv.GetName(), srv.GetName(), firstLowerName, firstLowerName)
	if _, ok := g.pkgExists["net/http"]; !ok {
		g.pkgs = append(g.pkgs, pkgImport{url: "net/http"})
		g.pkgExists["net/http"] = struct{}{}
	}
	service += fmt.Sprintf("s := &%sBFF {\n", firstLowerName)
	service += fmt.Sprintf("h: h,\n")
	service += fmt.Sprintf("errorFunc: func(c *gin.Context, err error) {\n")
	service += fmt.Sprintf("c.Error(err)\n")
	service += fmt.Sprintf("c.AbortWithError(http.StatusInternalServerError, err)\n")
	service += fmt.Sprintf("return\n")
	service += fmt.Sprintf("},\n")
	service += fmt.Sprintf("rspFunc: func(c *gin.Context, i interface{}) {\n")
	service += fmt.Sprintf("c.JSON(http.StatusOK, i)\n")
	service += fmt.Sprintf("return\n")
	service += fmt.Sprintf("},\n")
	service += fmt.Sprintf("middlewares: make(map[string][]gin.HandlerFunc),\n")
	service += fmt.Sprintf("routerMiddlewares: make(map[string][]gin.HandlerFunc),\n")
	service += fmt.Sprintf("}\n")
	service += fmt.Sprintf("for _, opt := range opts {\n")
	service += fmt.Sprintf("opt(s)\n")
	service += fmt.Sprintf("}\n")
	service += fmt.Sprintf("return s\n")
	service += fmt.Sprintf("}\n\n")
}

func (g *gen) genErrorAndSuccessFunc(srv *descriptorpb.ServiceDescriptorProto) {
	firstLowerName := util.FirstLower(srv.GetName())
	service += fmt.Sprintf("func New%sBFFErrorFunc(f func(*gin.Context, error)) %sBFFOptions {\n",
		srv.GetName(), firstLowerName)
	service += fmt.Sprintf("return func(s *%sBFF) {\n", firstLowerName)
	service += fmt.Sprintf("s.errorFunc = func(c *gin.Context, err error) {\n")
	service += fmt.Sprintf("f(c, err)\n")
	service += fmt.Sprintf("}\n")
	service += fmt.Sprintf("}\n")
	service += fmt.Sprintf("}\n\n")
	service += fmt.Sprintf("func New%sBFFRspFunc(f func(*gin.Context, interface{})) %sBFFOptions {\n",
		srv.GetName(), firstLowerName)
	service += fmt.Sprintf("return func(s *%sBFF) {\n", firstLowerName)
	service += fmt.Sprintf("s.rspFunc = func(c *gin.Context, rsp interface{}) {\n")
	service += fmt.Sprintf("f(c, rsp)\n")
	service += fmt.Sprintf("}\n")
	service += fmt.Sprintf("}\n")
	service += fmt.Sprintf("}\n\n")
}

func (g *gen) genMiddlewares(srv *descriptorpb.ServiceDescriptorProto) {
	firstLowerName := util.FirstLower(srv.GetName())
	service += fmt.Sprintf("func (b *%sBFF) AddGroupMiddleware(group string, handler ...gin.HandlerFunc) {\n",
		firstLowerName)
	service += fmt.Sprintf("if len(group) > 0 && handler != nil {\n")
	service += fmt.Sprintf("b.middlewares[group] = append(b.middlewares[group], handler...)\n")
	service += fmt.Sprintf("}\n")
	service += fmt.Sprintf("}\n\n")
	service += fmt.Sprintf("func (b *%sBFF) AddRouterMiddleware(router string, handler ...gin.HandlerFunc) {\n",
		firstLowerName)
	service += fmt.Sprintf("if len(router) > 0 && handler != nil {\n")
	service += fmt.Sprintf("b.routerMiddlewares[router] = append(b.routerMiddlewares[router], handler...)\n")
	service += fmt.Sprintf("}\n")
	service += fmt.Sprintf("}\n\n")
}

func (g *gen) genMethod(i *router) {
	service += fmt.Sprintf("{\n")
	service += fmt.Sprintf("handlers := append(b.routerMiddlewares[%#v], func(ctx *gin.Context) {\n", i.url)
	service += fmt.Sprintf("raw, err := ctx.GetRawData()\n")
	service += fmt.Sprintf("if err != nil {\n")
	service += fmt.Sprintf("b.errorFunc(ctx, err)\n")
	service += fmt.Sprintf("return\n")
	service += fmt.Sprintf("}\n")
	service += fmt.Sprintf("if len(raw) == 0 {\n")
	service += fmt.Sprintf("raw = []byte(%#v)\n", "{}")
	service += fmt.Sprintf("}\n")
	service += fmt.Sprintf("req := new(%s)\n", g.formatType(i.input))
	if _, ok := g.pkgExists["git.woa.com/enbox/enkits/xjson"]; !ok {
		g.pkgs = append(g.pkgs, pkgImport{url: "git.woa.com/enbox/enkits/xjson"})
		g.pkgExists["git.woa.com/enbox/enkits/xjson"] = struct{}{}
	}
	service += fmt.Sprintf("if err := xjson.Unmarshal(raw, req); err != nil {\n")
	service += fmt.Sprintf("b.errorFunc(ctx, err)\n")
	service += fmt.Sprintf("return\n")
	service += fmt.Sprintf("}\n")
	if !noValidate[strings.Trim(i.input, ".")] {
		service += fmt.Sprintf("if err := req.Validate(); err != nil {\n")
		service += fmt.Sprintf("if e, ok := err.(%sValidationError); ok {\n", g.formatType(i.input))
		if _, ok := g.pkgExists["git.woa.com/enbox/enkits/xerrors"]; !ok {
			g.pkgs = append(g.pkgs, pkgImport{url: "git.woa.com/enbox/enkits/xerrors"})
			g.pkgExists["git.woa.com/enbox/enkits/xerrors"] = struct{}{}
		}
		service += fmt.Sprintf("b.errorFunc(ctx, xerrors.NewReasonError(%#v+e.Field(), e.Error()))\n",
			"InvalidParameter.")
		service += fmt.Sprintf("return\n")
		service += fmt.Sprintf("}\n")
		service += fmt.Sprintf("b.errorFunc(ctx, err)\n")
		service += fmt.Sprintf("return\n")
		service += fmt.Sprintf("}\n")
	}
	service += fmt.Sprintf("rsp, err := b.h.%s(ctx, req)\n", i.fun)
	service += fmt.Sprintf("if err != nil {\n")
	service += fmt.Sprintf("b.errorFunc(ctx, err)\n")
	service += fmt.Sprintf("return\n")
	service += fmt.Sprintf("}\n")
	service += fmt.Sprintf("if b.rspFunc != nil {\n")
	service += fmt.Sprintf("b.rspFunc(ctx, rsp)\n")
	service += fmt.Sprintf("}\n")
	service += fmt.Sprintf("return\n")
	service += fmt.Sprintf("})\n")
	if _, ok := g.pkgExists["strings"]; !ok {
		g.pkgs = append(g.pkgs, pkgImport{url: "strings"})
		g.pkgExists["strings"] = struct{}{}
	}
	service += fmt.Sprintf(
		"%sGroup.%s(strings.TrimPrefix(%#v, %#v), handlers...)\n",
		util.VariableName(i.group), i.method, i.url, i.group,
	)
	service += fmt.Sprintf("}\n")
}

func (g *gen) genImports(gf *protogen.GeneratedFile) {
	gf.P("import (")
	for _, v := range g.pkgs {
		if len(v.alias) > 0 {
			gf.P(fmt.Sprintf("%s \"%s\"", v.alias, v.url))
		} else {
			gf.P(fmt.Sprintf("\"%s\"", v.url))
		}
	}
	gf.P(")")
}

func sortGroupKey(groups map[string][]*router) []string {
	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
