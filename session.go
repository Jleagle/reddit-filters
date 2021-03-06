package main

import (
	"net/http"
	"os"

	"github.com/gorilla/sessions"
)

const sessionToken = "token"
const sessionState = "state"

var sessionStore = sessions.NewCookieStore(
	[]byte(os.Getenv("REDDIT_SESSION_AUTHENTICATION")),
	[]byte(os.Getenv("REDDIT_SESSION_ENCRYPTION")),
)

func getSession(r *http.Request) (*sessions.Session, error) {

	session, err := sessionStore.Get(r, "reddit-filters-session")
	session.Options = &sessions.Options{
		MaxAge: 60 * 60 * 30,
		Path:   "/",
	}

	return session, err
}

func getSessionData(r *http.Request, key string) (value string, err error) {

	session, err := getSession(r)
	if err != nil {
		return "", err
	}

	if session.Values[key] == nil {
		session.Values[key] = ""
	}

	return session.Values[key].(string), err
}

func setSessionData(w http.ResponseWriter, r *http.Request, name string, value string) (err error) {

	session, err := getSession(r)
	if err != nil {
		return err
	}

	session.Values[name] = value

	return session.Save(r, w)
}
