# liddly
A local tiddlywiki server

## Building
Prerequisites:
golang v1.8

```
$ go get github.com/roelrymenants/liddly
```
## Usage
Prerequisites:
An empty tiddlywiki (with at least tiddlyweb/tiddlyspace plugin installed) called `index.html` in the working directory.

See also http://tiddlywiki.com/

Running
```
$ liddly
```
will start an http server on port 8080. Accessing it with a browser will serve tiddlywiki.
Tiddlers are saved in a sqlite db (`tiddlers.db`) in the working directory

To allow running from dropbox, a lock file (`tiddlers.lock`) is placed in the working directory.
This avoid multiple instances (possibly on different machines) accessing the database.

In order to allow a preemptive workflow, any new instance will create `tiddlers.shutdown` and exit.
Running instances will react by properly shutting down (removing lock file and shutdown file).
After the lock file and shutdown file have disappeared, you can start the local instance again.
