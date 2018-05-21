package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/Jleagle/reddit-go/reddit"
	"github.com/mssola/user_agent"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {

	ua := user_agent.New(r.UserAgent())

	u, state := client.Login([]reddit.AuthScope{reddit.ScopeRead, reddit.ScopeSave}, ua.Mobile(), "")

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
