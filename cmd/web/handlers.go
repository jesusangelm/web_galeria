package main

import (
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"jesusmarin.dev/galeria/internal/data"
)

// home handler for root route
func (app *application) home(w http.ResponseWriter, r *http.Request) {
	categories, err := app.models.Categories.List()
	if err != nil {
		app.serverError(w, err)
	}

	templatedata := app.newTemplateData(r)
	templatedata.Categories = categories

	app.render(w, http.StatusOK, "home.tmpl", templatedata)
}

func (app *application) categoryShow(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}

	category, err := app.models.Categories.Get(int64(id))
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

	data := app.newTemplateData(r)
	data.Category = category

	app.render(w, http.StatusOK, "category_view.tmpl", data)
}

func (app *application) itemShow(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}

	item, err := app.models.Items.Get(int64(id))
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

	data := app.newTemplateData(r)
	data.Item = item

	app.render(w, http.StatusOK, "item_view.tmpl", data)
}

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
