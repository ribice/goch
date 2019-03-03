package chat_test

// import (
// 	"bytes"
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"io"
// 	"net/http"
// 	"net/http/httptest"
// 	"reflect"
// 	"testing"

// 	"github.com/tonto/gossip/pkg/chat"
// 	h "github.com/tonto/kit/http"
// )

// type response struct {
// 	Code   int             `json:"code"`
// 	Data   json.RawMessage `json:"data,omitempty"`
// 	Errors []string        `json:"errors,omitempty"`
// }

// type createChanReq struct {
// 	Name    string `json:"name"`
// 	Private bool   `json:"private"`
// }

// type createChanResp struct {
// 	Secret string `json:"secret"`
// }

// func TestCreateChannel(t *testing.T) {
// 	cases := []struct {
// 		name     string
// 		username string
// 		password string
// 		store    *store
// 		req      createChanReq
// 		want     *createChanResp
// 		wantErr  bool
// 		wantCode int
// 	}{
// 		{
// 			name:     "test name req validation",
// 			req:      createChanReq{Private: false},
// 			username: "admin",
// 			password: "test",
// 			wantErr:  true,
// 			wantCode: http.StatusBadRequest,
// 		},
// 		{
// 			name:     "test name length validation short",
// 			req:      createChanReq{Name: "a"},
// 			username: "admin",
// 			password: "test",
// 			wantErr:  true,
// 			wantCode: http.StatusBadRequest,
// 		},
// 		{
// 			name:     "test name length validation long",
// 			req:      createChanReq{Name: "qwertyuiopasdfghjklzxcvbnk"},
// 			username: "admin",
// 			password: "test",
// 			wantErr:  true,
// 			wantCode: http.StatusBadRequest,
// 		},
// 		{
// 			name:     "test name alphanumeric",
// 			req:      createChanReq{Name: "ak ; )___"},
// 			username: "admin",
// 			password: "test",
// 			wantErr:  true,
// 			wantCode: http.StatusBadRequest,
// 		},
// 		{
// 			name:     "test invalid admin creds",
// 			req:      createChanReq{Name: "ak ; )___"},
// 			username: "admin",
// 			password: "admin",
// 			wantErr:  true,
// 			wantCode: http.StatusUnauthorized,
// 		},
// 		{
// 			store: &store{
// 				SaveFunc: func(c *chat.Chat) error {
// 					return nil
// 				},
// 			},
// 			name:     "test create public",
// 			req:      createChanReq{Name: "general"},
// 			username: "admin",
// 			password: "test",
// 			wantErr:  false,
// 			wantCode: http.StatusOK,
// 			want:     &createChanResp{Secret: ""},
// 		},
// 		{
// 			store: &store{
// 				SaveFunc: func(c *chat.Chat) error {
// 					return nil
// 				},
// 			},
// 			name:     "test create private",
// 			req:      createChanReq{Name: "general", Private: true},
// 			username: "admin",
// 			password: "test",
// 			wantErr:  false,
// 			wantCode: http.StatusOK,
// 		},
// 		{
// 			store: &store{
// 				SaveFunc: func(c *chat.Chat) error {
// 					return fmt.Errorf("could not store channel")
// 				},
// 			},
// 			name:     "test store error",
// 			req:      createChanReq{Name: "general"},
// 			username: "admin",
// 			password: "test",
// 			wantErr:  true,
// 			wantCode: http.StatusInternalServerError,
// 		},
// 	}

// 	for _, tc := range cases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			var handler h.HandlerFunc
// 			{
// 				api := chat.NewAPI(tc.store, "admin", "test")
// 				api.Prefix() // only for coverage
// 				for path, ep := range api.Endpoints() {
// 					if path == "/admin/create_channel" {
// 						handler = ep.Handler
// 					}
// 				}
// 			}

// 			req, _ := http.NewRequest("POST", "/admin/create_channel", reqBody(t, tc.req))
// 			req.SetBasicAuth(tc.username, tc.password)
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
// 				if tc.want != nil {
// 					var got createChanResp
// 					json.Unmarshal(resp.Data, &got)
// 					if !reflect.DeepEqual(got, *tc.want) {
// 						t.Errorf("unexpected response. want: %+v, got: %+v", *tc.want, got)
// 					}
// 				}
// 				if tc.wantErr != (resp.Errors != nil) {
// 					t.Errorf("unexpected err response. want: %v, got: %+v", tc.wantErr, resp.Errors)
// 				}
// 			}
// 		})
// 	}
// }

