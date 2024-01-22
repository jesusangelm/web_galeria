package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"jesusmarin.dev/galeria/internal/data"
)

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
