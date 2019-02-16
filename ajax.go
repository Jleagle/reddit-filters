package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/Jleagle/reddit-go/reddit"
	"golang.org/x/oauth2"
)

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

	options := reddit.ListingOptions{}
	options.Reddit = q.Get("reddit")
	options.After = q.Get("last")
	options.Time = reddit.ListingTime(q.Get("time"))
	options.Sort = reddit.ListingSort(q.Get("sort"))

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

	c, retu, err := getClientForAjax(w, r)
	if err != nil {
		fmt.Println(err)
	}
	if retu {
		return
	}

	err = c.Save(r.URL.Query().Get("id"), "")
	if err != nil {
		fmt.Println(err)
	}

	w.Write([]byte("OK"))
}

func UnsaveHandler(w http.ResponseWriter, r *http.Request) {

	c, retu, err := getClientForAjax(w, r)
	if err != nil {
		fmt.Println(err)
	}
	if retu {
		return
	}

	err = c.Unsave(r.URL.Query().Get("id"))
	if err != nil {
		fmt.Println(err)
	}

	w.Write([]byte("OK"))
}

func getClientForAjax(w http.ResponseWriter, r *http.Request) (c reddit.Reddit, ret bool, err error) {

	c = client

	tokString, err := getSessionData(r, sessionToken)
	if err != nil {
		return c, false, err
	}

	if tokString == "" {
		w.Write(errorToJsonBytes("<a class=\"nav-link\" href=\"/login\">Login</a>"))
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
