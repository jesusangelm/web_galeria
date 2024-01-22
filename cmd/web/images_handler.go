package main

import (
	"io"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) images(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())

	key := params.ByName("key")
	if key == "" {
		app.notFound(w)
		return
	}

	imageUrl := app.s3Manager.GetFileUrl(key)

	resp, err := http.Get(imageUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		app.notFound(w)
		return
	}

	contentType := resp.Header.Get("Content-Type")

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "max-age=3155695200, public")
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Fatal(err)
		app.serverError(w, err)
	}
}
