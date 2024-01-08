package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"jesusmarin.dev/galeria/ui"
)

// Routes, fileserver and middleware setup
func (app *application) routes() http.Handler {
	// Initiate a httproute router
	router := httprouter.New()

	// Define a fileserver. This use embed files in the binary app
	fileServer := http.FileServer(http.FS(ui.Files))
	// define the path for the fileserver
	router.Handler(http.MethodGet, "/static/*filepath", fileServer)

	dynamic := alice.New(app.sessionManager.LoadAndSave)

	// Routes definition
	router.Handler(http.MethodGet, "/", dynamic.ThenFunc(app.home))
	router.Handler(http.MethodGet, "/category/:id", dynamic.ThenFunc(app.categoryShow))
	router.Handler(http.MethodGet, "/item/:id", dynamic.ThenFunc(app.itemShow))

	// Standard middleware managed by alice with some custom middlewares
	// standard := alice.New(app.recoverPanic, app.logRequest, secureHeaders)
	standard := alice.New(app.recoverPanic, app.secureHeaders)

	return standard.Then(router)
}
