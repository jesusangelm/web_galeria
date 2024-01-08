package main

import (
	"errors"
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
