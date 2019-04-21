package chat_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ribice/goch/pkg/config"

	"github.com/gorilla/mux"

	"github.com/ribice/goch"

	"github.com/ribice/goch/internal/chat"
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
		goch.UIDLimit:         errors.New("uid must be between 20 and 20 characters long"),
		goch.SecretLimit:      errors.New("secret must be between 20 and 50 characters long"),
		goch.ChanLimit:        errors.New("channel must be between 10 and 20 characters long"),
		goch.ChanSecretLimit:  errors.New("channelSecret must be between 20 and 20 characters long"),
	},
}

type response struct {
	Code   int             `json:"code"`
	Data   json.RawMessage `json:"data,omitempty"`
	Errors []string        `json:"errors,omitempty"`
}

type createChanReq struct {
	Name    string `json:"name"`
	Private bool   `json:"private"`
}

type createChanResp struct {
	Secret string `json:"secret"`
}

type errorResp struct {
	Message string `json:"message"`
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
		want        *createChanResp
		wantCode    int
	}{
		{
			name:        "test name req validation",
			req:         &createChanReq{Private: false},
			wantCode:    http.StatusBadRequest,
			wantMessage: "error binding data: channel must be between 10 and 20 characters long",
		},
		// {
		// 	name:     "test name length validation short",
		// 	req:      &createChanReq{Name: "a"},
		// 	wantCode: http.StatusBadRequest,
		// },
		// {
		// 	name:     "test name length validation long",
		// 	req:      &createChanReq{Name: "qwertyuiopasdfghjklzxcvbnk"},
		// 	wantCode: http.StatusBadRequest,
		// },
		// {
		// 	name:     "test name alphanumeric",
		// 	req:      &createChanReq{Name: "ak ; )___"},
		// 	wantCode: http.StatusBadRequest,
		// },
		// {
		// 	store: &store{
		// 		SaveFunc: func(c *goch.Chat) error {
		// 			return nil
		// 		},
		// 	},
		// 	name:     "test create public",
		// 	req:      &createChanReq{Name: "general"},
		// 	wantCode: http.StatusOK,
		// 	want:     &createChanResp{Secret: ""},
		// },
		// {
		// 	store: &store{
		// 		SaveFunc: func(c *goch.Chat) error {
		// 			return nil
		// 		},
		// 	},
		// 	name:     "test create private",
		// 	req:      &createChanReq{Name: "general", Private: true},
		// 	wantCode: http.StatusOK,
		// },
		// {
		// 	store: &store{
		// 		SaveFunc: func(c *goch.Chat) error {
		// 			return errors.New("could not store channel")
		// 		},
		// 	},
		// 	name:     "test store error",
		// 	req:      &createChanReq{Name: "general"},
		// 	wantCode: http.StatusInternalServerError,
		// },
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

			if res.StatusCode != 200 {
				bts, err := ioutil.ReadAll(res.Body)
				if err != nil {
					t.Error(err)
				}
				if tc.wantMessage != strings.TrimSpace(string(bts)) {
					t.Errorf("unexpected response. want: %v, got: %v", tc.wantMessage, string(bts))
				}
			}

			if tc.want != nil {
				var resp createChanResp
				if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
					t.Error(err)
				}
				if tc.want.Secret != resp.Secret {
					t.Errorf("unexpected response. want: %v, got: %v", *tc.want, resp.Secret)
				}
			}

		})
	}
}

type registerReq struct {
	UID           string `json:"uid"`
	FullName      string `json:"name"`
	Email         string `json:"email"`
	Secret        string `json:"secret"`
	Channel       string `json:"channel"`
	ChannelSecret string `json:"channel_secret"` // Tennant
}

type registerResp struct {
	Secret string `json:"secret"`
}

