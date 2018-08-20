package main

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/Jleagle/reddit-go/reddit"
	"github.com/go-chi/chi"
)

var client = reddit.GetClient(
	os.Getenv("REDDIT_CLIENT"),
	os.Getenv("REDDIT_SECRET"),
	os.Getenv("REDDIT_AUTH_CALLBACK"),
	"reddit.jimeagle.com",
)

func main() {

	r := chi.NewRouter()

	r.Get("/", HomeHandler)
	r.Get("/r/{id}", HomeHandler)
	r.Get("/info", InfoHandler)

	r.Get("/login", LoginHandler)
	r.Get("/login/callback", LoginCallbackHandler)
	r.Get("/logout", LogoutHandler)

	r.Get("/ajax/listing", ListingHandler)
	r.Get("/ajax/save", SaveHandler)
	r.Get("/ajax/unsave", UnsaveHandler)

	// File server
	fileServer(r)

	err := http.ListenAndServe(":8080", r)
	if err != nil {
		fmt.Println(err)
	}
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {

	q := r.URL.Query()

	if q.Get("sort") == "" {
		q.Set("sort", "hot")
	}

	t := homeTemplate{}
	t.Query = q
	t.Reddit = chi.URLParam(r, "id")
	t.Fill(r)

	err := returnTemplate(w, "home", t)
	if err != nil {
		fmt.Println(err)
	}
}

type homeTemplate struct {
	globalTemplate
	Sort     string
	Time     string
	Location string
}

func InfoHandler(w http.ResponseWriter, r *http.Request) {

	t := infoTemplate{}
	t.Fill(r)

	err := returnTemplate(w, "info", t)
	if err != nil {
		fmt.Println(err)
	}
}

type infoTemplate struct {
	globalTemplate
}

type globalTemplate struct {
	IsLoggedIn bool
	Reddit     string
	RedditFull string
	Query      url.Values
}

func (g *globalTemplate) Fill(r *http.Request) {

	token, err := getSessionData(r, sessionToken)
	if err != nil {
		fmt.Println(err)
	}

	g.IsLoggedIn = token != ""

	if g.Reddit != "" {
		g.RedditFull = "/r/" + g.Reddit
	} else {
		g.RedditFull = "/"
	}
}

func returnTemplate(w http.ResponseWriter, page string, pageData interface{}) (err error) {

	w.Header().Set("Content-Type", "text/html")

	// Load templates needed
	t, err := template.New("t").Funcs(templateFuncs()).ParseFiles(
		"templates/_header.html",
		"templates/_footer.html",
		"templates/"+page+".html",
	)
	if err != nil {
		return err
	}

	// Write a respone
	buf := &bytes.Buffer{}
	err = t.ExecuteTemplate(buf, page, pageData)
	if err != nil {
		return err
	} else {
		// No error, send the content, HTTP 200 response status implied
		buf.WriteTo(w)
	}

	return nil
}

func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"get":    func(q url.Values, field string) string { return q.Get(field) },
		"has":    func(q url.Values, field string) bool { return q.Get(field) != "" },
		"ist":    func(q url.Values, field string) bool { return q.Get(field) == "t" },
		"isf":    func(q url.Values, field string) bool { return q.Get(field) == "f" },
		"option": func(q url.Values, field string, value string) bool { return q.Get(field) == value },
		"override": func(q url.Values, field string, value string) template.URL {
			var c = url.Values{}
			for k := range q {
				c.Set(k, q.Get(k))
			}
			c.Set(field, value)
			return template.URL(c.Encode())
		},
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
