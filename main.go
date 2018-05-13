package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"golang.org/x/oauth2"
)

var scopes = []AuthScope{ScopeIdentity, ScopeRead, ScopeHistory, ScopeSubscribe}

const (
	userAgent = "Reddit Filters"
)

var client = GetClient(
	os.Getenv("REDDIT_CLIENT"),
	os.Getenv("REDDIT_SECRET"),
	"http://localhost:8087/login/callback",
	userAgent,
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

func LoginHandler(w http.ResponseWriter, r *http.Request) {

	url, state := client.Login(scopes, false, "")

	err := setSessionData(w, r, sessionState, state)
	if err != nil {
		fmt.Println(err)
	}

	http.Redirect(w, r, url, 302)
}

func LoginCallbackHandler(w http.ResponseWriter, r *http.Request) {

	// Check the state
	state, err := getSessionData(r, sessionState)
	if err != nil {
		fmt.Println(err)
	}

	if state != r.URL.Query().Get("state") {
		fmt.Println(errors.New("invalid state"))
	}

	// Save token
	tok, err := client.GetToken(r)
	if err != nil {
		fmt.Println(err)
	}

	if tok != nil {

		j, err := json.Marshal(tok)
		if err != nil {
			fmt.Println(err)
		}

		setSessionData(w, r, sessionToken, string(j))
	}

	fmt.Println(tok)

	http.Redirect(w, r, "/", 302)
	return
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {

	c := client

	tokString, err := getSessionData(r, sessionToken)
	if err != nil {
		fmt.Println(err)
	}

	tok := new(oauth2.Token)

	err = json.Unmarshal([]byte(tokString), tok)
	if err != nil {
		fmt.Println(err)
	}

	c.SetToken(tok)

	posts, err := c.GetPosts("steam", SortTop, AgeMonth)
	if err != nil {
		fmt.Println(err)
	}

	t := homeTemplate{}
	t.Items = posts.Data.Children

	returnTemplate(w, r, "home", t)
}

type homeTemplate struct {
	Items []ListingPost
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
