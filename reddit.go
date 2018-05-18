package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/beefsack/go-rate"
	"golang.org/x/oauth2"
)

const (
	defaultUserAgent = "github.com/Jleagle/reddit-go"

	authURL        = "https://www.reddit.com/api/v1/authorize"
	authCompactURL = "https://www.reddit.com/api/v1/authorize.compact"
	tokenURL       = "https://www.reddit.com/api/v1/access_token"
	apiURL         = "https://oauth.reddit.com/"
)

type AuthScope string

const (
	ScopeAccount          AuthScope = "account"          // Update preferences and related account information. Will not have access to your email or password.
	ScopeCreddits                   = "creddits"         // Spend my reddit gold creddits on giving gold to other users.
	ScopeEdit                       = "edit"             // Edit and delete my comments and submissions.
	ScopeFlair                      = "flair"            // Select my subreddit flair. Change link flair on my submissions.
	ScopeHistory                    = "history"          // Access my voting history and comments or submissions I've saved or hidden.
	ScopeIdentity                   = "identity"         // Access my reddit username and signup date.
	ScopeLivemanage                 = "livemanage"       // Manage settings and contributors of live threads I contribute to.
	ScopeModconfig                  = "modconfig"        // Manage the configuration, sidebar, and CSS of subreddits I moderate.
	ScopeModcontributors            = "modcontributors"  // Add/remove users to approved submitter lists and ban/unban or mute/unmute users from subreddits I moderate.
	ScopeModflair                   = "modflair"         // Manage and assign flair in subreddits I moderate.
	ScopeModlog                     = "modlog"           // Access the moderation log in subreddits I moderate.
	ScopeModmail                    = "modmail"          // Access and manage modmail via mod.reddit.com.
	ScopeModothers                  = "modothers"        // Invite or remove other moderators from subreddits I moderate.
	ScopeModposts                   = "modposts"         // Approve, remove, mark nsfw, and distinguish content in subreddits I moderate.
	ScopeModself                    = "modself"          // Accept invitations to moderate a subreddit. Remove myself as a moderator or contributor of subreddits I moderate or contribute to.
	ScopeModtraffic                 = "modtraffic"       // Access traffic stats in subreddits I moderate.
	ScopeModwiki                    = "modwiki"          // Change editors and visibility of wiki pages in subreddits I moderate.
	ScopeMysubreddits               = "mysubreddits"     // Access the list of subreddits I moderate, contribute to, and subscribe to.
	ScopePrivatemessages            = "privatemessages"  // Access my inbox and send private messages to other users.
	ScopeRead                       = "read"             // Access posts and comments through my account.
	ScopeReport                     = "report"           // Report content for rules violations. Hide & show individual submissions.
	ScopeSave                       = "save"             // Save and unsave comments and submissions.
	ScopeStructuredstyles           = "structuredstyles" // Edit structured styles for a subreddit I moderate.
	ScopeSubmit                     = "submit"           // Submit links and comments from my account.
	ScopeSubscribe                  = "subscribe"        // Manage my subreddit subscriptions. Manage \"friends\" - users whose content I follow.
	ScopeVote                       = "vote"             // Submit and change my votes on comments and submissions.
	ScopeWikiedit                   = "wikiedit"         // Edit wiki pages on my behalf.
	ScopeWikiread                   = "wikiread"         // Read wiki pages through my account.
)

type ListingSort string

const (
	SortDefault       ListingSort = ""
	SortHot                       = "hot"
	SortNew                       = "new"
	SortRising                    = "rising"
	SortTop                       = "top"
	SortControversial             = "controversial"
)

type ListingTime string

const (
	TimeDefault ListingTime = ""
	TimeHour                = "hour"
	TimeDay                 = "day"
	TimeWeek                = "week"
	TimeMonth               = "month"
	TimeYear                = "year"
	TimeAllTime             = "all"
)

type ListingLocation string

