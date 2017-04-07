package repo

type TiddlerRepo interface {
	List() []Tiddler
	Get(key string) (Tiddler, bool)
	Put(tiddler Tiddler) (int, error)
	Remove(key string) error
}

type Tiddler struct {
	Title    string
	Meta     []byte
	Text     string
	Revision int
}
