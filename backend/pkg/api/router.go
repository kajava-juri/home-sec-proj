package api

import (
	"net/http"
)

type Router struct {
	routes map[string]map[string]http.HandlerFunc
}

func NewRouter() *Router {
	router := &Router{
		routes: make(map[string]map[string]http.HandlerFunc),
	}

	return router
}

func (r *Router) RegisterRoute(method, path string, handler http.HandlerFunc) {
	if r.routes[path] == nil {
		r.routes[path] = make(map[string]http.HandlerFunc)
	}
	r.routes[path][method] = handler
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if handlers, ok := r.routes[req.URL.Path]; ok {
		if handler, methodExists := handlers[req.Method]; methodExists {
			handler(w, req)
			return
		}
	}
	http.NotFound(w, req)
}
