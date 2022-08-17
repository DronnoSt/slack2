package slackdump

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime/trace"
	"time"

	"errors"

	"github.com/slack-go/slack"
	"golang.org/x/time/rate"

	"github.com/rusq/slackdump/v2/auth"
	"github.com/rusq/slackdump/v2/fsadapter"
	"github.com/rusq/slackdump/v2/internal/network"
	"github.com/rusq/slackdump/v2/internal/structures"
	"github.com/rusq/slackdump/v2/logger"
	"github.com/rusq/slackdump/v2/types"
)

//go:generate mockgen -destination internal/mocks/mock_os/mock_os.go os FileInfo
//go:generate mockgen -destination internal/mocks/mock_downloader/mock_downloader.go github.com/rusq/slackdump/v2/downloader Downloader
//go:generate sh -c "mockgen -source slackdump.go -destination clienter_mock_test.go -package slackdump -mock_names clienter=mockClienter,Reporter=mockReporter"
//go:generate sed -i ~ -e "s/NewmockClienter/newmockClienter/g" -e "s/NewmockReporter/newmockReporter/g" clienter_mock_test.go

// Session stores basic session parameters.
type Session struct {
	client clienter // Slack client

	wspInfo *slack.AuthTestResponse // workspace info

	fs fsadapter.FS // filesystem for saving attachments

	// Users contains the list of users and populated on NewSession
	Users     types.Users          `json:"users"`
	UserIndex structures.UserIndex `json:"-"`

	options Options
}

// clienter is the interface with some functions of slack.Client with the sole
// purpose of mocking in tests (see client_mock.go)
type clienter interface {
	GetConversationInfoContext(ctx context.Context, channelID string, includeLocale bool) (*slack.Channel, error)
	GetConversationHistoryContext(ctx context.Context, params *slack.GetConversationHistoryParameters) (*slack.GetConversationHistoryResponse, error)
	GetConversationRepliesContext(ctx context.Context, params *slack.GetConversationRepliesParameters) (msgs []slack.Message, hasMore bool, nextCursor string, err error)
	GetConversationsContext(ctx context.Context, params *slack.GetConversationsParameters) (channels []slack.Channel, nextCursor string, err error)
	GetFile(downloadURL string, writer io.Writer) error
	GetTeamInfo() (*slack.TeamInfo, error)
	GetUsersContext(ctx context.Context, options ...slack.GetUsersOption) ([]slack.User, error)
}

// Errors
var (
	ErrNoUserCache = errors.New("user cache unavailable")
)

// AllChanTypes enumerates all API-supported channel types as of 03/2022.
var AllChanTypes = []string{"mpim", "im", "public_channel", "private_channel"}

// New creates new session with the default options  and populates the internal
// cache of users and channels for lookups.
func New(ctx context.Context, creds auth.Provider, opts ...Option) (*Session, error) {
	options := DefOptions
	for _, opt := range opts {
		opt(&options)
	}

	return NewWithOptions(ctx, creds, options)
}

// New creates new Session with provided options, and populates the internal
// cache of users and channels for lookups.  If it fails to authenticate,
// AuthError is returned.
func NewWithOptions(ctx context.Context, authProvider auth.Provider, opts Options) (*Session, error) {
	ctx, task := trace.NewTask(ctx, "NewWithOptions")
	defer task.End()

	if err := authProvider.Validate(); err != nil {
		return nil, err
	}

	cl := slack.New(authProvider.SlackToken(), slack.OptionCookieRAW(toPtrCookies(authProvider.Cookies())...))

	authTestResp, err := cl.AuthTestContext(ctx)
	if err != nil {
		return nil, &AuthError{Err: err}
	}

	sd := &Session{
		client:  cl,
		options: opts,
		wspInfo: authTestResp,
		fs:      fsadapter.NewDirectory("."), // default is to save attachments to the current directory.
	}

	sd.propagateLogger(sd.l())

	if err := os.MkdirAll(opts.CacheDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create the cache directory: %s", err)
	}

	sd.l().Println("> checking user cache...")
	users, err := sd.GetUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("error fetching users: %w", err)
	}

	sd.Users = users
	sd.UserIndex = users.IndexByID()

	return sd, nil
}

// TestAuth attempts to authenticate with the given provider.  It will return
// AuthError if faled.
func TestAuth(ctx context.Context, provider auth.Provider) error {
	ctx, task := trace.NewTask(ctx, "TestAuth")
	defer task.End()

	cl := slack.New(provider.SlackToken(), slack.OptionCookieRAW(toPtrCookies(provider.Cookies())...))

	region := trace.StartRegion(ctx, "AuthTestContext")
	defer region.End()
	_, err := cl.AuthTestContext(ctx)
	if err != nil {
		return &AuthError{Err: err}
	}
	return nil
}

// Client returns the underlying slack.Client.
func (sd *Session) Client() *slack.Client {
	return sd.client.(*slack.Client)
}

// Me returns the current authenticated user in a rather dirty manner.
// If the user cache is unitnitialised, it returns ErrNoUserCache.
func (sd *Session) Me() (slack.User, error) {
	if len(sd.UserIndex) == 0 {
		return slack.User{}, ErrNoUserCache
	}
	return *sd.UserIndex[sd.CurrentUserID()], nil
}

func (sd *Session) CurrentUserID() string {
	return sd.wspInfo.UserID
}

// SetFS sets the filesystem to save attachments to (slackdump defaults to the
// current directory otherwise).
func (sd *Session) SetFS(fs fsadapter.FS) {
	if fs == nil {
		return
	}
	sd.fs = fs
}

func toPtrCookies(cc []http.Cookie) []*http.Cookie {
	var ret = make([]*http.Cookie, len(cc))
	for i := range cc {
		ret[i] = &cc[i]
	}
	return ret
}

func (sd *Session) limiter(t network.Tier) *rate.Limiter {
	return network.NewLimiter(t, sd.options.Tier3Burst, int(sd.options.Tier3Boost))
}

// withRetry will run the callback function fn. If the function returns
// slack.RateLimitedError, it will delay, and then call it again up to
// maxAttempts times. It will return an error if it runs out of attempts.
func withRetry(ctx context.Context, l *rate.Limiter, maxAttempts int, fn func() error) error {
	return network.WithRetry(ctx, l, maxAttempts, fn)
}

func checkCacheFile(filename string, maxAge time.Duration) error {
	if filename == "" {
		return errors.New("no cache filename")
	}
	fi, err := os.Stat(filename)
	if err != nil {
		return err
	}

	return validateCache(fi, maxAge)
}

func validateCache(fi os.FileInfo, maxAge time.Duration) error {
	if fi.IsDir() {
		return errors.New("cache file is a directory")
	}
	if fi.Size() == 0 {
		return errors.New("empty cache file")
	}
	if time.Since(fi.ModTime()) > maxAge {
		return errors.New("cache expired")
	}
	return nil
}

// l returns the current logger.
func (sd *Session) l() logger.Interface {
	if sd.options.Logger == nil {
		return logger.Default
	}
	return sd.options.Logger
}

// propagateLogger propagates the slackdump logger to some dumb packages.
func (sd *Session) propagateLogger(l logger.Interface) {
	network.Logger = l
}

// returnOrIgnore returns an error if sd.options.IgnoreErrors is true, and nil
// otherwise.
func (sd *Session) returnOrIgnore(caller string, err error) error {
	if err == nil {
		return nil
	}
	if sd.options.IgnoreErrors {
		sd.l().Printf("%s: ignoring error: %s", caller, err)
		return nil
	} else {
		return err
	}
}
