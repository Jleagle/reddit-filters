package main

import (
	"github.com/jzelinskie/geddit"
	"fmt"
	"net/http"
)

var client *geddit.OAuthSession

func getClient() (*geddit.OAuthSession, error) {

	var err error

	if client == nil {

		client, err = geddit.NewOAuthSession(
			"EzQZsF8LWCwuEg",
			"9rnC59qajPrntK_dTL2RGRmsmAM",
			"Reddit Filters (https://github.com/Jleagle/reddit-filters)",
			"http://localhost:8087/login/callback",
		)
		if err != nil {
			fmt.Println(err)
		}

	}

	return client, err
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {

	client, err := getClient()
	if err != nil {
		fmt.Println(err)
	}

	// todo, generate & save state
	url := client.AuthCodeURL("state", []string{"identity", "read", "history", "subscribe"})

	http.Redirect(w, r, url, 302)
	return
}

func LoginCallbackHandler(w http.ResponseWriter, r *http.Request) {

	// todo, check state

	client, err := getClient()
	if err != nil {
		fmt.Println(err)
	}

	// Create and set token using given auth code.
	err = client.CodeAuth(r.URL.Query().Get("code"))
	if err != nil {
		fmt.Println(err)
	}

	http.Redirect(w, r, "/", 302)
	return
}