// func TestRegisterNick(t *testing.T) {
// 	cases := []struct {
// 		name     string
// 		store    *store
// 		req      registerReq
// 		wantErr  bool
// 		wantCode int
// 		want     string
// 	}{
// 		{
// 			name:     "test req channel validation",
// 			req:      registerReq{UID: "foo"},
// 			wantErr:  true,
// 			wantCode: http.StatusBadRequest,
// 		},
// 		{
// 			name:     "test req nick validation",
// 			req:      registerReq{Channel: "foo"},
// 			wantErr:  true,
// 			wantCode: http.StatusBadRequest,
// 		},
// 		{
// 			name:     "test nick short",
// 			req:      registerReq{UID: "jo", Channel: "foo"},
// 			wantErr:  true,
// 			wantCode: http.StatusBadRequest,
// 		},
// 		{
// 			name:     "test nick long",
// 			req:      registerReq{UID: "joefokjdislijflskdjfh", Channel: "foo"},
// 			wantErr:  true,
// 			wantCode: http.StatusBadRequest,
// 		},
// 		{
// 			name: "test nick long",
// 			req: registerReq{
// 				UID:           "joe123",
// 				Channel:       "foobar",
// 				ChannelSecret: "1123456789012345678901234567890123456789012345678901234567890234567890",
// 			},
// 			wantErr:  true,
// 			wantCode: http.StatusBadRequest,
// 		},
// 		{
// 			name: "test fields too long",
// 			req: registerReq{
// 				UID:           "joe",
// 				Channel:       "foo",
// 				FullName:      "qwertyuiopasdfghjklvv",
// 				Email:         "qwertyuiopasdfghjklvv",
// 				ChannelSecret: "qwertyuiopasdfghjklvv",
// 			},
// 			wantErr:  true,
// 			wantCode: http.StatusBadRequest,
// 		},
// 		{
// 			name:     "test nick alphanumeric",
// 			req:      registerReq{UID: " ;' joe", Channel: "foo"},
// 			wantErr:  true,
// 			wantCode: http.StatusBadRequest,
// 		},
// 		{
// 			store: &store{
// 				GetFunc: func(id string) (*goch.Chat, error) {
// 					return nil, errors.New("err fetching chan")
// 				},
// 			},
// 			name:     "test err fetch chan",
// 			req:      registerReq{UID: "joe", Channel: "foo"},
// 			wantErr:  true,
// 			wantCode: http.StatusInternalServerError,
// 		},
// 		{
// 			store: &store{
// 				GetFunc: func(id string) (*goch.Chat, error) {
// 					return &goch.Chat{Secret: "foo"}, nil
// 				},
// 			},
// 			name:     "test invalid secret",
// 			req:      registerReq{UID: "joe", Channel: "foo"},
// 			wantErr:  true,
// 			wantCode: http.StatusInternalServerError,
// 		},
// 		{
// 			store: &store{
// 				GetFunc: func(id string) (*goch.Chat, error) {
// 					return &goch.Chat{Secret: "", Members: map[string]*goch.User{"joe": {UID: "joe"}}}, nil
// 				},
// 			},
// 			name:     "test nick exists",
// 			req:      registerReq{UID: "joe", Channel: "foo"},
// 			wantErr:  true,
// 			wantCode: http.StatusInternalServerError,
// 		},
// 		{
// 			store: &store{
// 				GetFunc: func(id string) (*goch.Chat, error) {
// 					ch := goch.NewChannel("foo", false)
// 					ch.Secret = "xxxyyy"
// 					return ch, nil
// 				},
// 				SaveFunc: func(ch *goch.Chat) error {
// 					return errors.New("unable to save")
// 				},
// 			},
// 			name:     "test save failed",
// 			req:      registerReq{UID: "joe", Channel: "foo", ChannelSecret: "xxxyyy"},
// 			wantErr:  true,
// 			wantCode: http.StatusInternalServerError,
// 		},
// 		{
// 			store: &store{
// 				GetFunc: func(id string) (*goch.Chat, error) {
// 					ch := goch.NewChannel("foo", false)
// 					ch.Secret = "xxxyyy"
// 					return ch, nil
// 				},
// 				SaveFunc: func(ch *goch.Chat) error { return nil },
// 			},
// 			name:     "test saved",
// 			req:      registerReq{UID: "joe", Channel: "foo", ChannelSecret: "xxxyyy"},
// 			wantErr:  false,
// 			wantCode: http.StatusOK,
// 		},
// 		{
// 			store: &store{
// 				GetFunc: func(id string) (*goch.Chat, error) {
// 					ch := goch.NewChannel("foo", false)
// 					ch.Secret = "xxxyyy"
// 					return ch, nil
// 				},
// 				SaveFunc: func(ch *goch.Chat) error { return nil },
// 			},
// 			name:     "test provided secret length",
// 			req:      registerReq{UID: "joe", Channel: "foo", Secret: "qwertyuiopasdfghjklmnbvcxzhdguui", ChannelSecret: "xxxyyy"},
// 			wantErr:  true,
// 			wantCode: http.StatusBadRequest,
// 		},
// 		{
// 			store: &store{
// 				GetFunc: func(id string) (*goch.Chat, error) {
// 					ch := goch.NewChannel("foo", false)
// 					ch.Secret = "xxxyyy"
// 					return ch, nil
// 				},
// 				SaveFunc: func(ch *goch.Chat) error { return nil },
// 			},
// 			name:     "test provided secret alphanumeric",
// 			req:      registerReq{UID: "joe", Channel: "foo", Secret: "asljfkd ' ';", ChannelSecret: "xxxyyy"},
// 			wantErr:  true,
// 			wantCode: http.StatusBadRequest,
// 		},
// 		{
// 			store: &store{
// 				GetFunc: func(id string) (*goch.Chat, error) {
// 					ch := goch.NewChannel("foo", false)
// 					ch.Secret = "xxxyyy"
// 					return ch, nil
// 				},
// 				SaveFunc: func(ch *goch.Chat) error { return nil },
// 			},
// 			name:     "test saved with provided secret",
// 			req:      registerReq{UID: "joe", Channel: "foo", Secret: "foobarbaz", ChannelSecret: "xxxyyy"},
// 			wantErr:  false,
// 			wantCode: http.StatusOK,
// 		},

