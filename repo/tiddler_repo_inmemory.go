package repo

import "fmt"

type inMemRepo struct {
	tiddlers map[string]Tiddler
}

func NewInMemory() TiddlerRepo {
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

func (repo inMemRepo) Put(tiddler Tiddler) (int, error) {
	var rev int

	if prev, ok := repo.Get(tiddler.Title); ok {
		rev = prev.Revision
		rev++
	}
	tiddler.Revision = rev

	repo.tiddlers[tiddler.Title] = tiddler

	return rev, nil
}

func (repo inMemRepo) Remove(key string) error {
	_, ok := repo.tiddlers[key]

	if !ok {
		return fmt.Errorf("Key %v does not exist in repo", key)
	}

	delete(repo.tiddlers, key)
	return nil
}
