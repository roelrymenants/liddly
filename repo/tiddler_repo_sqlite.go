package repo

import (
	"database/sql"
	"log"
	"os"
	"path"

	_ "github.com/mattn/go-sqlite3"
)

type sqliteRepo struct {
	Db *sql.DB
}

func DefaultDbFile() string {
	ex, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	exPath := path.Dir(ex)

	return path.Join(exPath, "tiddlers.db")
}

func NewSqlite(dbfile string) TiddlerRepo {
	db, err := sql.Open("sqlite3", dbfile)
	if err != nil {
		log.Fatal(err)
	}

	tableCreate :=
		`CREATE TABLE IF NOT EXISTS tiddlers ( 
    title TEXT, 
	meta BLOB, 
	text TEXT, 
	revision INTEGER,
	PRIMARY KEY(title, revision));`

	_, err = db.Exec(tableCreate)
	if err != nil {
		log.Fatal(err)
	}

	return sqliteRepo{db}
}

func (repo sqliteRepo) List() []Tiddler {
	rows, err := repo.Db.Query(`SELECT title, meta, text, revision 
FROM tiddlers`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var tiddlers []Tiddler

	for rows.Next() {
		var tiddler Tiddler
		err = rows.Scan(&tiddler.Title, &tiddler.Meta, &tiddler.Text, &tiddler.Revision)
		if err != nil {
			log.Fatal(err)
		}
		tiddlers = append(tiddlers, tiddler)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	return tiddlers
}

func (repo sqliteRepo) Get(key string) (tiddler Tiddler, ok bool) {
	stmt, err := repo.Db.Prepare(`SELECT title, meta, text, revision 
FROM tiddlers 
WHERE title=?
ORDER BY revision DESC
LIMIT 1`)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	err = stmt.QueryRow(key).Scan(&tiddler.Title, &tiddler.Meta, &tiddler.Text, &tiddler.Revision)
	if err != nil {
		if err == sql.ErrNoRows {
			return tiddler, false
		}
		log.Fatal(err)
	}
	return tiddler, true
}

func (repo sqliteRepo) Put(tiddler Tiddler) (int, error) {
	var rev int

	if prev, ok := repo.Get(tiddler.Title); ok {
		rev = prev.Revision
		rev++
	}
	tiddler.Revision = rev

	tx, err := repo.Db.Begin()
	if err != nil {
		return 0, err
	}
	stmt, err := tx.Prepare(`INSERT INTO tiddlers(title, meta, text, revision) 
VALUES(?, ?, ?, ?)`)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()
	_, err = stmt.Exec(tiddler.Title, tiddler.Meta, tiddler.Text, tiddler.Revision)
	if err != nil {
		return 0, err
	}
	tx.Commit()
	return rev, nil
}

func (repo sqliteRepo) Remove(key string) error {
	tx, err := repo.Db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(`DELETE FROM tiddlers WHERE title=?`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(key)
	if err != nil {
		return err
	}
	tx.Commit()
	return nil
}
