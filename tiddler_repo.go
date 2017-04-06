package main

import "fmt"

type TiddlerRepo interface {
	List() []Tiddler
	Get(key string) (Tiddler, bool)
	Put(tiddler Tiddler) error
	Remove(key string) error
}

type inMemRepo struct {
	tiddlers map[string]Tiddler
}

func InMemory() TiddlerRepo {
	return inMemRepo{tiddlers: make(map[string]Tiddler)}
}

func (repo inMemRepo) List() []Tiddler {
	var list = make([]Tiddler, len(repo.tiddlers))
	i := 0
	for _, tiddler := range repo.tiddlers {
		list[i] = tiddler
		i++
	}

	return list
}

func (repo inMemRepo) Get(key string) (tiddler Tiddler, ok bool) {
	tiddler, ok = repo.tiddlers[key]
	return
}

func (repo inMemRepo) Put(tiddler Tiddler) error {
	repo.tiddlers[tiddler.Title] = tiddler

	return nil
}

func (repo inMemRepo) Remove(key string) error {
	_, ok := repo.tiddlers[key]

	if !ok {
		return fmt.Errorf("Key %v does not exist in repo", key)
	}

	delete(repo.tiddlers, key)
	return nil
}

type Tiddler struct {
	Title string
	Meta  []byte
	Text  string
}
