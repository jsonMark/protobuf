package gozero

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"text/template"

	"github.com/Mikaelemmmm/protobuf/internal/descriptor"
	pluginOptions "github.com/Mikaelemmmm/protobuf/plugin/gen/go/v1"
	"github.com/Mikaelemmmm/protobuf/plugin/protoc-gen-zeroapi/internal/gozero/utils"
	"google.golang.org/protobuf/proto"
)

const (
	defaultRootDir = "app/gateway"
	defaultApp     = "Gateway"
	templateDir    = "template"

	confYamlDir = "etc"
	configDir   = "internal/config"
	handlerDir  = "internal/handler"
	logicDir    = "internal/logic"
	svcDir      = "internal/svc"

	prefixMiddlewareKeyFormat = "%s_%s"
	defaultMiddleware         = "default"
)

var (
	defaultDirs = []string{
		confYamlDir,
		configDir,
		handlerDir,
		logicDir,
		svcDir,
	}

	tplRoutesImports = make(map[string]struct{})
	tplRoutes        = make(map[string]*routes, 0)
	tplHandlers      = make(map[string]*handlers, 0)
	tplLogics        = make(map[string]*logics, 0)
)

type (
	Params struct {
		RootDir     string
		AppDir      string
		App         string
		TemplateDir string
	}

	// route
	route struct {
		Method  string
		Path    string
		Handler string
	}
	routes struct {
		Prefix         string
		HasMiddleware  bool
		Middleware     string
		MiddlewareRule string
		Routes         []route
	}

	// handler
	handler struct {
		Name              string
		InputType         string
		InputOutputGoPath string
		PackageName       string
	}
	handlers struct {
		Handlers []handler
	}

	// logic
	logic struct {
		Name              string
		InputType         string
		OutputType        string
		InputOutputGoPath string
		PackageName       string
	}
	logics struct {
		Logics []logic
	}
)

type Option func(*Params)

type generator struct {
	params *Params
}

func New(ops ...Option) *generator {

	p := &Params{
		RootDir:     defaultRootDir,
		App:         defaultApp,
		TemplateDir: templateDir,
	}

	for _, o := range ops {
		o(p)
	}

	// template
	p.TemplateDir = strings.TrimRight(p.TemplateDir, "/")

	// app dir
	rootDirs := strings.Split(p.RootDir, "/")
	p.AppDir = rootDirs[len(rootDirs)-1]

	return &generator{
		params: p,
	}
}

func (g *generator) Generate(targets []*descriptor.File) error {

	if err := g.collateData(targets); err != nil {
		return err
	}

	if err := g.genDefaultDir(); err != nil {
		return err
	}

	if err := g.genMainFile(); err != nil {
		return err
	}

	if err := g.genConfYamlFile(); err != nil {
		return err
	}

	if err := g.genConfigFile(); err != nil {
		return err
	}

	if err := g.genSvcCtxFile(); err != nil {
		return err
	}

	if err := g.genRoutesFile(); err != nil {
		return err
	}

	if err := g.genHandlerFiles(); err != nil {
		return err
	}

	return g.genLogicFiles()
}

func (g *generator) collateData(targets []*descriptor.File) error {

	for _, file := range targets {
		for _, service := range file.Services {
			if err := g.collateService(file.GoPkg, service); err != nil {
				return err
			}
		}
	}

	return nil
}

func (g *generator) collateService(goPkg descriptor.GoPackage, service *descriptor.Service) error {

	serviceOptions, _ := proto.GetExtension(service.ServiceDescriptorProto.Options, pluginOptions.E_ApiOptions).(*pluginOptions.ApiOptions)

	if err := g.collateRoutesFile(service, serviceOptions); err != nil {
		return err
	}

	if err := g.collateHandlerFiles(goPkg, service, serviceOptions); err != nil {
		return err
	}

	return g.collateLogicFiles(goPkg, service, serviceOptions)

}

func (g *generator) collateRoutesFile(service *descriptor.Service, apiOptions *pluginOptions.ApiOptions) error {

	// route top import handler path
	importPath := fmt.Sprintf("%s/%s/%s", g.params.RootDir, handlerDir, apiOptions.Group)
	tplRoutesImports[importPath] = struct{}{}

	//prefix
	apiOptions.Prefix = fmt.Sprintf("/%s", strings.TrimLeft(apiOptions.Prefix, "/"))

	// routes
	if len(apiOptions.Middlewares) > 0 {
		for _, mid := range apiOptions.Middlewares {
			prefixMiddlewareKey := fmt.Sprintf(prefixMiddlewareKeyFormat, apiOptions.Prefix, mid.Middleware)
			_, ok := tplRoutes[prefixMiddlewareKey]
			if !ok {
				tplRoutes[prefixMiddlewareKey] = &routes{
					Prefix:         apiOptions.Prefix,
					Middleware:     getSvcCtxMiddleware(mid.Middleware),
					MiddlewareRule: mid.Rule,
					HasMiddleware:  true,
				}
			}
		}
	}

	//apiOptions.Middlewares[i].Rule 是否符合规则
	for _, method := range service.Methods {
		curRoutes, err := getValidRoutes(apiOptions.Prefix, *method.Name)
		if err != nil {
			return err
		}
		curRoutes.Routes = append(curRoutes.Routes, route{
			Method:  g.getHttpMethod(method.Bindings[0].HTTPMethod),
			Path:    method.Bindings[0].PathTmpl.Template,
			Handler: fmt.Sprintf("%s.%sHandler", apiOptions.Group, *method.Name),
		})
	}

	return nil
}