const (
	GLOBAL ListingLocation = "GLOBAL"
	US                     = "US"
	AR                     = "AR"
	AU                     = "AU"
	BG                     = "BG"
	CA                     = "CA"
	CL                     = "CL"
	CO                     = "CO"
	HR                     = "HR"
	CZ                     = "CZ"
	FI                     = "FI"
	GR                     = "GR"
	HU                     = "HU"
	IS                     = "IS"
	IN                     = "IN"
	IE                     = "IE"
	JP                     = "JP"
	MY                     = "MY"
	MX                     = "MX"
	NZ                     = "NZ"
	PH                     = "PH"
	PL                     = "PL"
	PT                     = "PT"
	PR                     = "PR"
	RO                     = "RO"
	RS                     = "RS"
	SG                     = "SG"
	SE                     = "SE"
	TW                     = "TW"
	TH                     = "TH"
	TR                     = "TR"
	GB                     = "GB"
	US_WA                  = "US_WA"
	US_DE                  = "US_DE"
	US_DC                  = "US_DC"
	US_WI                  = "US_WI"
	US_WV                  = "US_WV"
	US_HI                  = "US_HI"
	US_FL                  = "US_FL"
	US_WY                  = "US_WY"
	US_NH                  = "US_NH"
	US_NJ                  = "US_NJ"
	US_NM                  = "US_NM"
	US_TX                  = "US_TX"
	US_LA                  = "US_LA"
	US_NC                  = "US_NC"
	US_ND                  = "US_ND"
	US_NE                  = "US_NE"
	US_TN                  = "US_TN"
	US_NY                  = "US_NY"
	US_PA                  = "US_PA"
	US_CA                  = "US_CA"
	US_NV                  = "US_NV"
	US_VA                  = "US_VA"
	US_CO                  = "US_CO"
	US_AK                  = "US_AK"
	US_AL                  = "US_AL"
	US_AR                  = "US_AR"
	US_VT                  = "US_VT"
	US_IL                  = "US_IL"
	US_GA                  = "US_GA"
	US_IN                  = "US_IN"
	US_IA                  = "US_IA"
	US_OK                  = "US_OK"
	US_AZ                  = "US_AZ"
	US_ID                  = "US_ID"
	US_CT                  = "US_CT"
	US_ME                  = "US_ME"
	US_MD                  = "US_MD"
	US_MA                  = "US_MA"
	US_OH                  = "US_OH"
	US_UT                  = "US_UT"
	US_MO                  = "US_MO"
	US_MN                  = "US_MN"
	US_MI                  = "US_MI"
	US_RI                  = "US_RI"
	US_KS                  = "US_KS"
	US_MT                  = "US_MT"
	US_MS                  = "US_MS"
	US_SC                  = "US_SC"
	US_KY                  = "US_KY"
	US_OR                  = "US_OR"
	US_SD                  = "US_SD"
)

var (
	errNoToken = errors.New("no token set")
	errNoCode  = errors.New("no code found")
)

type transport struct {
	http.RoundTripper // Interface
	useragent string
}

func (t *transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {

	req.Header.Set("User-Agent", t.useragent)
	return t.RoundTripper.RoundTrip(req)
}

type Reddit struct {
	oauthConfig oauth2.Config
	ctx         context.Context
	httpClient  *http.Client
	throttle    *rate.RateLimiter
}

func GetClient(client string, secret string, redirect string, userAgent string) (reddit Reddit) {

	if userAgent == "" {
		userAgent = defaultUserAgent
	}

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
		oauthConfig: config,
		ctx:         context.Background(),
	}

	// Use the custom HTTP client when requesting a token.
	httpClient := &http.Client{
		Timeout:   2 * time.Second,
		Transport: &transport{http.DefaultTransport, userAgent},
	}

	reddit.ctx = context.WithValue(reddit.ctx, oauth2.HTTPClient, httpClient)

	return reddit
}

func (r *Reddit) Throttle(duration time.Duration) {
	if duration == 0 {
		r.throttle = nil
	} else {
		r.throttle = rate.New(1, duration)
	}
}

func (r Reddit) Login(scopes []AuthScope, compact bool, state string) (string, string) {

	// Set scopes
	r.oauthConfig.Scopes = []string{}
	for _, v := range scopes {
		r.oauthConfig.Scopes = append(r.oauthConfig.Scopes, string(v))
	}

	// Set auth URL
	if compact {
		r.oauthConfig.Endpoint.AuthURL = authCompactURL
	}

	// Generate state
	if state == "" {
		state = strconv.Itoa(int(rand.Int31()))
	}

	// Generate login URL
	u := r.oauthConfig.AuthCodeURL(
		state,
		oauth2.SetAuthURLParam("response_type", "code"),
		oauth2.SetAuthURLParam("duration", "permanent"),
	)

	return u, state
}