// type registerNickReq struct {
// 	Nick          string `json:"nick"`
// 	FullName      string `json:"name"`
// 	Email         string `json:"email"`
// 	Secret        string `json:"secret"`
// 	Channel       string `json:"channel"`
// 	ChannelSecret string `json:"channel_secret"` // Tennant
// }

// type registerNickResp struct {
// 	Secret string `json:"secret"`
// }

// func TestRegisterNick(t *testing.T) {
// 	cases := []struct {
// 		name     string
// 		store    *store
// 		req      registerNickReq
// 		wantErr  bool
// 		wantCode int
// 		want     string
// 	}{
// 		{
// 			name:     "test req channel validation",
// 			req:      registerNickReq{Nick: "foo"},
// 			wantErr:  true,
// 			wantCode: http.StatusBadRequest,
// 		},
// 		{
// 			name:     "test req nick validation",
// 			req:      registerNickReq{Channel: "foo"},
// 			wantErr:  true,
// 			wantCode: http.StatusBadRequest,
// 		},
// 		{
// 			name:     "test nick short",
// 			req:      registerNickReq{Nick: "jo", Channel: "foo"},
// 			wantErr:  true,
// 			wantCode: http.StatusBadRequest,
// 		},
// 		{
// 			name:     "test nick long",
// 			req:      registerNickReq{Nick: "joefokjdislijflskdjfh", Channel: "foo"},
// 			wantErr:  true,
// 			wantCode: http.StatusBadRequest,
// 		},
// 		{
// 			name: "test nick long",
// 			req: registerNickReq{
// 				Nick:          "joe123",
// 				Channel:       "foobar",
// 				ChannelSecret: "1123456789012345678901234567890123456789012345678901234567890234567890",
// 			},
// 			wantErr:  true,
// 			wantCode: http.StatusBadRequest,
// 		},
// 		{
// 			name: "test fields too long",
// 			req: registerNickReq{
// 				Nick:          "joe",
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
// 			req:      registerNickReq{Nick: " ;' joe", Channel: "foo"},
// 			wantErr:  true,
// 			wantCode: http.StatusBadRequest,
// 		},
// 		{
// 			store: &store{
// 				GetFunc: func(id string) (*chat.Chat, error) {
// 					return nil, fmt.Errorf("err fetching chan")
// 				},
// 			},
// 			name:     "test err fetch chan",
// 			req:      registerNickReq{Nick: "joe", Channel: "foo"},
// 			wantErr:  true,
// 			wantCode: http.StatusInternalServerError,
// 		},
// 		{
// 			store: &store{
// 				GetFunc: func(id string) (*chat.Chat, error) {
// 					return &chat.Chat{Secret: "foo"}, nil
// 				},
// 			},
// 			name:     "test invalid secret",
// 			req:      registerNickReq{Nick: "joe", Channel: "foo"},
// 			wantErr:  true,
// 			wantCode: http.StatusInternalServerError,
// 		},
// 		{
// 			store: &store{
// 				GetFunc: func(id string) (*chat.Chat, error) {
// 					return &chat.Chat{Secret: "", Members: map[string]chat.User{"joe": {Nick: "joe"}}}, nil
// 				},
// 			},
// 			name:     "test nick exists",
// 			req:      registerNickReq{Nick: "joe", Channel: "foo"},
// 			wantErr:  true,
// 			wantCode: http.StatusInternalServerError,
// 		},
// 		{
// 			store: &store{
// 				GetFunc: func(id string) (*chat.Chat, error) {
// 					ch := chat.NewChannel("foo", false)
// 					ch.Secret = "xxxyyy"
// 					return ch, nil
// 				},
// 				SaveFunc: func(ch *chat.Chat) error {
// 					return fmt.Errorf("unable to save")
// 				},
// 			},
// 			name:     "test save failed",
// 			req:      registerNickReq{Nick: "joe", Channel: "foo", ChannelSecret: "xxxyyy"},
// 			wantErr:  true,
// 			wantCode: http.StatusInternalServerError,
// 		},
// 		{
// 			store: &store{
// 				GetFunc: func(id string) (*chat.Chat, error) {
// 					ch := chat.NewChannel("foo", false)
// 					ch.Secret = "xxxyyy"
// 					return ch, nil
// 				},
// 				SaveFunc: func(ch *chat.Chat) error { return nil },
// 			},
// 			name:     "test saved",
// 			req:      registerNickReq{Nick: "joe", Channel: "foo", ChannelSecret: "xxxyyy"},
// 			wantErr:  false,
// 			wantCode: http.StatusOK,
// 		},
// 		{
// 			store: &store{
// 				GetFunc: func(id string) (*chat.Chat, error) {
// 					ch := chat.NewChannel("foo", false)
// 					ch.Secret = "xxxyyy"
// 					return ch, nil
// 				},
// 				SaveFunc: func(ch *chat.Chat) error { return nil },
// 			},
// 			name:     "test provided secret length",
// 			req:      registerNickReq{Nick: "joe", Channel: "foo", Secret: "qwertyuiopasdfghjklmnbvcxzhdguui", ChannelSecret: "xxxyyy"},
// 			wantErr:  true,
// 			wantCode: http.StatusBadRequest,
// 		},
// 		{
// 			store: &store{
// 				GetFunc: func(id string) (*chat.Chat, error) {
// 					ch := chat.NewChannel("foo", false)
// 					ch.Secret = "xxxyyy"
// 					return ch, nil
// 				},
// 				SaveFunc: func(ch *chat.Chat) error { return nil },
// 			},
// 			name:     "test provided secret alphanumeric",
// 			req:      registerNickReq{Nick: "joe", Channel: "foo", Secret: "asljfkd ' ';", ChannelSecret: "xxxyyy"},
// 			wantErr:  true,
// 			wantCode: http.StatusBadRequest,
// 		},
// 		{
// 			store: &store{
// 				GetFunc: func(id string) (*chat.Chat, error) {
// 					ch := chat.NewChannel("foo", false)
// 					ch.Secret = "xxxyyy"
// 					return ch, nil
// 				},
// 				SaveFunc: func(ch *chat.Chat) error { return nil },
// 			},
// 			name:     "test saved with provided secret",
// 			req:      registerNickReq{Nick: "joe", Channel: "foo", Secret: "foobarbaz", ChannelSecret: "xxxyyy"},
// 			wantErr:  false,
// 			wantCode: http.StatusOK,
// 		},

