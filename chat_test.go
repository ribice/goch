package goch_test

// import (
// 	"reflect"
// 	"testing"

// 	"github.com/tonto/gossip/pkg/chat"
// )

// func TestChannelRegister(t *testing.T) {
// 	cases := []struct {
// 		name    string
// 		channel string
// 		private bool
// 		users   []chat.User
// 		secret  string
// 		wantErr bool
// 	}{
// 		{
// 			name:    "test register single user public chan",
// 			channel: "general",
// 			private: false,
// 			users: []chat.User{
// 				{
// 					Nick:     "foo",
// 					FullName: "0",
// 					Email:    "john@email.com",
// 				},
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			name:    "test register single user custom secret",
// 			channel: "general",
// 			private: false,
// 			secret:  "xxx-yyy-zzz",
// 			users: []chat.User{
// 				{
// 					Nick:     "foo",
// 					FullName: "0",
// 					Email:    "john@email.com",
// 				},
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			name:    "test register single user private chan",
// 			channel: "general",
// 			private: true,
// 			users: []chat.User{
// 				{
// 					Nick:     "foo",
// 					FullName: "0",
// 					Email:    "john@email.com",
// 				},
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			name:    "test register multiple users private chan",
// 			channel: "general",
// 			private: true,
// 			users: []chat.User{
// 				{
// 					Nick:     "foo",
// 					FullName: "0",
// 					Email:    "john@email.com",
// 				},
// 				{
// 					Nick:     "bar",
// 					FullName: "1",
// 					Email:    "john@email.com",
// 				},
// 				{
// 					Nick:     "baz",
// 					FullName: "2",
// 					Email:    "john@email.com",
// 				},
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			name:    "test register multiple users nick exists err",
// 			channel: "general",
// 			private: true,
// 			users: []chat.User{
// 				{
// 					Nick:     "foo",
// 					FullName: "0",
// 					Email:    "john@email.com",
// 				},
// 				{
// 					Nick:     "bar",
// 					FullName: "1",
// 					Email:    "john@email.com",
// 				},
// 				{
// 					Nick:     "foo",
// 					FullName: "2",
// 					Email:    "john@email.com",
// 				},
// 			},
// 			wantErr: true,
// 		},
// 	}

// 	for _, tc := range cases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			ch := chat.NewChannel("general", tc.private)
// 			if tc.private && ch.Secret == "" {
// 				t.Errorf("secret should not be empty for private channels")
// 			}

// 			var e error

// 			for i := range tc.users {
// 				secret, err := ch.Register(&tc.users[i], tc.secret)
// 				if err != nil {
// 					e = err
// 					continue
// 				}

// 				if tc.secret != "" && tc.secret != secret {
// 					t.Errorf("custom secret not set")
// 				}
// 			}

// 			if (e != nil) != tc.wantErr {
// 				t.Fatalf("error = %v, wantErr %v", e, tc.wantErr)
// 				return
// 			}

// 			if tc.wantErr {
// 				return
// 			}

// 			users := []chat.User{}

// 			for _, m := range ch.Members {
// 				users = append(users, m)
// 			}

// 			for _, w := range tc.users {
// 				found := false
// 				for _, g := range users {
// 					if w.Nick == g.Nick {
// 						found = true
// 					}
// 				}
// 				if !found {
// 					t.Errorf("Members = %v, want %v", users, tc.users)
// 				}
// 			}
// 		})
// 	}
// }

// func TestChannelJoin(t *testing.T) {
// 	cases := []struct {
// 		name    string
// 		chat    chat.Chat
// 		nick    string
// 		secret  string
// 		want    chat.User
// 		wantErr bool
// 	}{
// 		{
// 			name: "test join chat",
// 			chat: chat.Chat{
// 				Secret: "",
// 				Members: map[string]chat.User{
// 					"foo": chat.User{Nick: "foo", Secret: "123-fa6"},
// 				},
// 			},
// 			nick:    "foo",
// 			secret:  "123-fa6",
// 			want:    chat.User{Nick: "foo"},
// 			wantErr: false,
// 		},
// 		{
// 			name: "test join chat multiple users",
// 			chat: chat.Chat{
// 				Secret: "xxx",
// 				Members: map[string]chat.User{
// 					"foo": chat.User{Nick: "foo", Secret: "453-fa6"},
// 					"bar": chat.User{Nick: "bar", Secret: "137-fa6"},
// 					"baz": chat.User{Nick: "baz", Secret: "123-fa6"},
// 				},
// 			},
// 			nick:    "baz",
// 			secret:  "123-fa6",
// 			want:    chat.User{Nick: "baz"},
// 			wantErr: false,
// 		},
// 		{
// 			name: "test invalid secret",
// 			chat: chat.Chat{
// 				Secret: "",
// 				Members: map[string]chat.User{
// 					"foo": chat.User{Nick: "foo", Secret: "123-fa6"},
// 				},
// 			},
// 			nick:    "foo",
// 			secret:  "123-fa6x",
// 			want:    chat.User{Nick: "foo"},
// 			wantErr: true,
// 		},
// 		{
// 			name: "test secret nick missmatch",
// 			chat: chat.Chat{
// 				Secret: "",
// 				Members: map[string]chat.User{
// 					"bar": chat.User{Nick: "bar", Secret: "xxxx"},
// 				},
// 			},
// 			nick:    "foo",
// 			secret:  "123-fa6",
// 			want:    chat.User{Nick: "foo"},
// 			wantErr: true,
// 		},
// 	}

// 	for _, tc := range cases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			user, err := tc.chat.Join(tc.nick, tc.secret)
// 			if (err != nil) != tc.wantErr {
// 				t.Errorf("err fail. want: %v, got: %v", tc.wantErr, err)
// 			}

// 			if !tc.wantErr {
// 				if !reflect.DeepEqual(&tc.want, user) {
// 					t.Errorf("user fail. want: %v, got: %v", tc.want, user)
// 				}
// 			}
// 		})
// 	}
// }
