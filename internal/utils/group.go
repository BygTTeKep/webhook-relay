package utils

import (
	"fmt"
	"net/http"
	"strings"
)

type Middleware func(http.Handler) http.Handler


func MW(name string) Middleware {
	return  func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			fmt.Println("→", name)
			next.ServeHTTP(w, req)
			fmt.Println("←", name)
		})
	}
}

type Group struct {
	Mux *http.ServeMux
	prefix string
	mws []Middleware
}

func (g * Group) Group(prefix string, mws ...Middleware) *Group {
	return &Group{
		Mux: g.Mux,
		prefix: prefix,
		mws: mws,
	}
}

func (g *Group)Handle(pattern string, h http.HandlerFunc) {
	method, path, ok := strings.Cut(pattern, " ");
	if (!ok) {
		method, path = "", pattern
	}
	full := g.prefix + path
	if method != "" {
		full = method + " " + full
	}
	var handler http.Handler = h;
	for i := len(g.mws) -1; i > 0; i-- {
		handler = g.mws[i](handler)
	}
	g.Mux.Handle(full, handler)
}