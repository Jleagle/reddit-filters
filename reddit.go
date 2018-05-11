package main

import (
	"strconv"
	"math/rand"
	"golang.org/x/oauth2"
	"net/http"
	"context"
	"time"
)

const (
	userAgent      = "Reddit Filters (https://github.com/Jleagle/reddit-filters)"
	authURL        = "https://www.reddit.com/api/v1/authorize"
	authCompactURL = "https://www.reddit.com/api/v1/authorize.compact"
	tokenURL       = "https://www.reddit.com/api/v1/access_token"
)

const (
	ScopeAccount          = "account"          // Update preferences and related account information. Will not have access to your email or password.
	ScopeCreddits         = "creddits"         // Spend my reddit gold creddits on giving gold to other users.
	ScopeEdit             = "edit"             // Edit and delete my comments and submissions.
	ScopeFlair            = "flair"            // Select my subreddit flair. Change link flair on my submissions.
	ScopeHistory          = "history"          // Access my voting history and comments or submissions I've saved or hidden.
	ScopeIdentity         = "identity"         // Access my reddit username and signup date.
	ScopeLivemanage       = "livemanage"       // Manage settings and contributors of live threads I contribute to.
	ScopeModconfig        = "modconfig"        // Manage the configuration, sidebar, and CSS of subreddits I moderate.
	ScopeModcontributors  = "modcontributors"  // Add/remove users to approved submitter lists and ban/unban or mute/unmute users from subreddits I moderate.
	ScopeModflair         = "modflair"         // Manage and assign flair in subreddits I moderate.
	ScopeModlog           = "modlog"           // Access the moderation log in subreddits I moderate.
	ScopeModmail          = "modmail"          // Access and manage modmail via mod.reddit.com.
	ScopeModothers        = "modothers"        // Invite or remove other moderators from subreddits I moderate.
	ScopeModposts         = "modposts"         // Approve, remove, mark nsfw, and distinguish content in subreddits I moderate.
	ScopeModself          = "modself"          // Accept invitations to moderate a subreddit. Remove myself as a moderator or contributor of subreddits I moderate or contribute to.
	ScopeModtraffic       = "modtraffic"       // Access traffic stats in subreddits I moderate.
	ScopeModwiki          = "modwiki"          // Change editors and visibility of wiki pages in subreddits I moderate.
	ScopeMysubreddits     = "mysubreddits"     // Access the list of subreddits I moderate, contribute to, and subscribe to.
	ScopePrivatemessages  = "privatemessages"  // Access my inbox and send private messages to other users.
	ScopeRead             = "read"             // Access posts and comments through my account.
	ScopeReport           = "report"           // Report content for rules violations. Hide & show individual submissions.
	ScopeSave             = "save"             // Save and unsave comments and submissions.
	ScopeStructuredstyles = "structuredstyles" // Edit structured styles for a subreddit I moderate.
	ScopeSubmit           = "submit"           // Submit links and comments from my account.
	ScopeSubscribe        = "subscribe"        // Manage my subreddit subscriptions. Manage \"friends\" - users whose content I follow.
	ScopeVote             = "vote"             // Submit and change my votes on comments and submissions.
	ScopeWikiedit         = "wikiedit"         // Edit wiki pages on my behalf.
	ScopeWikiread         = "wikiread"         // Read wiki pages through my account.
)

type transport struct {
	http.RoundTripper
	useragent string
}

func (t *transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {

	req.Header.Set("User-Agent", t.useragent)
	req.Header.Set("Host", "www.reddit.com")

	return t.RoundTripper.RoundTrip(req)
}

type Reddit struct {
	Agent        string
	CompactLogin bool
	OauthConfig  oauth2.Config
	ctx          context.Context
	httpClient   *http.Client
}



func GetClient(client string, secret string, redirect string) (reddit Reddit) {

	config := oauth2.Config{
		ClientID:     client,
		ClientSecret: secret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  authURL,
			TokenURL: tokenURL,
		},
		RedirectURL: redirect,
	}

	reddit = Reddit{
		Agent:       userAgent,
		OauthConfig: config,
		ctx:         context.Background(),
	}

	// Use the custom HTTP client when requesting a token.
	httpClient := &http.Client{
		Timeout: 2 * time.Second,
	}

	reddit.ctx = context.WithValue(reddit.ctx, oauth2.HTTPClient, httpClient)

	return reddit
}

func (r Reddit) Login(scopes []string, compact bool) (url string, state string) {

	r.OauthConfig.Scopes = scopes

	if compact {
		r.OauthConfig.Endpoint.AuthURL = authCompactURL
	}

	state = strconv.Itoa(int(rand.Int31()))

	url = r.OauthConfig.AuthCodeURL(
		state,
		oauth2.SetAuthURLParam("response_type", "code"),
		oauth2.SetAuthURLParam("duration", "permanent"),
	)

	return url, state
}

func (r Reddit) GetToken(code string) (url string, state string, err error) {

	tok, err := r.OauthConfig.Exchange(r.ctx, code)
	if err != nil {
		return
	}

	r.httpClient = r.OauthConfig.Client(r.ctx, tok)
	return
}
