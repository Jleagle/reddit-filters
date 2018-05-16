package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
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

	// Handle errors
	errStr := r.URL.Query().Get("error")
	if errStr != "" {
		fmt.Println(errStr)
	}

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

	http.Redirect(w, r, "/", 302)
	return
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {

	t := homeTemplate{}
	t.Query = r.URL.Query()

	returnTemplate(w, r, "home", t)
}

type homeTemplate struct {
	Query url.Values
}

func ListingHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()
	c := client

	tokString, err := getSessionData(r, sessionToken)
	if err != nil {
		fmt.Println(err)
	}

	if tokString == "" {

		j, err := json.Marshal(struct{ Error string `json:"error"` }{"not logged in"})
		if err != nil {
			fmt.Println(err)
		}

		w.Write(j)
		return
	}

	tok := new(oauth2.Token)

	err = json.Unmarshal([]byte(tokString), tok)
	if err != nil {
		fmt.Println(err)
	}

	c.SetToken(tok)

	options := ListingOptions{}
	options.After = q.Get("last")

	posts, err := c.GetPosts("all", SortTop, AgeMonth, options)
	if err != nil {
		fmt.Println(err)
	}

	var ret []listingItemTemplate
	var lastID string

	for _, v := range posts.Data.Children {

		lastID = v.Kind + "_" + v.Data.ID

		if !strings.HasPrefix(v.Data.Thumbnail, "http") {
			v.Data.Thumbnail = "/assets/logo.png"
		}

		if q.Get("images") == "t" && !v.Data.IsImage() {
			continue
		} else if q.Get("images") == "f" && v.Data.IsImage() {
			continue
		}

		if q.Get("nsfw") == "t" && !v.Data.Over18 {
			continue
		} else if q.Get("nsfw") == "f" && v.Data.Over18 {
			continue
		}

		ret = append(ret, listingItemTemplate{
			ID:            v.Kind + "_" + v.Data.ID,
			Title:         v.Data.Title,
			Icon:          v.Data.Thumbnail,
			Subreddit:     v.Data.Subreddit,
			Link:          v.Data.URL,
			CommentsLink:  "https://www.reddit.com" + v.Data.Permalink,
			CommentsCount: v.Data.NumComments,
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

	w.Write(b)
}

type listingTemplate struct {
	Items  []listingItemTemplate `json:"items"`
	LastID string                `json:"last_id"`
}

type listingItemTemplate struct {
	ID            string `json:"id"`
	Title         string `json:"title"`
	Icon          string `json:"icon"`
	Subreddit     string `json:"reddit"`
	Link          string `json:"link"`
	CommentsLink  string `json:"comments_link"`
	CommentsCount int    `json:"comments_count"`
}

func returnTemplate(w http.ResponseWriter, r *http.Request, page string, pageData interface{}) (err error) {

	w.Header().Set("Content-Type", "text/html")

	// Load templates needed
	t, err := template.New("t").Funcs(templateFuncs()).ParseFiles(page + ".html")
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

func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"has":    func(q url.Values, field string) bool { return q.Get(field) != "" },
		"ist":    func(q url.Values, field string) bool { return q.Get(field) == "t" },
		"isf":    func(q url.Values, field string) bool { return q.Get(field) == "f" },
		"option": func(q url.Values, field string, value string) bool { return q.Get(field) == value },
	}
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
