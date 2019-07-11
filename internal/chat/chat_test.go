package chat_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/ribice/goch"
	"github.com/ribice/goch/internal/chat"
	"github.com/ribice/goch/pkg/config"
)

var cfg = &config.Config{
	Limits: map[goch.Limit][2]int{
		goch.DisplayNameLimit: [2]int{3, 128},
		goch.UIDLimit:         [2]int{20, 20},
		goch.SecretLimit:      [2]int{20, 50},
		goch.ChanLimit:        [2]int{10, 20},
		goch.ChanSecretLimit:  [2]int{20, 20},
	},
	LimitErrs: map[goch.Limit]error{
		goch.DisplayNameLimit: errors.New("displayName must be between 3 and 128 characters long"),
		goch.UIDLimit:         errors.New("uid must be exactly 20 characters long"),
		goch.SecretLimit:      errors.New("secret must be between 20 and 50 characters long"),
		goch.ChanLimit:        errors.New("channel must be between 10 and 20 characters long"),
		goch.ChanSecretLimit:  errors.New("channelSecret must be exactly 20 characters long"),
	},
}

type createChanReq struct {
	Name      string `json:"name"`
	IsPrivate bool   `json:"is_private"`
}

type createChanResp struct {
	Secret string `json:"secret"`
}

func middleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	})
}
func TestCreateChannel(t *testing.T) {
	cases := []struct {
		name        string
		store       *store
		req         *createChanReq
		wantMessage string
		wantCode    int
	}{
		{
			name:        "ChannelName length validation fail",
			req:         &createChanReq{},
			wantCode:    http.StatusBadRequest,
			wantMessage: "error binding request: channel must be between 10 and 20 characters long",
		},
		{
			name:        "ChannelName alfanum fail",
			req:         &createChanReq{Name: "$%a!"},
			wantCode:    http.StatusBadRequest,
			wantMessage: "error binding request: name must contain only alphanumeric and underscores",
		},
		{
			name: "Fail on saving channel",
			req:  &createChanReq{Name: "abcdefghijklmnop"},
			store: &store{
				SaveFunc: func(*goch.Chat) error {
					return errors.New("error saving channel")
				},
			},
			wantMessage: "could not create channel: error saving channel",
			wantCode:    http.StatusInternalServerError,
		},
		{
			name: "Create public channel",
			req:  &createChanReq{Name: "abcdefghijklmnop"},
			store: &store{
				SaveFunc: func(*goch.Chat) error {
					return nil
				},
			},
			wantCode: http.StatusOK,
		},
		{
			name: "Create private channel",
			req:  &createChanReq{Name: "abcdefghijklmnop", IsPrivate: true},
			store: &store{
				SaveFunc: func(*goch.Chat) error {
					return nil
				},
			},
			wantCode: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := mux.NewRouter()
			chat.New(m, tc.store, cfg, middleware)
			srv := httptest.NewServer(m)
			defer srv.Close()
			path := srv.URL + "/admin/channels"

			req, err := json.Marshal(tc.req)
			if err != nil {
				t.Error(err)
			}

			res, err := http.Post(path, "application/json", bytes.NewBuffer(req))
			if err != nil {
				t.Error(err)
			}

			if tc.wantCode != res.StatusCode {
				t.Errorf("unexpected response code. want: %d, got: %d", tc.wantCode, res.StatusCode)
			}

			bts, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Error(err)
			}

			if tc.wantMessage != "" && tc.wantMessage != strings.TrimSpace(string(bts)) {
				t.Errorf("unexpected response. want: %v, got: %v", tc.wantMessage, string(bts))
			}

		})
	}
}

type registerReq struct {
	UID           string `json:"uid"`
	DisplayName   string `json:"display_name"`
	Email         string `json:"email"`
	Secret        string `json:"secret"`
	Channel       string `json:"channel"`
	ChannelSecret string `json:"channel_secret"`
}

type registerResp struct {
	Secret string `json:"secret"`
}

