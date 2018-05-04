package main

import (
	"fmt"
	"net/http"
	"bytes"
	"html/template"
	"github.com/go-chi/chi"
	"github.com/jzelinskie/geddit"
)

func main() {

	r := chi.NewRouter()

	r.Get("/", HomeHandler)
	r.Get("/login", LoginHandler)
	r.Get("/login/callback", LoginCallbackHandler)

	err := http.ListenAndServe(":8087", r)
	if err != nil {
		fmt.Println(err)
	}
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {

	client, err := getClient()
	if err != nil {
		fmt.Println(err)
	}

	items, err := client.SubredditSubmissions("all", geddit.PopularitySort(geddit.HotSubmissions), geddit.ListingOptions{})
	if err != nil {
		fmt.Println(err)
	}

	t := homeTemplate{}
	t.Items = items

	returnTemplate(w, r, "home", t)
}

type homeTemplate struct {
	Items []*geddit.Submission
}

func returnTemplate(w http.ResponseWriter, r *http.Request, page string, pageData interface{}) (err error) {

	w.Header().Set("Content-Type", "text/html")

	// Load templates needed
	t, err := template.New("t").ParseFiles(page + ".html")
	if err != nil {
		fmt.Println(err)
		return err
	}

	// Write a respone
	buf := &bytes.Buffer{}
	err = t.ExecuteTemplate(buf, page, pageData)
	if err != nil {
		fmt.Println(err)
		return err
	} else {
		// No error, send the content, HTTP 200 response status implied
		buf.WriteTo(w)
	}

	return nil
}
