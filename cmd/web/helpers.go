package main

import (
	"bytes"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"
)

// Helper that write an error message and stack trace to the errorLog.
// This also send a generic 500 Internal server error response to the user
func (app *application) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.logger.PrintError(fmt.Errorf("%s", trace), nil)

	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// The clientError helper sends a specific status code and corresponding description
// to the user.
func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

// Simple wrapper over clientError for sending 404 error response to user
func (app *application) notFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)
}

// Create an newTemplateData() helper, which returns a pointer to a templateData
// struct initialized with the current year.
func (app *application) newTemplateData(r *http.Request) *templateData {
	return &templateData{
		CurrentYear: time.Now().Year(),
	}
}

func (app *application) render(w http.ResponseWriter, status int, page string, data *templateData) {
	// Retrieve the appropriate template set from the cache based on the page
	// name (like 'home.tmpl'). If no entry exists in the cache with the
	// provided name, then create a new error an
	templateSet, ok := app.templateCache[page]
	if !ok {
		err := fmt.Errorf("the template %s does not exist", page)
		app.serverError(w, err)
		return
	}

	// initialize a new buffer
	buf := new(bytes.Buffer)

	// Execute the template set and write the response body. Again, if there
	// is any error we call the the serverError() helper.
	err := templateSet.ExecuteTemplate(buf, "base", data)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Write out the provided HTTP status code ('200 OK', '400 Bad Request', etc).
	w.WriteHeader(status)

	buf.WriteTo(w)
}