func TestRegister(t *testing.T) {
	cases := []struct {
		name       string
		store      *store
		req        registerReq
		wantCode   int
		wantErrMsg string
	}{
		{
			name:       "validation Test: Empty request",
			wantCode:   http.StatusBadRequest,
			wantErrMsg: "error binding request: invalid email address",
		},
		{
			name:     "validation Test: Invalid UID",
			wantCode: http.StatusBadRequest,
			req: registerReq{
				UID:           "joe??",
				Channel:       "foo",
				DisplayName:   "qwertyuiopasdfghjklvv",
				Email:         "ribice@gmail.com",
				ChannelSecret: "qwertyuiopasdfghjklvv",
			},
			wantErrMsg: "error binding request: uid must contain only alphanumeric and underscores",
		},
		{
			name:     "validation Test: Invalid Secret",
			wantCode: http.StatusBadRequest,
			req: registerReq{
				UID:           "12324Ab",
				Channel:       "foo",
				DisplayName:   "qwertyuiopasdfghjklvv",
				Email:         "ribice@gmail.com",
				ChannelSecret: ">??>^^@!#$@$1@$",
				Secret:        ">??>^^@!#$@$1@$",
			},
			wantErrMsg: "error binding request: secret must contain only alphanumeric and underscores",
		},
		{
			store: &store{
				GetFunc: func(id string) (*goch.Chat, error) {
					return nil, errors.New("err fetching chan")
				},
			},
			name:       "Error fetching channel",
			req:        registerReq{UID: "EmirABCDEF1234567890", Channel: "foo1234567", Email: "ribice@gmail.com", ChannelSecret: "ABCDEFGHIJDKLOMNSOPR", DisplayName: "Emir", Secret: "12345678901234567890ABC"},
			wantCode:   http.StatusInternalServerError,
			wantErrMsg: "invalid secret or unexisting channel: err fetching chan",
		},
		{
			store: &store{
				GetFunc: func(id string) (*goch.Chat, error) {
					return &goch.Chat{Secret: "foo"}, nil
				},
			},
			name:     "test invalid secret",
			req:      registerReq{UID: "EmirABCDEF1234567890", Channel: "foo1234567", Email: "ribice@gmail.com", ChannelSecret: "ABCDEFGHIJDKLOMNSOPR", DisplayName: "Emir", Secret: "12345678901234567890ABC"},
			wantCode: http.StatusInternalServerError,
		},
		{
			store: &store{
				GetFunc: func(id string) (*goch.Chat, error) {
					return &goch.Chat{Secret: "ABCDEFGHIJDKLOMNSOPR", Members: map[string]*goch.User{
						"EmirABCDEF1234567890": &goch.User{},
					}}, nil
				},
			},
			name:       "test uid already registered",
			req:        registerReq{UID: "EmirABCDEF1234567890", Channel: "foo1234567", Email: "ribice@gmail.com", ChannelSecret: "ABCDEFGHIJDKLOMNSOPR", DisplayName: "Emir", Secret: "12345678901234567890ABC"},
			wantCode:   http.StatusInternalServerError,
			wantErrMsg: "error registering to channel: chat: uid already registered in this chat",
		},
		{
			store: &store{
				GetFunc: func(id string) (*goch.Chat, error) {
					return &goch.Chat{Secret: "ABCDEFGHIJDKLOMNSOPR", Members: map[string]*goch.User{
						"EmirABCDEF1234567890ASD": &goch.User{},
					}}, nil
				},
				SaveFunc: func(*goch.Chat) error {
					return errors.New("error saving to redis")
				},
			},
			name:       "test uid already registered",
			req:        registerReq{UID: "EmirABCDEF1234567890", Channel: "foo1234567", Email: "ribice@gmail.com", ChannelSecret: "ABCDEFGHIJDKLOMNSOPR", DisplayName: "Emir", Secret: "12345678901234567890ABC"},
			wantCode:   http.StatusInternalServerError,
			wantErrMsg: "could not update channel membership: error saving to redis",
		},
		{
			store: &store{
				GetFunc: func(id string) (*goch.Chat, error) {
					return &goch.Chat{Secret: "ABCDEFGHIJDKLOMNSOPR", Members: map[string]*goch.User{
						"EmirABCDEF1234567890ASD": &goch.User{},
					}}, nil
				},
				SaveFunc: func(ch *goch.Chat) error { return nil },
			},
			name:     "success",
			req:      registerReq{UID: "EmirABCDEF1234567890", Channel: "foo1234567", Email: "ribice@gmail.com", ChannelSecret: "ABCDEFGHIJDKLOMNSOPR", DisplayName: "Emir", Secret: "12345678901234567890ABC"},
			wantCode: http.StatusOK,
		},
	}

	type registerResp struct {
		Secret string `json:"string"`
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := mux.NewRouter()
			chat.New(m, tc.store, cfg, middleware)
			srv := httptest.NewServer(m)
			defer srv.Close()
			path := srv.URL + "/channels/register"

			req, err := json.Marshal(tc.req)
			if err != nil {
				t.Error(err)
			}

			res, err := http.Post(path, "application/json", bytes.NewBuffer(req))
			if err != nil {
				t.Error(err)
			}

			if res.StatusCode != tc.wantCode {
				t.Errorf("unexpected response code. want: %d, got: %d", tc.wantCode, res.StatusCode)
			}

			bts, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Error(err)
			}

			msg := strings.TrimSpace(string(bts))

			if tc.wantErrMsg != "" && tc.wantErrMsg != msg {
				t.Errorf("expected message: %v but got: %v", tc.wantErrMsg, msg)
			}

			if tc.wantCode == http.StatusOK {

				secret := msg[11 : len(msg)-2]

				if tc.req.Secret != "" && secret != tc.req.Secret {
					t.Errorf("invalid secret received, expected %v got %v", tc.req.Secret, secret)
				}

			}

		})
	}
}

