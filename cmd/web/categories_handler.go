package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"jesusmarin.dev/galeria/internal/data"
	"jesusmarin.dev/galeria/internal/validator"
)

// home handler for root route
func (app *application) listCategories(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name string
		data.Filters
	}

	v := validator.New()
	qs := r.URL.Query() // To get filter parameters from the QueryString

	input.Name = app.readString(qs, "name", "")
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 6, v)
	input.Filters.Sort = app.readString(qs, "sort", "-created_at")
	input.Filters.SortSafeList = []string{
		"id", "name", "created_at", "-id", "-name", "-created_at",
	}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.notFound(w)
		return
	}

	categories, metadata, err := app.models.Categories.List(input.Name, input.Filters)
	if err != nil {
		app.serverError(w, err)
	}

	templatedata := app.newTemplateData(r)
	templatedata.Categories = categories
	templatedata.Metadata = metadata

	app.render(w, http.StatusOK, "home.tmpl", templatedata)
}

func (app *application) categoryShow(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name string
		data.Filters
	}

	v := validator.New()
	qs := r.URL.Query() // To get filter parameters from the QueryString

	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 6, v)
	input.Filters.Sort = app.readString(qs, "sort", "-created_at")
	input.Filters.SortSafeList = []string{
		"id", "name", "created_at", "-id", "-name", "-created_at",
	}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.notFound(w)
		return
	}

	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}

	category, metadata, err := app.models.Categories.Get(int64(id), input.Filters)
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
	data.Metadata = metadata

	app.render(w, http.StatusOK, "category_view.tmpl", data)
}
