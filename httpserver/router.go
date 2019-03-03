package httpserver

import "github.com/gin-gonic/gin"

type HandlerFunc func(ctx *Context)

type Routes []Route

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc func(ctx *Context)
}

type Router struct {
	*gin.Engine
	httpService *Server
	routes      Routes
}

// IHttpRoutes represents a HTTP route
type IHttpRoutes interface {
	gin.IRoutes
}

// NewRouter creates a HTTP router
func NewRouter(routes Routes) *Router {
	return &Router{Engine: gin.Default(), routes: routes}
}

// Use attaches a global middleware to the router.
func (router *Router) Use(handlers ...HandlerFunc) IHttpRoutes {
	return router.Engine.Use(router.castHandlersToGinHandlers(handlers)...)
}

// Handle registers a new request handle and middleware with the given path and method.
func (router *Router) Handle(httpMethod, relativePath string, handlers ...HandlerFunc) IHttpRoutes {
	return router.Engine.Handle(httpMethod, relativePath, router.castHandlersToGinHandlers(handlers)...)
}

// POST is a shortcut for router.Handle("POST", path, handle)
func (router *Router) POST(relativePath string, handlers ...HandlerFunc) IHttpRoutes {
	return router.Engine.POST(relativePath, router.castHandlersToGinHandlers(handlers)...)
}

// GET is a shortcut for router.Handle("GET", path, handle)
func (router *Router) GET(relativePath string, handlers ...HandlerFunc) IHttpRoutes {
	return router.Engine.GET(relativePath, router.castHandlersToGinHandlers(handlers)...)
}

// DELETE is a shortcut for router.Handle("DELETE", path, handle)
func (router *Router) DELETE(relativePath string, handlers ...HandlerFunc) IHttpRoutes {
	return router.Engine.DELETE(relativePath, router.castHandlersToGinHandlers(handlers)...)
}

// PATCH is a shortcut for router.Handle("PATCH", path, handle)
func (router *Router) PATCH(relativePath string, handlers ...HandlerFunc) IHttpRoutes {
	return router.Engine.PATCH(relativePath, router.castHandlersToGinHandlers(handlers)...)
}

// PUT is a shortcut for router.Handle("PUT", path, handle)
func (router *Router) PUT(relativePath string, handlers ...HandlerFunc) IHttpRoutes {
	return router.Engine.PUT(relativePath, router.castHandlersToGinHandlers(handlers)...)
}

// OPTIONS is a shortcut for router.Handle("OPTIONS", path, handle)
func (router *Router) OPTIONS(relativePath string, handlers ...HandlerFunc) IHttpRoutes {
	return router.Engine.OPTIONS(relativePath, router.castHandlersToGinHandlers(handlers)...)
}

// HEAD is a shortcut for router.Handle("HEAD", path, handle)
func (router *Router) HEAD(relativePath string, handlers ...HandlerFunc) IHttpRoutes {
	return router.Engine.HEAD(relativePath, router.castHandlersToGinHandlers(handlers)...)
}

func (router *Router) castHandlersToGinHandlers(handlers []HandlerFunc) []gin.HandlerFunc {
	ginHandlers := make([]gin.HandlerFunc, len(handlers))
	for i, handler := range handlers {
		ginHandlers[i] = func(c *gin.Context) {
			handler(&Context{
				Context: c,
			})
		}
	}
	return ginHandlers
}
