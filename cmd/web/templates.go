package main

import (
	"html/template"
	"io/fs"
	"path/filepath"
	"time"

	"jesusmarin.dev/galeria/internal/data"
	"jesusmarin.dev/galeria/ui"
)

// Define a templateData type to act as the holding structure for
// any dynamic data that we want to pass to our HTML templates.
type templateData struct {
	CurrentYear int
	Flash       string
	Category    *data.Category
	Categories  []*data.Category
	Item        *data.Item
	Items       []*data.Item
}

// Create a humanDate function which returns a nicely formatted string
// representation of a time.Time object.
func humanDate(t time.Time) string {
	return t.Format("02 jan 2006 at 15:04")
}

// Initialize a template.FuncMap object and store it in a global variable. This is
// essentially a string-keyed map which acts as a lookup between the names of our
// custom template functions and the functions themselves.
var functions = template.FuncMap{
	"humanDate": humanDate,
}

// function for read and copy all .tmpl templates in a map variable stored in memory.
// this will improve the read and generation of html from the templates
func newTemplateCache() (map[string]*template.Template, error) {
	// Initialize a new map to act as the cache.
	cache := map[string]*template.Template{}

	// Use fs.Glob() to get a slice of all filepaths in the ui.Files embedded
	// filesystem which match the pattern 'html/pages/*.tmpl'. This essentially
	// gives us a slice of all the 'page' templates for the application, just
	// like before.
	pages, err := fs.Glob(ui.Files, "html/pages/*.tmpl")
	if err != nil {
		return nil, err
	}

	// Loop through the page filepaths one-by-one.
	for _, page := range pages {
		// Extract the file name (like 'home.tmpl') from the full filepath
		// and assign it to the name variable.
		name := filepath.Base(page)

		// Create a slice containing the filepath patterns for the templates we
		// want to parse.
		patterns := []string{
			"html/base.tmpl",
			"html/partials/*.tmpl",
			page,
		}

		// Call ParseFiles() *on this template set* to add the page template.
		templateSet, err := template.New(name).Funcs(functions).ParseFS(ui.Files, patterns...)
		if err != nil {
			return nil, err
		}

		// Add the template set to the map, using the name of the page
		// (like 'home.tmpl') as the key.
		cache[name] = templateSet
	}

	return cache, nil
}