type unreadCountResp struct {
	Count uint64 `json:"count"`
}

func TestUnreadCount(t *testing.T) {
	cases := []struct {
		name     string
		store    *store
		chanName string
		uid      string
		wantCode int
		wantResp uint64
	}{{
		name:     "fail on limits",
		chanName: "channel",
		uid:      "uid",
		wantCode: 400,
	},
		{
			name:     "Success",
			chanName: "12345678901",
			uid:      "1234567890ABCDEFGHIJ",
			store: &store{
				GetUnreadCountFunc: func(string, string) uint64 { return 12 },
			},
			wantCode: 200,
			wantResp: 12,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := mux.NewRouter()
			chat.New(m, tc.store, cfg, middleware)
			srv := httptest.NewServer(m)
			defer srv.Close()
			path := srv.URL + "/admin/channels/" + tc.chanName + "/user/" + tc.uid
			res, err := http.Get(path)
			if err != nil {
				t.Error(err)
			}

			if res.StatusCode != tc.wantCode {
				t.Errorf("unexpected response code. want: %d, got: %d", tc.wantCode, res.StatusCode)
			}

			if res.StatusCode == 200 {
				bts, err := ioutil.ReadAll(res.Body)
				if err != nil {
					t.Error(err)
				}

				var uc unreadCountResp

				if err := json.Unmarshal(bts, &uc); err != nil {
					t.Error(err)
				}

				if uc.Count != tc.wantResp {
					t.Errorf("expected count: %v but got: %v", tc.wantResp, uc.Count)
				}
			}
		})
	}
}