// 		// TODO - Test server username/pass (empty/nonempty)
// 	}

// 	for _, tc := range cases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			var handler h.HandlerFunc
// 			{
// 				api := chat.New(tc.store, "admin", "test")
// 				for path, ep := range api.Endpoints() {
// 					if path == "/register_nick" {
// 						handler = ep.Handler
// 					}
// 				}
// 			}

// 			req, _ := http.NewRequest("POST", "/register_nick", reqBody(t, tc.req))
// 			rw := httptest.NewRecorder()

// 			handler(context.Background(), rw, req)

// 			var resp response
// 			{
// 				if rw.Code != tc.wantCode {
// 					t.Errorf("unexpected response code. want: %d, got: %d", tc.wantCode, rw.Code)
// 				}

// 				if rw.Code == http.StatusUnauthorized {
// 					return
// 				}

// 				respBody(t, rw.Body, &resp)
// 				if tc.wantErr != (resp.Errors != nil) {
// 					t.Errorf("unexpected err response. want: %v, got: %+v", tc.wantErr, resp.Errors)
// 					return
// 				}

// 				if rw.Code != http.StatusOK {
// 					return
// 				}

// 				var got registerResp
// 				json.Unmarshal(resp.Data, &got)

// 				if tc.req.Secret != "" && got.Secret != tc.req.Secret {
// 					t.Errorf("custom secret not set")
// 				}

// 				if got.Secret == "" {
// 					t.Errorf("unexpected response. want nonempty secret")
// 				}
// 			}
// 		})
// 	}
// }

// type channelMembersReq struct {
// 	Channel       string `json:"channel"`
// 	ChannelSecret string `json:"channel_secret"`
// }

// func TestChannelMembers(t *testing.T) {
// 	cases := []struct {
// 		name     string
// 		store    *store
// 		req      channelMembersReq
// 		want     []goch.User
// 		wantErr  bool
// 		wantCode int
// 	}{
// 		{
// 			name:     "test req channel validation",
// 			req:      channelMembersReq{Channel: ""},
// 			wantErr:  true,
// 			wantCode: http.StatusBadRequest,
// 		},
// 		{
// 			name:     "test req channel length validation",
// 			req:      channelMembersReq{Channel: "aandasdfkjllandasdfkjllndasdfkjll"},
// 			wantErr:  true,
// 			wantCode: http.StatusBadRequest,
// 		},
// 		{
// 			name:     "test req channel length validation",
// 			req:      channelMembersReq{Channel: "foo", ChannelSecret: "aaandasdfkjllndasdfkjllaandasdfkjllndasdfkjllaandasdfkjllndasdfkjllaandasdfkj"},
// 			wantErr:  true,
// 			wantCode: http.StatusBadRequest,
// 		},
// 		{
// 			store: &store{
// 				GetFunc: func(id string) (*goch.Chat, error) {
// 					return nil, errors.New("err fetching chan")
// 				},
// 			},
// 			name:     "test err fetch chan",
// 			req:      channelMembersReq{Channel: "foo"},
// 			wantErr:  true,
// 			wantCode: http.StatusInternalServerError,
// 		},
// 		{
// 			store: &store{
// 				GetFunc: func(id string) (*goch.Chat, error) {
// 					return &goch.Chat{
// 						Secret: "",
// 						Members: map[string]*goch.User{
// 							"joe": {
// 								UID: "joe",
// 							},
// 							"foo": {
// 								UID: "foo",
// 							},
// 							"bar": {
// 								UID: "bar",
// 							},
// 							"baz": {
// 								UID: "baz",
// 							},
// 						},
// 					}, nil
// 				},
// 			},
// 			name:    "test success",
// 			req:     channelMembersReq{Channel: "foo"},
// 			wantErr: false,
// 			want: []goch.User{
// 				{
// 					UID: "joe",
// 				},
// 				{
// 					UID: "foo",
// 				},
// 				{
// 					UID: "bar",
// 				},
// 				{
// 					UID: "baz",
// 				},
// 			},
// 			wantCode: http.StatusOK,
// 		},

// 		// TODO - Test server username/pass (empty/nonempty)
// 	}