func (r Reddit) GetToken(re *http.Request) (tok *oauth2.Token, err error) {

	code := re.URL.Query().Get("code")

	if code == "" {
		return tok, errNoCode
	}

	return r.oauthConfig.Exchange(r.ctx, code)
}

func (r *Reddit) SetToken(tok *oauth2.Token) {
	r.httpClient = r.oauthConfig.Client(r.ctx, tok)
}

func (r Reddit) GetPosts(options ListingOptions) (posts *ListingResponse, err error) {

	err = options.Validate()
	if err != nil {
		fmt.Println(err.Error())
	}

	q := url.Values{}

	if options.After != "" {
		q.Set("after", options.After)
	}
	if options.Before != "" {
		q.Set("before", options.Before)
	}
	if options.Count > 0 {
		q.Set("count", strconv.Itoa(options.Count))
	}
	if options.Limit > 0 {
		q.Set("limit", strconv.Itoa(options.Limit))
	}
	if options.Show {
		q.Set("show", "all")
	}
	if options.Detail {
		q.Set("sr_detail", "")
	}
	if options.Time == SortTop || options.Time == SortControversial {
		q.Set("t", string(options.Time))
	}

	var u = apiURL
	if options.Reddit != "" {
		u = u + "r/" + options.Reddit
	}
	if options.Sort != "" {
		u = u + "/" + string(options.Sort)
	}
	encoded := q.Encode()
	if encoded != "" {
		u = u + "?" + encoded
	}

	posts = new(ListingResponse)
	err = r.fetch(u, posts)
	if err != nil {
		return posts, err
	}

	return posts, err
}

type ListingOptions struct {
	After    string
	Before   string
	Count    int
	Limit    int
	Show     bool
	Detail   bool
	Location ListingLocation // Hot only
	Time     ListingTime     // Top & Controversial only
	Sort     ListingSort
	Reddit   string
}

func (l ListingOptions) Validate() error {

	return nil
}

type ListingResponse struct {
	Kind string `json:"kind"`
	Data struct {
		Modhash  string        `json:"modhash"`
		Dist     int           `json:"dist"`
		Children []ListingPost `json:"children"`
		After    string        `json:"after"`
		Before   interface{}   `json:"before"`
	} `json:"data"`
}

type ListingPost struct {
	Kind string          `json:"kind"`
	Data ListingPostData `json:"data"`
}

