package gee

import (
	"html/template"
	"log"
	"net/http"
	"path"
	"strings"
)

// HandlerFunc defines the request handler used by gee
type HandlerFunc func(*Context)

type RouterGroup struct {
	prefix      string
	middlewares []HandlerFunc
	engine      *Engine
}

// Engine implement the interface of ServeHTTP
type Engine struct {
	*RouterGroup                     //包含RouterGroup中的所有方法，可直接引用
	router        *router            // 所有的路由
	groups        []*RouterGroup     // store all groups

	htmlTemplates *template.Template //for html render
	funcMap       template.FuncMap   //for html render
}

// New is the constructor of gee.Engine
func New() *Engine {
	engine := &Engine{router: newRouter()}
	engine.RouterGroup = &RouterGroup{engine: engine}
	engine.groups = []*RouterGroup{engine.RouterGroup}
	return engine
}

func (group *RouterGroup) Group(prefix string) *RouterGroup {
	engine := group.engine
	newGroup := &RouterGroup{
		prefix: group.prefix + prefix,
		engine: engine,
	}
	engine.groups = append(engine.groups, newGroup)
	return newGroup
}

// GET defines the method to add GET request
func (group *RouterGroup) GET(pattern string, handler HandlerFunc) {
	group.addRoute("GET", pattern, handler)
}

// POST defines the method to add POST request
func (group *RouterGroup) POST(pattern string, handler HandlerFunc) {
	group.addRoute("POST", pattern, handler)
}

// Run defines the method to start a http server
func (engine *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, engine)
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var middlewares []HandlerFunc
	for _, group := range engine.groups {
		if strings.HasPrefix(req.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}
	c := newContext(w, req)
	c.handlers = middlewares
	c.engine = engine // 实例化赋值
	engine.router.handle(c)
}

func (group *RouterGroup) addRoute(method string, comp string, handler HandlerFunc) {
	pattern := group.prefix + comp
	log.Printf("Route %4s - %s", method, pattern)
	group.engine.router.addRoute(method, pattern, handler)
}

// 所谓的中间件插口，用户如果自定义了中间件，可以用Use在代码中使用
func (group *RouterGroup) Use(middlewares ...HandlerFunc) {
	group.middlewares = append(group.middlewares, middlewares...)
}

//static handler
func (group *RouterGroup) creatStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
	absolutePath := path.Join(group.prefix, relativePath)
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))
	return func(c *Context) {
		file := c.Param("filepath")
		if _, err := fs.Open(file); err != nil {
			c.Status(http.StatusNotFound)
			return
		}

		fileServer.ServeHTTP(c.Writer, c.Req)
	}
}

// server static files
func (group *RouterGroup) Static(relativePath string, root string) {
	handler := group.creatStaticHandler(relativePath, http.Dir(root))
	urlPattern := path.Join(relativePath, "/*filepath")
	// Register GET handlers
	group.GET(urlPattern, handler)
}

// html
func (engine *Engine) SetFuncMap(funcMap template.FuncMap){
	engine.funcMap = funcMap
}

func (engine *Engine) LoadHTMLGlob(pattern string){
	engine.htmlTemplates = template.Must(template.New("").Funcs(engine.funcMap).ParseGlob(pattern))
}

