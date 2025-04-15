package app

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type httpApplication struct {
	router      *gin.Engine
	middlewares []Middleware
	cfg         *Config
	log         ILogger
}

func newServer(cfg *Config, log ILogger) IRouter {
	app := gin.New()
	app.Use(gin.Recovery())

	return &httpApplication{
		router: app,
		cfg:    cfg,
		log:    log,
	}
}

func (app *httpApplication) Get(path string, handler HandleFunc, middlewares ...Middleware) {
	// app.router.GET(path, func(c *gin.Context) {
	// 	preHandle(handler, preMiddleware(app.middlewares, middlewares)...)(newMuxContext(c, &app.cfg.KafkaConfig, app.log))
	// })
	app.router.GET(path, app.wrapHandler(handler, middlewares...))
}

func (app *httpApplication) wrapHandler(handler HandleFunc, middlewares ...Middleware) gin.HandlerFunc {
	return func(c *gin.Context) {
		preHandle(handler, preMiddleware(app.middlewares, middlewares)...)(newMuxContext(c, &app.cfg.KafkaConfig, app.log))
	}
}

func (app *httpApplication) Post(path string, handler HandleFunc, middlewares ...Middleware) {
	app.router.POST(path, app.wrapHandler(handler, middlewares...))
}

func (app *httpApplication) Put(path string, handler HandleFunc, middlewares ...Middleware) {
	app.router.PUT(path, app.wrapHandler(handler, middlewares...))
}

func (app *httpApplication) Delete(path string, handler HandleFunc, middlewares ...Middleware) {
	app.router.DELETE(path, app.wrapHandler(handler, middlewares...))
}

func (app *httpApplication) Patch(path string, handler HandleFunc, middlewares ...Middleware) {
	app.router.PATCH(path, app.wrapHandler(handler, middlewares...))
}

func (app *httpApplication) Use(middlewares ...Middleware) {
	app.middlewares = append(app.middlewares, middlewares...)
}

func (app *httpApplication) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	app.router.ServeHTTP(w, r)
}

func (app *httpApplication) Register() *http.Server {
	srv := &http.Server{
		Addr:    ":" + app.cfg.AppConfig.Port,
		Handler: app.router,
	}

	return srv
}
