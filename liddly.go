package main

import (
	"log"
	"net/http"

	"context"
	"os"
	"os/signal"
	"time"

	"github.com/roelrymenants/liddly/repo"
	"github.com/roelrymenants/liddly/tiddlyweb"
)

const lockfile = "./liddly.lock"
const watchfile = "./liddly.shutdown"

var repository repo.TiddlerRepo
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

	shutdownOnCreate(folderWatch, watchfile, asyncShutdown)

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		asyncShutdown()
	}()

	repository = repo.NewSqlite("./tiddlers.db")
	tiddlyweb.Register(repository)

	log.Println(srv.ListenAndServe())
}

func asyncShutdown() {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

	if err := srv.Shutdown(ctx); err != nil {
		log.Panic(err)
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
