package main

import (
	"log"
	"net/http"

	"context"
	"os"
	"os/signal"
	"time"

	"flag"

	"github.com/roelrymenants/liddly/repo"
	"github.com/roelrymenants/liddly/tiddlyweb"
)

const lockfile = "./liddly.lock"
const watchfile = "./liddly.shutdown"
const dbfile = "./tiddlers.db"

var repository repo.TiddlerRepo
var srv http.Server

func main() {
	var address = flag.String("bind", ":8080", "The ip:port to listen on")
	var preemptive = flag.Bool("preemptive", true, "Whether to create a shutdown file when already locked")
	flag.Parse()

	srv = http.Server{
		Addr: *address,
	}

	lock, err := Acquire(lockfile)
	if err != nil {
		if *preemptive {
			os.Create(watchfile)
			log.Println("Lock file exists. Initialized remote shutdown.")
		} else {
			log.Println("Lock file exists. Did not initialize remote shutdown. Exit.")
		}
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

	repository = repo.NewSqlite(dbfile)
	tiddlyweb.Register(repository)

	log.Println("Listening for connections on", *address)
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
