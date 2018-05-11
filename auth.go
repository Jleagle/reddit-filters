package main

import (
	"fmt"
	"net/http"
	"errors"
	"os"
)

var scopes = []string{ScopeIdentity, ScopeRead, ScopeHistory, ScopeSubscribe}

var client = GetClient(
	os.Getenv("REDDIT_CLIENT"),
	os.Getenv("REDDIT_SECRET"),
	"http://localhost:8087/login/callback",
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {

	url, state := client.Login(scopes, false)

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

	//// Get token
	//client, err := getSteamClient(r)
	//if err != nil {
	//	fmt.Println(err)
	//}
	//
	//t, err := client.OAuthConfig.Exchange(redditContext, r.URL.Query().Get("code"))
	//
	//// Save token
	//j, err := json.Marshal(t)
	//if err != nil {
	//	fmt.Println(err)
	//}
	//
	//setSessionData(w, r, sessionToken, string(j))

	http.Redirect(w, r, "/", 302)
	return
}
