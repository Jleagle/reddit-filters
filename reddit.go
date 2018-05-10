package main

import (
	"strconv"
	"math/rand"
	"golang.org/x/oauth2"
	"os"
)

var reddit Reddit

func GetClient(scopes []string) (reddit Reddit) {

	var authURL string
	if reddit.CompactLogin {
		authURL = "https://www.reddit.com/api/v1/authorize.compact"
	} else {
		authURL = "https://www.reddit.com/api/v1/authorize"
	}

	config := oauth2.Config{
		ClientID:     os.Getenv("REDDIT_CLIENT"),
		ClientSecret: os.Getenv("REDDIT_SECRET"),
		Endpoint: oauth2.Endpoint{
			AuthURL:  authURL,
			TokenURL: "https://www.reddit.com/api/v1/access_token",
		},
		RedirectURL: "http://localhost:8087/login/callback",
		Scopes:      scopes,
	}

	reddit = Reddit{
		Agent:        "Reddit Filters (https://github.com/Jleagle/reddit-filters)",
		OauthConfig:  &config,
		CompactLogin: false,
	}

	return reddit
}

type Reddit struct {
	Agent        string
	OauthConfig  *oauth2.Config
	CompactLogin bool
}

var ParamResponseType = oauth2.SetAuthURLParam("response_type", "code")
var ParamDuration = oauth2.SetAuthURLParam("duration", "permanent")

func (r Reddit) AuthPath() (path string, state string) {

	state = strconv.Itoa(int(rand.Int31()))

	return reddit.OauthConfig.AuthCodeURL(state, ParamResponseType, ParamDuration), state
}

var (
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
