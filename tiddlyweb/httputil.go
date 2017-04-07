package tiddlyweb

import (
	"fmt"
	"log"
	"net/http"
)

func register( pattern string, handler func(string) http.HandlerFunc) {
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		log.Println(r)

		handler(pattern)(w, r)
	})
}
func strictPath(handler func(string) http.HandlerFunc) func(string) http.HandlerFunc {
	return func(pattern string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != pattern {
				log.Println("Request not allowed for strict path", r.URL.Path)
				http.Error(w, fmt.Sprintf("Path not mapped: %v", r.URL.Path), http.StatusNotFound)
				return
			}

			handler(pattern)(w, r)
		}
	}
}

func allowOnly(handler func(string) http.HandlerFunc, methods ...string) func(string) http.HandlerFunc {
	return func(pattern string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			var allowed = false

			for _, method := range methods {
				if r.Method == method {
					allowed = true
				}
			}
			if !allowed {
				log.Printf("Method '%v' not allowed for path '%v'", r.Method, r.URL.Path)
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}

			handler(pattern)(w, r)
		}
	}
}

func jsonResponse(w http.ResponseWriter) http.ResponseWriter {
	w.Header().Set("Content-Type", "application/json")

	return w
}