type ListingPostData struct {
	ApprovedAtUtc  interface{}   `json:"approved_at_utc"`
	Subreddit      string        `json:"subreddit"`
	Selftext       string        `json:"selftext"`
	UserReports    []interface{} `json:"user_reports"`
	IsSaved        bool          `json:"saved"`
	ModReasonTitle interface{}   `json:"mod_reason_title"`
	Gilded         int           `json:"gilded"`
	IsClicked      bool          `json:"clicked"`
	Title          string        `json:"title"`
	LinkFlairRichtext []struct {
		E string `json:"e"`
		T string `json:"t"`
	} `json:"link_flair_richtext"`
	SubredditNamePrefixed      string      `json:"subreddit_name_prefixed"`
	IsHidden                   bool        `json:"hidden"`
	Pwls                       int         `json:"pwls"`
	LinkFlairCSSClass          string      `json:"link_flair_css_class"`
	Downs                      int         `json:"downs"`
	ThumbnailHeight            int         `json:"thumbnail_height"`
	ParentWhitelistStatus      string      `json:"parent_whitelist_status"`
	HideScore                  bool        `json:"hide_score"`
	Name                       string      `json:"name"`
	Quarantine                 bool        `json:"quarantine"`
	LinkFlairTextColor         string      `json:"link_flair_text_color"`
	AuthorFlairBackgroundColor interface{} `json:"author_flair_background_color"`
	SubredditType              string      `json:"subreddit_type"`
	Ups                        int         `json:"ups"`
	Domain                     string      `json:"domain"`
	MediaEmbed struct {
	} `json:"media_embed"`
	ThumbnailWidth        int         `json:"thumbnail_width"`
	AuthorFlairTemplateID interface{} `json:"author_flair_template_id"`
	IsOriginalContent     bool        `json:"is_original_content"`
	SecureMedia           interface{} `json:"secure_media"`
	IsRedditMediaDomain   bool        `json:"is_reddit_media_domain"`
	Category              interface{} `json:"category"`
	SecureMediaEmbed struct {
	} `json:"secure_media_embed"`
	LinkFlairText string      `json:"link_flair_text"`
	CanModPost    bool        `json:"can_mod_post"`
	Score         int         `json:"score"`
	ApprovedBy    interface{} `json:"approved_by"`
	Thumbnail     string      `json:"thumbnail"`
	//Edited              bool          `json:"edited"` // Timestamp or false
	AuthorFlairCSSClass string        `json:"author_flair_css_class"`
	AuthorFlairRichtext []interface{} `json:"author_flair_richtext"`
	PostHint            string        `json:"post_hint"`
	IsSelf              bool          `json:"is_self"`
	ModNote             interface{}   `json:"mod_note"`
	Created             float64       `json:"created"`
	LinkFlairType       string        `json:"link_flair_type"`
	Wls                 int           `json:"wls"`
	PostCategories      interface{}   `json:"post_categories"`
	BannedBy            interface{}   `json:"banned_by"`
	AuthorFlairType     string        `json:"author_flair_type"`
	ContestMode         bool          `json:"contest_mode"`
	SelftextHTML        interface{}   `json:"selftext_html"`
	Likes               interface{}   `json:"likes"`
	SuggestedSort       interface{}   `json:"suggested_sort"`
	BannedAtUtc         interface{}   `json:"banned_at_utc"`
	ViewCount           interface{}   `json:"view_count"`
	Archived            bool          `json:"archived"`
	NoFollow            bool          `json:"no_follow"`
	IsCrosspostable     bool          `json:"is_crosspostable"`
	Pinned              bool          `json:"pinned"`
	IsOver18            bool          `json:"over_18"`
	Preview struct {
		Images []struct {
			Source struct {
				URL    string `json:"url"`
				Width  int    `json:"width"`
				Height int    `json:"height"`
			} `json:"source"`
			Resolutions []struct {
				URL    string `json:"url"`
				Width  int    `json:"width"`
				Height int    `json:"height"`
			} `json:"resolutions"`
			Variants struct {
			} `json:"variants"`
			ID string `json:"id"`
		} `json:"images"`
		Enabled bool `json:"enabled"`
	} `json:"preview"`
	CanGild              bool          `json:"can_gild"`
	IsSpoiler            bool          `json:"spoiler"`
	Locked               bool          `json:"locked"`
	AuthorFlairText      string        `json:"author_flair_text"`
	RteMode              string        `json:"rte_mode"`
	IsVisited            bool          `json:"visited"`
	NumReports           interface{}   `json:"num_reports"`
	Distinguished        interface{}   `json:"distinguished"`
	SubredditID          string        `json:"subreddit_id"`
	ModReasonBy          interface{}   `json:"mod_reason_by"`
	RemovalReason        interface{}   `json:"removal_reason"`
	ID                   string        `json:"id"`
	ReportReasons        interface{}   `json:"report_reasons"`
	Author               string        `json:"author"`
	NumCrossposts        int           `json:"num_crossposts"`
	NumComments          int           `json:"num_comments"`
	SendReplies          bool          `json:"send_replies"`
	ModReports           []interface{} `json:"mod_reports"`
	AuthorFlairTextColor interface{}   `json:"author_flair_text_color"`
	Permalink            string        `json:"permalink"`
	WhitelistStatus      string        `json:"whitelist_status"`
	Stickied             bool          `json:"stickied"`
	URL                  string        `json:"url"`
	SubredditSubscribers int           `json:"subreddit_subscribers"`
	CreatedUtc           float64       `json:"created_utc"`
	Media                interface{}   `json:"media"`
	IsVideo              bool          `json:"is_video"`
}

func (d ListingPostData) IsImage() bool {
	return strings.HasSuffix(d.URL, ".jpg") || strings.HasSuffix(d.URL, ".jpeg") || strings.HasSuffix(d.URL, ".png") || strings.HasSuffix(d.URL, ".gif")
}

func (r Reddit) fetch(url string, i interface{}) (err error) {

	if r.httpClient == nil {
		return errNoToken
	}

	if r.throttle != nil {
		r.throttle.Wait()
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, i)
	if err != nil {
		return err
	}

	return nil
}
