package main

import (
	"fmt"
	"log"
	"net/http"

	"context"
	"os"
	"os/signal"
	"time"
)

const lockfile = "./liddly.lock"
const watchfile = "./liddly.shutdown"

var repo TiddlerRepo
var srv = http.Server{
	Addr: ":8080",
}

func main() {
	lock, err := Acquire(lockfile)
	if err != nil {
		os.Create(watchfile)
		log.Println("Lock file exists. Initialized remote shutdown.")
		return
	}
	defer lock.Release()

	folderWatch, err := NewFolderWatch()
	if err != nil {
		log.Println("Could not start watch on current dir")
		return
	}
	defer folderWatch.Close()

	register("/", strictPath(allowOnly(index, "GET", "OPTIONS")))
	register("/status", strictPath(allowOnly(status, "GET")))
	register("/recipes/all/tiddlers.json", strictPath(allowOnly(list, "GET")))
	register("/recipes/all/tiddlers/", allowOnly(detail, "GET", "PUT"))
	register("/bags/bag/tiddlers/", allowOnly(remove, "DELETE"))

	shutdownOnCreate(folderWatch, watchfile, asyncShutdown)

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		asyncShutdown()
	}()

	repo = NewSqliteRepo("./tiddlers.db")

	log.Println(srv.ListenAndServe())
}

func asyncShutdown() {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

	if err := srv.Shutdown(ctx); err != nil {
		log.Panic(err)
	}
}

func register(pattern string, handler func(string) http.HandlerFunc) {
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

func shutdownOnCreate(folderWatch *FolderWatch, createdFile string, shutdownCallback func()) {
	go func() {
		for {
			if e := <-folderWatch.Events; e != createdFile {
				continue
			} else {
				defer os.Remove(createdFile)

				shutdownCallback()
				return
			}
		}
	}()
}

func jsonResponse(w http.ResponseWriter) http.ResponseWriter {
	w.Header().Set("Content-Type", "application/json")

	return w
}