// 	for _, tc := range cases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			var handler h.HandlerFunc
// 			{
// 				api := chat.New(tc.store, "admin", "test")
// 				for path, ep := range api.Endpoints() {
// 					if path == "/channel_members" {
// 						handler = ep.Handler
// 					}
// 				}
// 			}

// 			req, _ := http.NewRequest("POST", "/channel_members", reqBody(t, tc.req))
// 			rw := httptest.NewRecorder()

// 			handler(context.Background(), rw, req)

// 			var resp response
// 			{
// 				if rw.Code != tc.wantCode {
// 					t.Errorf("unexpected response code. want: %d, got: %d", tc.wantCode, rw.Code)
// 				}

// 				if rw.Code == http.StatusUnauthorized {
// 					return
// 				}

// 				respBody(t, rw.Body, &resp)
// 				if tc.wantErr != (resp.Errors != nil) {
// 					t.Errorf("unexpected err response. want: %v, got: %+v", tc.wantErr, resp.Errors)
// 					return
// 				}

// 				if rw.Code != http.StatusOK {
// 					return
// 				}

// 				var got []goch.User
// 				json.Unmarshal(resp.Data, &got)

// 				for _, w := range tc.want {
// 					found := false
// 					for _, g := range got {
// 						if w.Nick == g.Nick {
// 							found = true
// 						}
// 					}
// 					if !found {
// 						t.Errorf("unexpected response. want: %v, got: %+v", tc.want, got)
// 					}
// 				}
// 			}
// 		})
// 	}
// }

// func TestListChannels(t *testing.T) {
// 	cases := []struct {
// 		name     string
// 		store    *store
// 		want     []string
// 		wantErr  bool
// 		wantCode int
// 	}{
// 		{
// 			store: &store{
// 				ListChansFunc: func() ([]string, error) {
// 					return nil, errors.New("err fetching chan")
// 				},
// 			},
// 			name:     "test err fetch chan",
// 			wantErr:  true,
// 			wantCode: http.StatusInternalServerError,
// 		},
// 		{
// 			store: &store{
// 				ListChansFunc: func() ([]string, error) {
// 					return []string{"general", "random"}, nil
// 				},
// 			},
// 			name:     "test success",
// 			wantErr:  false,
// 			want:     []string{"general", "random"},
// 			wantCode: http.StatusOK,
// 		},
// 	}

// 	for _, tc := range cases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			var handler h.HandlerFunc
// 			{
// 				api := chat.New(tc.store, "admin", "test")
// 				for path, ep := range api.Endpoints() {
// 					if path == "/list_channels" {
// 						handler = ep.Handler
// 					}
// 				}
// 			}

// 			req, _ := http.NewRequest("GET", "/list_channels", nil)
// 			rw := httptest.NewRecorder()

// 			handler(context.Background(), rw, req)

// 			var resp response
// 			{
// 				if rw.Code != tc.wantCode {
// 					t.Errorf("unexpected response code. want: %d, got: %d", tc.wantCode, rw.Code)
// 				}

// 				if rw.Code == http.StatusUnauthorized {
// 					return
// 				}

// 				respBody(t, rw.Body, &resp)
// 				if tc.wantErr != (resp.Errors != nil) {
// 					t.Errorf("unexpected err response. want: %v, got: %+v", tc.wantErr, resp.Errors)
// 					return
// 				}

// 				if rw.Code != http.StatusOK {
// 					return
// 				}

// 				var got []string
// 				json.Unmarshal(resp.Data, &got)
// 				if !reflect.DeepEqual(tc.want, got) {
// 					t.Errorf("unexpected response. want: %v, got: %+v", tc.want, got)
// 				}

// 			}
// 		})
// 	}
// }

// func reqBody(t *testing.T, i interface{}) io.Reader {
// 	data, err := json.Marshal(i)
// 	if err != nil {
// 		t.Fatalf("json encode err: %v", err)
// 	}
// 	return bytes.NewReader(data)
// }

// func respBody(t *testing.T, r io.Reader, v interface{}) {
// 	err := json.NewDecoder(r).Decode(v)
// 	if err != nil {
// 		t.Fatalf("json encode err: %v", err)
// 	}
// }

type store struct {
	SaveFunc      func(*goch.Chat) error
	GetFunc       func(string) (*goch.Chat, error)
	ListChansFunc func() ([]string, error)
}

func (s *store) Save(c *goch.Chat) error              { return s.SaveFunc(c) }
func (s *store) Get(id string) (*goch.Chat, error)    { return s.GetFunc(id) }
func (s *store) ListChannels() ([]string, error)      { return s.ListChansFunc() }
func (s *store) GetUnreadCount(string, string) uint64 { panic("not implemented") }