// 		// TODO - Test server username/pass (empty/nonempty)
// 	}

// 	for _, tc := range cases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			var handler h.HandlerFunc
// 			{
// 				api := chat.NewAPI(tc.store, "admin", "test")
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

// 				var got registerNickResp
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
// 		want     []chat.User
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
// 				GetFunc: func(id string) (*chat.Chat, error) {
// 					return nil, fmt.Errorf("err fetching chan")
// 				},
// 			},
// 			name:     "test err fetch chan",
// 			req:      channelMembersReq{Channel: "foo"},
// 			wantErr:  true,
// 			wantCode: http.StatusInternalServerError,
// 		},
// 		{
// 			store: &store{
// 				GetFunc: func(id string) (*chat.Chat, error) {
// 					return &chat.Chat{
// 						Secret: "",
// 						Members: map[string]chat.User{
// 							"joe": {
// 								Nick: "joe",
// 							},
// 							"foo": {
// 								Nick: "foo",
// 							},
// 							"bar": {
// 								Nick: "bar",
// 							},
// 							"baz": {
// 								Nick: "baz",
// 							},
// 						},
// 					}, nil
// 				},
// 			},
// 			name:    "test success",
// 			req:     channelMembersReq{Channel: "foo"},
// 			wantErr: false,
// 			want: []chat.User{
// 				{
// 					Nick: "joe",
// 				},
// 				{
// 					Nick: "foo",
// 				},
// 				{
// 					Nick: "bar",
// 				},
// 				{
// 					Nick: "baz",
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
// 				api := chat.NewAPI(tc.store, "admin", "test")
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

// 				var got []chat.User
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
// 					return nil, fmt.Errorf("err fetching chan")
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
// 				api := chat.NewAPI(tc.store, "admin", "test")
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

// type store struct {
// 	SaveFunc      func(*chat.Chat) error
// 	GetFunc       func(string) (*chat.Chat, error)
// 	ListChansFunc func() ([]string, error)
// }

// func (s *store) Save(c *chat.Chat) error              { return s.SaveFunc(c) }
// func (s *store) Get(id string) (*chat.Chat, error)    { return s.GetFunc(id) }
// func (s *store) ListChannels() ([]string, error)      { return s.ListChansFunc() }
// func (s *store) GetUnreadCount(string, string) uint64 { panic("not implemented") }
