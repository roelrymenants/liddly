package main

import "testing"

func TestInMemRepo(t *testing.T) {
	var repo = InMemory()

	if len(repo.List()) != 0 {
		t.Error("Not empty repo on init")
	}

	key := "x"
	tiddler := Tiddler{Title: key}

	err := repo.Put(tiddler)

	if err != nil {
		t.Error("Error putting tiddler in repo")
	}

	if get, ok := repo.Get(key); !ok || get != tiddler {
		t.Error("Put and get are not symmetric")
	}

	if len(repo.List()) != 1 {
		t.Error("List has more tiddlers than added: %v", len(repo.List()))
	}
}