func getSvcCtxMiddleware(middlewares string) string {

	if len(middlewares) > 0 {
		middlewaresArr := strings.Split(middlewares, ",")
		var tmpMiddleware []string
		for _, val := range middlewaresArr {
			tmpMiddleware = append(tmpMiddleware, fmt.Sprintf("serverCtx.%s", utils.UpFirstCamelCase(val)))
		}

		return strings.Join(tmpMiddleware, ",")
	}

	return middlewares
}

func getValidRoutes(servicePrefix, methodName string) (*routes, error) {

	for _, curRoutes := range tplRoutes {
		isValid, err := regexp.MatchString(curRoutes.MiddlewareRule, methodName)
		if err != nil {
			return nil, err
		}
		if isValid && curRoutes.Prefix == servicePrefix {
			return curRoutes, nil
		}
	}

	return getDefaultValidRoutes(servicePrefix), nil
}

// 默认的routes
func getDefaultValidRoutes(servicePrefix string) *routes {

	defaultPrefixMiddlewareKey := fmt.Sprintf(prefixMiddlewareKeyFormat, servicePrefix, defaultMiddleware)
	_, ok := tplRoutes[defaultPrefixMiddlewareKey]
	if !ok {
		tplRoutes[defaultPrefixMiddlewareKey] = &routes{
			Prefix:         servicePrefix,
			MiddlewareRule: defaultMiddleware,
		}
	}
	return tplRoutes[defaultPrefixMiddlewareKey]
}

func (g *generator) collateHandlerFiles(goPkg descriptor.GoPackage, service *descriptor.Service, apiOptions *pluginOptions.ApiOptions) error {

	_, ok := tplHandlers[apiOptions.Group]
	if !ok {
		tplHandlers[apiOptions.Group] = &handlers{}
	}

	groupHandlers := tplHandlers[apiOptions.Group]

	for _, method := range service.Methods {
		if len(method.Bindings) > 0 {
			inputTypeTmp := strings.Split(*method.InputType, ".")
			inputType := inputTypeTmp[len(inputTypeTmp)-1]

			groupHandlers.Handlers = append(groupHandlers.Handlers, handler{
				Name:              *method.Name,
				InputType:         inputType,
				InputOutputGoPath: goPkg.Path,
				PackageName:       apiOptions.Group,
			})
		}
	}

	return nil
}

func (g *generator) collateLogicFiles(goPkg descriptor.GoPackage, service *descriptor.Service, apiOptions *pluginOptions.ApiOptions) error {

	_, ok := tplLogics[apiOptions.Group]
	if !ok {
		tplLogics[apiOptions.Group] = &logics{}
	}
	groupLogic := tplLogics[apiOptions.Group]

	for _, method := range service.Methods {
		if len(method.Bindings) > 0 {

			inputTypeTmp := strings.Split(*method.InputType, ".")
			inputType := inputTypeTmp[len(inputTypeTmp)-1]

			outTypeTmp := strings.Split(*method.OutputType, ".")
			outType := outTypeTmp[len(outTypeTmp)-1]

			groupLogic.Logics = append(groupLogic.Logics, logic{
				Name:              *method.Name,
				InputType:         inputType,
				OutputType:        outType,
				InputOutputGoPath: goPkg.Path,
				PackageName:       apiOptions.Group,
			})
		}
	}

	return nil

}

func (g *generator) genDefaultDir() error {

	for _, dir := range defaultDirs {
		if err := utils.CreateNoExistsDir(fmt.Sprintf("%s/%s", g.params.AppDir, dir)); err != nil {
			return err
		}
	}

	return nil
}

func (g *generator) genMainFile() error {

	fileName := fmt.Sprintf("%s/%s.go", g.params.AppDir, utils.SnakeCase(g.params.App))

	if utils.FileOrDirExists(fileName) {
		return nil
	}

	f, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}
	defer f.Close()

	tpl, err := template.ParseFiles(g.params.TemplateDir + "/main.tpl")
	if err != nil {
		return err
	}

	return tpl.Execute(f, struct {
		RootDir string
	}{
		RootDir: g.params.RootDir,
	})

}

func (g *generator) genConfYamlFile() error {

	fileName := fmt.Sprintf("%s/%s/%s.yaml", g.params.AppDir, confYamlDir, utils.SnakeCase(g.params.App))

	if utils.FileOrDirExists(fileName) {
		return nil
	}

	f, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}
	defer f.Close()

	tpl, err := template.ParseFiles(g.params.TemplateDir + "/conf.yaml.tpl")
	if err != nil {
		return err
	}

	return tpl.Execute(f, struct {
		App string
	}{
		App: utils.UpFirstCamelCase(g.params.App),
	})
}

