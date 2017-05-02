//go:generate statik -src=./web

package main

import (
	"log"
	"net/http"

	"context"
	"os"
	"os/signal"
	"time"

	"flag"
	"fmt"

	"os/exec"
	"runtime"

	"strings"

	"bytes"

	"github.com/roelrymenants/liddly/repo"
	"github.com/roelrymenants/liddly/tiddlyweb"

	"github.com/rakyll/statik/fs"
	_ "github.com/roelrymenants/liddly/statik"
)

const lockfile = "./liddly.lock"
const watchfile = "./liddly.shutdown"
const dbfile = "./tiddlers.db"
const version = "v0.2"

var repository repo.TiddlerRepo
var srv http.Server

func main() {
	var address = flag.String("bind", ":8080", "The ip:port to listen on")
	var preemptive = flag.Bool("preemptive", true, "Whether to create a shutdown file when already locked")
	var showVersion = flag.Bool("version", false, "Display version string")
	var openBrowser = flag.Bool("browser", true, "Open a browser on start-up")

	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		return
	}

	srv = http.Server{
		Addr: *address,
	}

	lock, err := Acquire(lockfile)
	if err != nil {
		exit(preemptive)
		return
	}
	defer lock.Release()

	folderWatch, err := NewFolderWatch()
	if err != nil {
		log.Println("Could not start watch on current dir", err)
		return
	}
	defer folderWatch.Close()

	shutdownOnCreate(folderWatch, watchfile, asyncShutdown)
	shutdownOnSignal(asyncShutdown)

	statikFS, err := fs.New()
	if err != nil {
		log.Println(err)
	}

	repository = repo.NewSqlite(dbfile)
	tiddlyweb.RegisterHandlers(repository, statikFS)

	log.Println("Listening for connections on", *address)

	var done = make(chan struct{})

	go func(chan struct{}) {
		log.Println(srv.ListenAndServe())
		done <- struct{}{}
	}(done)

	if *openBrowser {
		url := composeUrl(*address)

		open(url)
	}

	<-done
}
func composeUrl(bindAddress string) string {
	var buffer bytes.Buffer

	buffer.WriteString("http://")

	if strings.HasPrefix(bindAddress, ":") {
		buffer.WriteString("localhost")
	}

	buffer.WriteString(bindAddress)

	return buffer.String()
}

func exit(preemptive *bool) {
	if *preemptive {
		file, err := os.Create(watchfile)
		if err != nil {
			log.Println("Error creating shutdown file", err)
			return
		}

		file.Close()
		log.Println("Lock file exists. Initialized remote shutdown.")
	} else {
		log.Println("Lock file exists. Did not initialize remote shutdown. Exit.")
	}
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

func shutdownOnSignal(shutdownCallback func()) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		shutdownCallback()
	}()
}

func open(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}
