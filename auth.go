package main

import (
	"fmt"
	"net/http"
	"context"
	"math/rand"
	"strconv"
	"errors"
	"encoding/json"
	"golang.org/x/oauth2"
	"github.com/jzelinskie/geddit"
)

var redditClient *geddit.OAuthSession
var redditContext = context.Background()

func init() {

	var err error

	redditClient, err = geddit.NewOAuthSession(
		"EzQZsF8LWCwuEg",
		"9rnC59qajPrntK_dTL2RGRmsmAM",
		"Reddit Filters (https://github.com/Jleagle/reddit-filters)",
		"http://localhost:8087/login/callback",
	)
	if err != nil {
		fmt.Println(err)
	}
}

func getSteamClient(r *http.Request) (client geddit.OAuthSession, err error) {

	client = *redditClient

	// Get the token
	bytes, err := getSessionData(r, sessionToken)
	if err != nil {
		fmt.Println(err)
	}

	// If we have a token, set it on the client
	if len(bytes) > 0 {

		t := new(oauth2.Token)

		err = json.Unmarshal([]byte(bytes), t)
		if err != nil {
			fmt.Println(err)
		}

		client.Client = redditClient.OAuthConfig.Client(redditContext, t)
	}

	return client, err
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {

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
