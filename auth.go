package main

import (
	"fmt"
	"net/http"
	"math/rand"
	"strconv"
	"errors"
	"encoding/json"
)

var scopes = []string{ScopeIdentity, ScopeRead, ScopeHistory, ScopeSubscribe}

func LoginHandler(w http.ResponseWriter, r *http.Request) {

	client := GetClient(scopes)

	client, err := getSteamClient(r)
	if err != nil {
		fmt.Println(err)
	}

	// Generate state
	s := strconv.Itoa(int(rand.Int31()))

	// Save state
	err = setSessionData(w, r, sessionState, s)
	if err != nil {
		fmt.Println(err)
	}

	// Redirect
	url := client.AuthCodeURL(s, []string{"identity", "read", "history", "subscribe"})

	http.Redirect(w, r, url, 302)
	return
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

	// Get token
	client, err := getSteamClient(r)
	if err != nil {
		fmt.Println(err)
	}

	t, err := client.OAuthConfig.Exchange(redditContext, r.URL.Query().Get("code"))

	// Save token
	j, err := json.Marshal(t)
	if err != nil {
		fmt.Println(err)
	}

	setSessionData(w, r, sessionToken, string(j))

	http.Redirect(w, r, "/", 302)
	return
}
