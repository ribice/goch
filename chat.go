package goch

import (
	"errors"
	"fmt"

	"github.com/rs/xid"
	"github.com/vmihailenco/msgpack"
)

// NewChannel creates new channel chat
func NewChannel(name string, private bool) *Chat {
	ch := Chat{
		Name:    name,
		Members: make(map[string]*User),
	}

	if private {
		ch.Secret = newSecret()
	}

	return &ch
}

// Chat represents private or channel chat
type Chat struct {
	Name    string           `json:"name"`
	Secret  string           `json:"secret"`
	Members map[string]*User `json:"members"`
}

// Chat errors
var (
	errAlreadyRegistered = errors.New("chat: uid already registered in this chat")
	errNotRegistered     = errors.New("chat: not a member of this channel")
	errInvalidSecret     = errors.New("chat: invalid secret")
)

// Register registers user with a chat and returns secret which should
// be stored on the client side, and used for subsequent join requests
func (c *Chat) Register(u *User) (string, error) {
	if _, ok := c.Members[u.UID]; ok {
		return "", errAlreadyRegistered
	}
	if u.Secret == "" {
		u.Secret = newSecret()
	}
	c.Members[u.UID] = u
	return u.Secret, nil
}

// Join attempts to join user to chat
func (c *Chat) Join(uid, secret string) (*User, error) {
	u, ok := c.Members[uid]
	if !ok {
		return nil, errNotRegistered
	}
	if u.Secret != secret {
		return nil, errInvalidSecret
	}
	u.Secret = ""
	return u, nil
}

// Leave removes user from channel
func (c *Chat) Leave(uid string) {
	delete(c.Members, uid)
}

func newSecret() string {
	return xid.New().String()
}

// ListMembers returns list of members associated to a chat
func (c *Chat) ListMembers() []*User {
	if len(c.Members) < 1 {
		return nil
	}
	var members []*User
	for _, u := range c.Members {
		u.Secret = ""
		members = append(members, u)
	}
	return members
}

// DecodeChat tries to decode binary formatted message in b to Message
func DecodeChat(b string) (*Chat, error) {
	var c Chat
	if err := msgpack.Unmarshal([]byte(b), &c); err != nil {
		return nil, fmt.Errorf("client: unable to unmarshal chat: %v", err)
	}
	return &c, nil
}

// Encode encodes provided chat in binary format
func (c *Chat) Encode() ([]byte, error) {
	return msgpack.Marshal(c)
}

// TODO: Private chats (can only init private chat with people in the same channel)
