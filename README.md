# liddly
A local [tiddlywiki](http://tiddlywiki.com) server with a sqlite backend

## Building
Prerequisites:
* golang v1.8.1
* github.com/rakyll/static
* Everything necessary to compile github.com/mattn/go-sqlite3

```
$ go get github.com/rakyll/statik
$ go get github.com/roelrymenants/liddly
$ cd $GOPATH/src/github.com/roelrymenants/liddly
$ go generate
$ go install github.com/roelrymenants/liddly
```

## Usage
Running `liddly` will start an http server on port 8080. The default browser will open automatically
Tiddlers are saved in a sqlite db (`tiddlers.db`) in the working directory

To allow running from dropbox, a lock file (`tiddlers.lock`) is placed in the working directory.
This avoid multiple instances (possibly on different machines) accessing the database.

In order to allow a preemptive workflow, any new instance will create `tiddlers.shutdown` and exit.
Running instances will react by properly shutting down (removing lock file and shutdown file).
After the lock file and shutdown file have disappeared, you can start the local instance again.

See also `liddly --help` for more info
