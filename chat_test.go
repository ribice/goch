package goch_test

import (
	"reflect"
	"testing"

	"github.com/ribice/goch"
)

func TestNewChannel(t *testing.T) {
	channel := "channelName"
	ch := goch.NewChannel(channel, true)
	if ch.Secret == "" {
		t.Error("expected channel to have secret but does not")
	}
	if ch.Name != channel {
		t.Error("invalid channel name")
	}
}

func TestRegister(t *testing.T) {
	cases := []struct {
		name    string
		c       *goch.Chat
		req     *goch.User
		wantErr string
	}{
		{
			name: "User already registered",
			c: &goch.Chat{
				Members: map[string]*goch.User{
					"ABC": &goch.User{},
				},
			},
			req:     &goch.User{UID: "ABC"},
			wantErr: "chat: uid already registered in this chat",
		},
		{
			name: "User with secret",
			c: &goch.Chat{
				Members: map[string]*goch.User{
					"ABC": &goch.User{Secret: "secret2"},
				},
			},
			req: &goch.User{UID: "DAF", Secret: "Secret"},
		},
		{
			name: "User without secret",
			c: &goch.Chat{
				Members: map[string]*goch.User{
					"ABC": &goch.User{},
				},
			},
			req: &goch.User{UID: "DAF"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.c == nil {
				t.Error("Chat has to be instantiated")
			}
			secret, err := tc.c.Register(tc.req)
			if err != nil && tc.wantErr != err.Error() {
				t.Errorf("expected err %s but got %s", tc.wantErr, err.Error())
			}

			if tc.wantErr == "" {
				if tc.req.Secret != "" {
					if secret != tc.req.Secret {
						t.Errorf("expected secret %s but got %s", tc.req.Secret, secret)
					}
				} else if len(secret) != 20 {
					t.Errorf("expected len to be 20 but got %v", len(secret))
				}
			}

		})
	}
}

func TestJoin(t *testing.T) {
	cases := []struct {
		name    string
		c       *goch.Chat
		uid     string
		secret  string
		want    *goch.User
		wantErr string
	}{
		{
			name: "User not registered",
			c: &goch.Chat{
				Members: map[string]*goch.User{
					"ABC": &goch.User{},
				},
			},
			uid:     "DFA",
			wantErr: "chat: not a member of this channel",
		},
		{
			name: "Invalid secret",
			c: &goch.Chat{
				Members: map[string]*goch.User{
					"ABC": &goch.User{Secret: "secret2"},
				},
			},
			uid:     "ABC",
			secret:  "secret1",
			wantErr: "chat: invalid secret",
		},
		{
			name: "Success",
			c: &goch.Chat{
				Members: map[string]*goch.User{
					"ABC": &goch.User{Secret: "secret1", DisplayName: "John", UID: "ABC", Email: "john@doe.com"},
				},
			},
			uid:    "ABC",
			secret: "secret1",
			want:   &goch.User{DisplayName: "John", UID: "ABC", Email: "john@doe.com"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.c == nil {
				t.Error("Chat has to be instantiated")
			}
			user, err := tc.c.Join(tc.uid, tc.secret)
			if err != nil && tc.wantErr != err.Error() {
				t.Errorf("expected err %s but got %s", tc.wantErr, err.Error())
			}

			if tc.want != nil && !reflect.DeepEqual(tc.want, user) {
				t.Errorf("expected user %v but got %v", tc.want, user)
			}
		})
	}
}

func TestLeave(t *testing.T) {
	c := &goch.Chat{
		Members: map[string]*goch.User{
			"user1": &goch.User{},
		},
	}
	c.Leave("user1")
	if len(c.Members) > 0 {
		t.Errorf("expected 0 mumbers, but found: %v", len(c.Members))
	}
}

func TestMembers(t *testing.T) {
	c := &goch.Chat{}
	mems := c.ListMembers()
	if len(mems) > 0 {
		t.Errorf("expected 0 members, but got %v", len(mems))
	}
	c.Members = map[string]*goch.User{
		"User1": &goch.User{Secret: "secret1"},
		"User2": &goch.User{Secret: "secret2"},
	}
	newMems := c.ListMembers()
	if len(newMems) != len(c.Members) {
		t.Errorf("expected %v members, but got %v", len(c.Members), len(newMems))
	}
	for _, v := range newMems {
		if v.Secret != "" {
			t.Error("expected secret to be empty but was not")
		}
	}
}

func TestChatEncode(t *testing.T) {
	c := &goch.Chat{Name: "msgPack", Secret: "packMsg",
		Members: map[string]*goch.User{
			"User1": &goch.User{UID: "User1"},
		}}
	bts, err := c.Encode()
	if err != nil {
		t.Errorf("did not expect error but received: %v", err)
	}
	ch, err := goch.DecodeChat(string(bts))
	if err != nil {
		t.Errorf("did not expect error but received: %v", err)
	}
	if !reflect.DeepEqual(c, ch) {
		t.Errorf("expected chat %v but got %v", c, ch)
	}

	_, err = goch.DecodeChat("test")
	if err == nil {
		t.Error("expected error but received nil")
	}
}