func TestListMembers(t *testing.T) {
	cases := []struct {
		name     string
		store    *store
		chanName string
		secret   string
		wantCode int
		want     []goch.User
	}{
		{
			name:     "Fail on validation",
			chanName: "abc",
			secret:   "?secret=123",
			wantCode: http.StatusBadRequest,
		},
		{
			store: &store{
				GetFunc: func(id string) (*goch.Chat, error) {
					return nil, errors.New("err fetching chan")
				},
			},
			name:     "error fetching channel",
			chanName: "1234567890",
			secret:   "?secret=12345678901234567890",
			wantCode: http.StatusInternalServerError,
		},
		{
			name: "invalid secret",
			store: &store{
				GetFunc: func(id string) (*goch.Chat, error) {
					return &goch.Chat{
						Secret: "invalid",
					}, nil
				},
			},
			chanName: "1234567890",
			secret:   "?secret=12345678901234567890",
			wantCode: http.StatusInternalServerError,
		},
		{
			store: &store{
				GetFunc: func(id string) (*goch.Chat, error) {
					return &goch.Chat{
						Secret: "12345678901234567890", Members: map[string]*goch.User{
							"joe": {UID: "joe"},
						},
					}, nil
				},
			},
			name:     "test success",
			chanName: "1234567890",
			secret:   "?secret=12345678901234567890",
			want:     []goch.User{{UID: "joe"}},
			wantCode: http.StatusOK,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := mux.NewRouter()
			chat.New(m, tc.store, cfg, middleware)
			srv := httptest.NewServer(m)
			defer srv.Close()
			path := srv.URL + "/channels/" + tc.chanName + tc.secret
			res, err := http.Get(path)
			if err != nil {
				t.Error(err)
			}

			if res.StatusCode != tc.wantCode {
				t.Errorf("unexpected response code. want: %d, got: %d", tc.wantCode, res.StatusCode)
			}

			if res.StatusCode == 200 {
				bts, err := ioutil.ReadAll(res.Body)
				if err != nil {
					t.Error(err)
				}

				var users []goch.User

				if err := json.Unmarshal(bts, &users); err != nil {
					t.Error(err)
				}

				if !reflect.DeepEqual(users, tc.want) {
					t.Errorf("expected users: %v but got: %v", tc.want, users)
				}
			}
		})
	}
}

func TestListChannels(t *testing.T) {
	cases := []struct {
		name     string
		store    *store
		wantCode int
		want     []string
	}{
		{
			store: &store{
				ListChansFunc: func() ([]string, error) {
					return nil, errors.New("err fetching chan")
				},
			},
			name:     "error fetching channels",
			wantCode: http.StatusInternalServerError,
		},
		{
			store: &store{
				ListChansFunc: func() ([]string, error) {
					return []string{"chan1", "chan2"}, nil
				},
			},
			name:     "success",
			wantCode: http.StatusOK,
			want:     []string{"chan1", "chan2"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := mux.NewRouter()
			chat.New(m, tc.store, cfg, middleware)
			srv := httptest.NewServer(m)
			defer srv.Close()
			path := srv.URL + "/admin/channels"
			res, err := http.Get(path)
			if err != nil {
				t.Error(err)
			}

			if res.StatusCode != tc.wantCode {
				t.Errorf("unexpected response code. want: %d, got: %d", tc.wantCode, res.StatusCode)
			}

			if res.StatusCode == 200 {
				bts, err := ioutil.ReadAll(res.Body)
				if err != nil {
					t.Error(err)
				}

				var chans []string

				if err := json.Unmarshal(bts, &chans); err != nil {
					t.Error(err)
				}

				if !reflect.DeepEqual(tc.want, chans) {
					t.Errorf("expected chans: %v but got: %v", tc.want, chans)
				}
			}

		})
	}
}

type store struct {
	SaveFunc           func(*goch.Chat) error
	GetFunc            func(string) (*goch.Chat, error)
	ListChansFunc      func() ([]string, error)
	GetUnreadCountFunc func(string, string) uint64
}

func (s *store) Save(c *goch.Chat) error           { return s.SaveFunc(c) }
func (s *store) Get(id string) (*goch.Chat, error) { return s.GetFunc(id) }
func (s *store) ListChannels() ([]string, error)   { return s.ListChansFunc() }
func (s *store) GetUnreadCount(uid, chanName string) uint64 {
	return s.GetUnreadCountFunc(uid, chanName)
}
