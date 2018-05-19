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

var client = GetClient(
	os.Getenv("REDDIT_CLIENT"),
	os.Getenv("REDDIT_SECRET"),
	"http://localhost:8087/login/callback",
	"Reddit Filters",
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

	err := http.ListenAndServe(":8087", r)
	if err != nil {
		fmt.Println(err)
	}
}

type globalTemplate struct {
	IsLoggedIn bool
	Reddit     string
	Query      url.Values
}

func (g *globalTemplate) Fill(r *http.Request) {

	token, err := getSessionData(r, sessionToken)
	if err != nil {
		fmt.Println(err)
	}

	g.IsLoggedIn = token != ""
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {

	u, state := client.Login([]AuthScope{ScopeRead, ScopeSave}, false, "")

	err := setSessionData(w, r, sessionState, state)
	if err != nil {
		fmt.Println(err)
	}

	http.Redirect(w, r, u, 302)
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

func LogoutHandler(w http.ResponseWriter, r *http.Request) {

	err := setSessionData(w, r, sessionToken, "")
	if err != nil {
		fmt.Println(err)
	}

	http.Redirect(w, r, "/", 302)
	return
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {

	q := r.URL.Query()

	t := homeTemplate{}
	t.Fill(r)
	t.Query = q
	t.Reddit = chi.URLParam(r, "id")

	//t.Sort = q.Get("sort")
	//t.Time = q.Get("time")
	//t.Location = q.Get("location")

	returnTemplate(w, "home", t)
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

	returnTemplate(w, "info", t)
}

type infoTemplate struct {
	globalTemplate
}

func ListingHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	c, retu, err := getClientForAjax(w, r)
	if err != nil {
		fmt.Println(err)
	}
	if retu {
		return
	}

	q := r.URL.Query()

	options := ListingOptions{}
	options.Reddit = q.Get("reddit")
	options.After = q.Get("last")
	options.Time = ListingTime(q.Get("time"))
	options.Sort = ListingSort(q.Get("sort"))

	posts, err := c.GetListing(options)
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

		if q.Get("videos") == "t" && !v.Data.IsVideo {
			continue
		} else if q.Get("videos") == "f" && v.Data.IsVideo {
			continue
		}

		if q.Get("selfs") == "t" && !v.Data.IsSelf {
			continue
		} else if q.Get("selfs") == "f" && v.Data.IsSelf {
			continue
		}

		if q.Get("spoilers") == "t" && !v.Data.IsSpoiler {
			continue
		} else if q.Get("spoilers") == "f" && v.Data.IsSpoiler {
			continue
		}

		if q.Get("saved") == "t" && !v.Data.IsSaved {
			continue
		} else if q.Get("saved") == "f" && v.Data.IsSaved {
			continue
		}

		if q.Get("clicked") == "t" && !v.Data.IsClicked {
			continue
		} else if q.Get("clicked") == "f" && v.Data.IsClicked {
			continue
		}

		if q.Get("hidden") == "t" && !v.Data.IsHidden {
			continue
		} else if q.Get("hidden") == "f" && v.Data.IsHidden {
			continue
		}

		if q.Get("visited") == "t" && !v.Data.IsVisited {
			continue
		} else if q.Get("visited") == "f" && v.Data.IsVisited {
			continue
		}

		if q.Get("original") == "t" && !v.Data.IsOriginalContent {
			continue
		} else if q.Get("original") == "f" && v.Data.IsOriginalContent {
			continue
		}

		if q.Get("nsfw") == "t" && !v.Data.IsOver18 {
			continue
		} else if q.Get("nsfw") == "f" && v.Data.IsOver18 {
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
			Saved:         v.Data.IsSaved,
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
	Saved         bool   `json:"saved"`
}

func SaveHandler(w http.ResponseWriter, r *http.Request) {

	//w.Header().Set("Content-Type", "application/json")

	c, retu, err := getClientForAjax(w, r)
	if err != nil {
		fmt.Println(err)
	}
	if retu {
		return
	}

	_, err = c.Save(r.URL.Query().Get("id"), "")
	if err != nil {
		fmt.Println(err)
	}

	w.Write([]byte("OK"))
}

func UnsaveHandler(w http.ResponseWriter, r *http.Request) {

	//w.Header().Set("Content-Type", "application/json")

	c, retu, err := getClientForAjax(w, r)
	if err != nil {
		fmt.Println(err)
	}
	if retu {
		return
	}

	_, err = c.Unsave(r.URL.Query().Get("id"))
	if err != nil {
		fmt.Println(err)
	}

	w.Write([]byte("OK"))
}

func getClientForAjax(w http.ResponseWriter, r *http.Request) (c Reddit, ret bool, err error) {

	c = client

	tokString, err := getSessionData(r, sessionToken)
	if err != nil {
		return c, false, err
	}

	if tokString == "" {
		w.Write(errorToJsonBytes("not logged in"))
		return c, true, err
	}

	tok := new(oauth2.Token)

	err = json.Unmarshal([]byte(tokString), tok)
	if err != nil {
		return c, false, err
	}

	c.SetToken(tok)

	return c, false, err
}

func errorToJsonBytes(error string) ([]byte) {

	j, err := json.Marshal(struct{ Error string `json:"error"` }{error})
	if err != nil {
		return []byte(err.Error())
	}

	return j
}

func returnTemplate(w http.ResponseWriter, page string, pageData interface{}) (err error) {

	w.Header().Set("Content-Type", "text/html")

	// Load templates needed
	t, err := template.New("t").Funcs(templateFuncs()).ParseFiles("templates/_header.html", "templates/_footer.html", "templates/"+page+".html")
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
