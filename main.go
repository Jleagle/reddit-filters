package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi"
	"golang.org/x/oauth2"
)

var scopes = []AuthScope{ScopeIdentity, ScopeRead, ScopeHistory, ScopeSubscribe}

var client = GetClient(
	os.Getenv("REDDIT_CLIENT"),
	os.Getenv("REDDIT_SECRET"),
	"http://localhost:8087/login/callback",
	"Reddit Filters",
)

func main() {

	r := chi.NewRouter()

	r.Get("/", HomeHandler)
	r.Get("/listing", ListingHandler)
	r.Get("/login", LoginHandler)
	r.Get("/login/callback", LoginCallbackHandler)

	// File server
	fileServer(r)

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

	// todo, handle errors, no code etc

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

	http.Redirect(w, r, "/", 302)
	return
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {

	t := homeTemplate{}
	returnTemplate(w, r, "home", t)
}

type homeTemplate struct {
}

func ListingHandler(w http.ResponseWriter, r *http.Request) {

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

	options := ListingOptions{}
	options.After = r.URL.Query().Get("last")

	posts, err := c.GetPosts("all", SortTop, AgeMonth, options)
	if err != nil {
		fmt.Println(err)
	}

	var ret []listingItemTemplate
	var lastID string

	for _, v := range posts.Data.Children {

		lastID = v.Kind + "_" + v.Data.ID

		if v.Data.Thumbnail == "self" {
			v.Data.Thumbnail = "/assets/logo.png"
		}

		// Filters

		ret = append(ret, listingItemTemplate{
			ID:    v.Kind + "_" + v.Data.ID,
			Title: v.Data.Title,
			Icon:  v.Data.Thumbnail,
		})
	}

	//// Save the token to session again, incase it has been refreshed
	//tok, err = c.GetToken(r)
	//if err != nil {
	//	fmt.Println(err)
	//}
	//
	//j, err := json.Marshal(tok)
	//if err != nil {
	//	fmt.Println(err)
	//}
	//
	//err = setSessionData(w, r, sessionToken, string(j))
	//if err != nil {
	//	fmt.Println(err)
	//}

	// Encode
	b, err := json.Marshal(listingTemplate{
		LastID: lastID,
		Items:  ret,
	})
	if err != nil {
		fmt.Println(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

type listingTemplate struct {
	Items  []listingItemTemplate `json:"items"`
	LastID string                `json:"last_id"`
}

type listingItemTemplate struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Icon  string `json:"icon"`
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

func fileServer(r chi.Router) {

	path := "/assets"

	if strings.ContainsAny(path, "{}*") {
		fmt.Println("FileServer does not permit URL parameters")
	}

	fs := http.StripPrefix(path, http.FileServer(http.Dir("./assets")))

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}))
}
