package agent

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"

	"github.com/gorilla/mux"

	"github.com/ribice/goch"

	"github.com/gorilla/websocket"
	"github.com/ribice/goch/internal/broker"
)

var (
	alfaRgx *regexp.Regexp
)

// NewAPI creates new websocket api
func NewAPI(m *mux.Router, br *broker.Broker, store ChatStore, lim Limiter) *API {
	api := API{
		broker: br,
		store:  store,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     func(r *http.Request) bool { return true },
		},
	}
	alfaRgx = regexp.MustCompile("^[a-zA-Z0-9_]*$")

	m.HandleFunc("/connect", api.connect).Methods("GET")

	return &api
}

// API represents websocket api service
type API struct {
	broker   *broker.Broker
	store    ChatStore
	upgrader websocket.Upgrader
	rlim     Limiter
}

// Limiter represents chat service limit checker
type Limiter interface {
	ExceedsAny(map[string]goch.Limit) error
}

func (api *API) connect(w http.ResponseWriter, r *http.Request) {
	conn, err := api.upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("error while upgrading to ws connection: %v", err), 500)
		return
	}

	req, err := api.waitConnInit(conn)
	if err != nil {
		if err == errConnClosed {
			return
		}
		writeErr(conn, err.Error())
		return
	}

	agent := New(api.broker, api.store)
	agent.HandleConn(conn, req)
}

type initConReq struct {
	Channel string  `json:"channel"`
	UID     string  `json:"uid"`
	Secret  string  `json:"secret"` // User secret
	LastSeq *uint64 `json:"last_seq"`
}

func (api *API) bindReq(r *initConReq) error {
	if !alfaRgx.MatchString(r.Secret) {
		return errors.New("secret must contain only alphanumeric and underscores")
	}
	if !alfaRgx.MatchString(r.Channel) {
		return errors.New("channel must contain only alphanumeric and underscores")
	}

	return api.rlim.ExceedsAny(map[string]goch.Limit{
		r.UID:     goch.UIDLimit,
		r.Secret:  goch.SecretLimit,
		r.Channel: goch.ChanLimit,
	})
}

var errConnClosed = errors.New("connection closed")

func (api *API) waitConnInit(conn *websocket.Conn) (*initConReq, error) {
	t, wsr, err := conn.NextReader()
	if err != nil || t == websocket.CloseMessage {
		return nil, errConnClosed
	}

	var req *initConReq

	err = json.NewDecoder(wsr).Decode(req)
	if err != nil {
		return nil, err
	}

	if err = api.bindReq(req); err != nil {
		return nil, err
	}

	return req, nil
}