func (g *generator) genConfigFile() error {

	fileName := fmt.Sprintf("%s/%s/config.go", g.params.AppDir, configDir)
	if utils.FileOrDirExists(fileName) {
		return nil
	}

	f, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}
	defer f.Close()

	tpl, err := template.ParseFiles(g.params.TemplateDir + "/config.tpl")
	if err != nil {
		return err
	}

	return tpl.Execute(f, struct {
		App string
	}{
		App: utils.UpFirstCamelCase(g.params.App),
	})
}
func (g *generator) genSvcCtxFile() error {

	fileName := fmt.Sprintf("%s/%s/serviceContext.go", g.params.AppDir, svcDir)

	if utils.FileOrDirExists(fileName) {
		return nil
	}

	f, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}
	defer f.Close()

	tpl, err := template.ParseFiles(g.params.TemplateDir + "/svc.tpl")
	if err != nil {
		return err
	}

	return tpl.Execute(f, struct {
		RootDir string
	}{
		RootDir: g.params.RootDir,
	})
}

func (g *generator) genRoutesFile() error {

	var routers []*routes
	for _, r := range tplRoutes {
		// except has middleware , but no valid method
		if len(r.Routes) > 0 {
			routers = append(routers, r)
		}
	}
	if len(routers) > 0 {
		sort.Slice(routers, func(i, j int) bool {
			return routers[i].Prefix > routers[j].Prefix
		})
	}

	// gen
	fileName := fmt.Sprintf("%s/%s/routes.go", g.params.AppDir, handlerDir)

	if exists := utils.FileOrDirExists(fileName); exists {
		if err := os.Remove(fileName); err != nil {
			return err
		}
	}

	f, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}
	defer f.Close()

	tpl, err := template.ParseFiles(g.params.TemplateDir + "/routes.tpl")
	if err != nil {
		return err
	}

	return tpl.Execute(f, struct {
		Routes  []*routes
		RootDir string
		Imports map[string]struct{}
	}{
		Routes:  routers,
		RootDir: g.params.RootDir,
		Imports: tplRoutesImports,
	})
}

func (g *generator) genHandlerFiles() error {

	for group, curHandlers := range tplHandlers {

		dirPath := fmt.Sprintf("%s/%s/%s", g.params.AppDir, handlerDir, group)
		if err := utils.CreateNoExistsDir(dirPath); err != nil {
			return err
		}

		for _, curHandler := range curHandlers.Handlers {
			fileName := fmt.Sprintf("%s/%sHandler.go", dirPath, utils.SnakeCase(curHandler.Name))
			if utils.FileOrDirExists(fileName) {
				continue
			}

			f, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0755)
			if err != nil {
				return err
			}

			tpl, err := template.ParseFiles(g.params.TemplateDir + "/handler.tpl")
			if err != nil {
				return err
			}
			err = tpl.Execute(f, struct {
				RootDir string
				Handler handler
				Package string
			}{
				RootDir: g.params.RootDir,
				Handler: curHandler,
			})
			_ = f.Close()
			if err != nil {
				return err
			}
		}
	}

	return nil

}

func (g *generator) genLogicFiles() error {

	for group, curLogics := range tplLogics {

		dirPath := fmt.Sprintf("%s/%s/%s", g.params.AppDir, logicDir, group)
		if err := utils.CreateNoExistsDir(dirPath); err != nil {
			return err
		}

		for _, curLogic := range curLogics.Logics {
			fileName := fmt.Sprintf("%s/%sLogic.go", dirPath, utils.SnakeCase(curLogic.Name))
			if utils.FileOrDirExists(fileName) {
				continue
			}

			f, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0755)
			if err != nil {
				return err
			}

			tpl, err := template.ParseFiles(g.params.TemplateDir + "/logic.tpl")
			if err != nil {
				return err
			}
			err = tpl.Execute(f, struct {
				RootDir string
				Logic   logic
			}{
				RootDir: g.params.RootDir,
				Logic:   curLogic,
			})
			_ = f.Close()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (g *generator) getHttpMethod(strMethod string) string {

	upMethod := strings.ToUpper(strMethod)
	switch upMethod {
	case "GET":
		return "http.MethodGet"
	case "POST":
		return "http.MethodPost"
	case "PUT":
		return "http.MethodPut"
	case "PATCH":
		return "http.MethodPatch"
	case "DELETE":
		return "http.MethodDelete"
	case "HEAD":
		return "http.MethodHead"
	case "OPTIONS":
		return "http.MethodOptions"
	case "CONNECT":
		return "http.MethodConnect"
	case "TRACE":
		return "http.MethodTrace"
	default:
		return "unkonwn"
	}

}

func WithRootDir(RootDir string) Option {
	return func(p *Params) {
		if RootDir != "" {
			p.RootDir = RootDir
		}

	}
}

func WithApp(app string) Option {
	return func(p *Params) {
		if app != "" {
			p.App = app
		}
	}
}

func WithTemplateDir(templateDir string) Option {
	return func(p *Params) {
		if templateDir != "" {
			p.TemplateDir = templateDir
		}
	}
}
