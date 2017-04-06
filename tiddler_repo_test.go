package main

import "testing"

func TestInMemRepo(t *testing.T) {
	var repo = InMemory()

	if len(repo.List()) != 0 {
		t.Error("Not empty repo on init")
	}

	key := "x"
	tiddler := Tiddler{Title: key}

	rev, err := repo.Put(tiddler)

	if err != nil || rev != 0 {
		t.Error("Error putting tiddler in repo")
	}

	if _, ok := repo.Get(key); !ok {
		t.Error("Put and get are not symmetric")
	}

	if len(repo.List()) != 1 {
		t.Error("List has more tiddlers than added: %v", len(repo.List()))
	}
}
