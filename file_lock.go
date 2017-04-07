package main

import (
	"errors"
	"os"

	"log"

	"github.com/fsnotify/fsnotify"
)

type Lock struct {
	file *os.File
}

var ErrLockTaken = errors.New("Lock taken")

func Acquire(name string) (*Lock, error) {
	if _, err := os.Stat(name); err == nil {
		return nil, ErrLockTaken
	}
	if lockfile, err := os.Create(name); err != nil {
		return nil, err
	} else {
		return &Lock{file: lockfile}, nil
	}
}

func (lock *Lock) Release() error {
	if err := lock.file.Close(); err != nil {
		return err
	}

	if err := os.Remove(lock.file.Name()); err != nil {
		return err
	}

	return nil
}

type FolderWatch struct {
	watcher *fsnotify.Watcher
	Events  chan string
}

func NewFolderWatch() (*FolderWatch, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	f := &FolderWatch{
		watcher: w,
		Events:  make(chan string),
	}

	f.start()

	return f, nil
}

func (folderWatch *FolderWatch) start() {
	go func() {
		for {
			select {
			case event := <-folderWatch.watcher.Events:
				if event.Op&fsnotify.Create == fsnotify.Create {
					folderWatch.Events <- event.Name
				}
			case err := <-folderWatch.watcher.Errors:
				if err != nil {
					log.Println("Folderwatch error:", err)
				}
			}
		}
	}()

	folderWatch.watcher.Add(".")
}

func (folderWatch *FolderWatch) Close() error {
	return folderWatch.watcher.Close()
}
