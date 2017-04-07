package tiddlyweb

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

	"github.com/roelrymenants/liddly/repo"
)

var repository repo.TiddlerRepo

func RegisterHandlers(r repo.TiddlerRepo) {
	repository = r

	register("/", strictPath(allowOnly(index, "GET", "OPTIONS")))
	register("/status", strictPath(allowOnly(status, "GET")))
	register("/recipes/all/tiddlers.json", strictPath(allowOnly(list, "GET")))
	register("/recipes/all/tiddlers/", allowOnly(detail, "GET", "PUT"))
	register("/bags/bag/tiddlers/", allowOnly(remove, "DELETE"))
}

func index(basepath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			http.ServeFile(w, r, "index.html")
		case "OPTIONS":
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

func status(basepath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w).Write([]byte(`{"username":"me","space":{"recipe":"all"}}`))
	}
}

func list(basepath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		list := repository.List()

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
}

func detail(basepath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		title := strings.TrimPrefix(r.URL.Path, basepath)

		switch r.Method {
		case "GET":
			tiddler, ok := repository.Get(title)
			if !ok {
				log.Printf("Tiddler not found: '%v'", title)
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
			var tiddler repo.Tiddler

			var js map[string]interface{}
			err := json.NewDecoder(r.Body).Decode(&js)
			if err != nil {
				log.Printf("Error decoding tiddler: '%v'", title)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			io.Copy(ioutil.Discard, r.Body)
			js["bag"] = "bag"

			text, _ := js["text"].(string)
			delete(js, "text")

			meta, err := json.Marshal(js)
			if err != nil {
				log.Printf("Error marshalling tiddler meta: '%v'", js)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			tiddler.Text = text
			tiddler.Title = title
			tiddler.Meta = meta

			//create the tiddler
			rev, err := repository.Put(tiddler)
			if err != nil {
				log.Printf("Error saving tiddler '%v'", title)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			etag := fmt.Sprintf(`"bag/%s/%d:%032x"`, url.QueryEscape(title), rev, md5.Sum(meta))
			w.Header().Set("ETag", etag)
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

func remove(basepath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		title := strings.TrimPrefix(r.URL.Path, basepath)

		if err := repository.Remove(title); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
