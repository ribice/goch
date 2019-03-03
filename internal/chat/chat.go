package chat

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"

	"github.com/ribice/goch"

	"github.com/ribice/msv/bind"
	"github.com/ribice/msv/render"
)

var (
	exceedsAny func(map[string]goch.Limit) error
	exceeds    func(string, goch.Limit) error
	alfaRgx    *regexp.Regexp
	mailRgx    *regexp.Regexp
)

// Limiter represents chat service limit checker
type Limiter interface {
	Exceeds(string, goch.Limit) error
	ExceedsAny(map[string]goch.Limit) error
}

// NewAPI creates new websocket api
func NewAPI(store Store, l Limiter) *API {
	api := API{
		store: store,
	}

	exceeds = l.Exceeds
	alfaRgx = regexp.MustCompile("^[a-zA-Z0-9_]*$")
	mailRgx = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

	// api.RegisterEndpoint(
	// 	"POST",
	// 	"/admin/create_channel",
	// 	api.createChannel,
	// 	WithHTTPBasicAuth(admin, password),
	// )

	// api.RegisterEndpoint(
	// 	"POST",
	// 	"/admin/unread_count",
	// 	api.unreadCount,
	// 	WithHTTPBasicAuth(admin, password),
	// )

	// api.RegisterHandler("GET", "/list_channels", api.listChannels)
	// api.RegisterEndpoint("POST", "/register_nick", api.registerNick)
	// api.RegisterEndpoint("POST", "/channel_members", api.channelMembers)

	return &api
}

// API represents websocket api service
type API struct {
	store Store
}

// Store represents chat store interface
type Store interface {
	Save(*goch.Chat) error
	Get(string) (*goch.Chat, error)
	ListChannels() ([]string, error)
	GetUnreadCount(string, string) uint64
}

type createChanReq struct {
	Name      string `json:"name"`
	IsPrivate bool   `json:"is_private"`
}

func (cr *createChanReq) Bind() error {
	if err := exceeds(cr.Name, goch.ChanLimit); err != nil {
		return err
	}
	if match, err := regexp.MatchString("^[a-zA-Z0-9_]*$", cr.Name); !match || err != nil {
		return fmt.Errorf("name must contain only alphanumeric and underscores")
	}
	return nil
}

func (api *API) createChannel(w http.ResponseWriter, r *http.Request) {
	var req createChanReq
	if err := bind.JSON(w, r, &req); err != nil {
		return
	}
	ch := goch.NewChannel(req.Name, req.IsPrivate)
	if err := api.store.Save(ch); err != nil {
		http.Error(w, fmt.Sprintf("could not create channel: %v", err), 500)
		return
	}
	render.JSON(w, r, ch.Secret)
}

type registerReq struct {
	UID           string `json:"uid"`
	DisplayName   string `json:"display_name"`
	Email         string `json:"email"`
	Secret        string `json:"secret"`
	Channel       string `json:"channel"`
	ChannelSecret string `json:"channel_secret"` // Tennant
}

type registerResp struct {
	Secret string `json:"secret"`
}

func (r *registerReq) Bind() error {
	if !alfaRgx.MatchString(r.UID) {
		return errors.New("uid must contain only alphanumeric and underscores")
	}
	if !alfaRgx.MatchString(r.Secret) {
		return errors.New("secret must contain only alphanumeric and underscores")
	}
	if !mailRgx.MatchString(r.Email) {
		return errors.New("invalid email address")
	}

	return exceedsAny(map[string]goch.Limit{
		r.UID:           goch.UIDLimit,
		r.DisplayName:   goch.DisplayNameLimit,
		r.ChannelSecret: goch.ChanSecretLimit,
		r.Secret:        goch.SecretLimit,
		r.Channel:       goch.ChanLimit,
	})
}

func (api *API) registerNick(w http.ResponseWriter, r *http.Request) {
	var req registerReq
	if err := bind.JSON(w, r, &req); err != nil {
		return
	}
	ch, err := api.store.Get(req.Channel)
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid secret or unexisting channel: %v", err), 500)
		return
	}

	if ch.Secret != req.ChannelSecret {
		http.Error(w, fmt.Sprintf("invalid secret or unexisting channel: %v", err), 500)
		return
	}

	secret, err := ch.Register(&goch.User{
		UID:         req.UID,
		DisplayName: req.DisplayName,
		Email:       req.Email,
	}, req.Secret)

	if err != nil {
		http.Error(w, fmt.Sprintf("error registering to channel: %v", err), 500)
		return
	}

	if err = api.store.Save(ch); err != nil {
		ch.Leave(req.UID)
		http.Error(w, fmt.Sprintf("could not update channel membership: %v", err), 500)
		return
	}

	render.JSON(w, r, secret)

}

type unreadCountReq struct {
	Channel string `json:"channel"`
	UID     string `json:"uid"`
}

type unreadCountResp struct {
	Count uint64 `json:"count"`
}

func (r *unreadCountReq) Bind() error {
	return exceedsAny(map[string]goch.Limit{
		r.UID:     goch.UIDLimit,
		r.Channel: goch.ChanLimit,
	})
}

func (api *API) unreadCount(w http.ResponseWriter, r *http.Request) {
	var req unreadCountReq
	if err := bind.JSON(w, r, &req); err != nil {
		return
	}
	uc := api.store.GetUnreadCount(req.UID, req.Channel)
	render.JSON(w, r, &unreadCountResp{uc})

}

type channelMembersReq struct {
	Channel       string `json:"channel"`
	ChannelSecret string `json:"channel_secret"`
}

func (r *channelMembersReq) Bind() error {
	return exceedsAny(map[string]goch.Limit{
		r.Channel:       goch.ChanLimit,
		r.ChannelSecret: goch.ChanSecretLimit,
	})
}

func (api *API) channelMembers(w http.ResponseWriter, r *http.Request) {
	var req channelMembersReq
	if err := bind.JSON(w, r, &req); err != nil {
		return
	}

	ch, err := api.store.Get(req.Channel)
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid secret or unexisting channel: %v", err), 500)
		return
	}

	if ch.Secret != req.ChannelSecret {
		http.Error(w, fmt.Sprintf("invalid secret or unexisting channel: %v", err), 500)
		return
	}

	render.JSON(w, r, ch.ListMembers())
}

func (api *API) listChannels(w http.ResponseWriter, r *http.Request) {
	chans, err := api.store.ListChannels()
	if err != nil {
		http.Error(w, fmt.Sprintf("unable to fetch channels: %v", err), 500)
		return
	}
	render.JSON(w, r, chans)

}
