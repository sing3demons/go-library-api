package app

import (
	"context"
	"net/http"
	"strings"
)

func removeBraces(str string) string {
	return strings.ReplaceAll(strings.ReplaceAll(str, "{", ""), "}", "")
}

type ContextKey string

func setParam(path string, r *http.Request) *http.Request {
	subPath := strings.Split(path, "/")
	sss := strings.Split(r.URL.Path, "/")

	// Remove empty strings from the slice
	if len(subPath) == len(sss) {
		for i := 0; i < len(subPath); i++ {
			var k, v string
			if subPath[i] != "" {
				k = subPath[i]
			}
			if sss[i] != "" {
				v = sss[i]
			}

			if k != "" && v != "" && k != v {
				key := removeBraces(k)
				if key != "" {
					ctx := context.WithValue(r.Context(), ContextKey(key), v)
					r = r.WithContext(ctx)
				}

			}
		}
	}
	return r
}

func preHandle(final HandleFunc, middlewares ...Middleware) HandleFunc {
	if final == nil {
		panic("no final handler")
		// Or return a default handler.
	}
	// Execute the middleware in the same order and return the final func.
	// This is a confusing and tricky construct :)
	// We need to use the reverse order since we are chaining inwards.
	for i := len(middlewares) - 1; i >= 0; i-- {
		final = middlewares[i](final) // mw1(mw2(mw3(final)))
	}
	return final
}

func preMiddleware(app []Middleware, middlewares []Middleware) []Middleware {
	var m []Middleware
	if len(app) > 0 {
		m = append(m, app...)
	}

	if len(middlewares) > 0 {
		m = append(m, middlewares...)
	}
	return m
}
