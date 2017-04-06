package main

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

var repo = InMemory()

func main() {
	fmt.Println("Ghello")

	http.HandleFunc("/", index)
	http.HandleFunc("/status", status)
	http.HandleFunc("/recipes/all/tiddlers.json", list)
	http.HandleFunc("/recipes/all/tiddlers/", detail)
	http.HandleFunc("/bags/bag/tiddlers/", remove)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, fmt.Sprintf("Path not mapped: %v", r.URL.Path), http.StatusNotFound)
		return
	}

	http.ServeFile(w, r, "index.html")
}

func status(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	jsonResponse(w).Write([]byte(`{"username":"me","space":{"recipe":"all"}}`))
}

func list(w http.ResponseWriter, r *http.Request) {
	list := repo.List()

	var buff bytes.Buffer
	buff.WriteString("[")

	for i, tiddler := range list {
		if i != 0 {
			buff.WriteString(",")
		}
		buff.Write(tiddler.Meta)

	}

	buff.WriteString("]")

	jsonResponse(w).Write(buff.Bytes())
}

func detail(w http.ResponseWriter, r *http.Request) {
	title := strings.TrimPrefix(r.URL.Path, "/recipes/all/tiddlers/")

	switch r.Method {
	case "GET":
		tiddler, ok := repo.Get(title)
		if !ok {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		var js map[string]interface{}
		err := json.Unmarshal(tiddler.Meta, &js)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if tiddler.Text != "" {
			js["text"] = tiddler.Text
		}

		json.NewEncoder(w).Encode(js)
	case "PUT":
		var tiddler Tiddler

		var js map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&js)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		io.Copy(ioutil.Discard, r.Body)
		js["bag"] = "bag"

		text, _ := js["text"].(string)
		delete(js, "text")

		meta, err := json.Marshal(js)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		tiddler.Text = text
		tiddler.Title = title
		tiddler.Meta = meta

		//create the tiddler
		err = repo.Put(tiddler)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		etag := fmt.Sprintf(`"bag/%s/%d:%032x"`, url.QueryEscape(title), 0, md5.Sum(meta))
		w.Header().Set("ETag", etag)
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func remove(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	title := strings.TrimPrefix(r.URL.Path, "/bags/bag/tiddlers/")

	if err := repo.Remove(title); err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func jsonResponse(w http.ResponseWriter) http.ResponseWriter {
	w.Header().Set("Content-Type", "application/json")

	return w
}
