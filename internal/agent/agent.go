package agent

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/ribice/goch"

	"github.com/gorilla/websocket"
)

// New creates new connection agent instance
func New(mb MessageBroker, store ChatStore) *Agent {
	return &Agent{
		mb:    mb,
		store: store,
		done:  make(chan struct{}, 1),
	}
}

// Agent represents chat connection agent which handles end to end comm client - broker
type Agent struct {
	chat        *goch.Chat
	uid         string
	displayName string
	done        chan struct{}
	closeSub    func()
	closed      bool

	conn *websocket.Conn
	mb   MessageBroker

	store ChatStore
}

// ChatStore represents chat store interface
type ChatStore interface {
	Get(string) (*goch.Chat, error)
	GetRecent(string, int64) ([]goch.Message, uint64, error)
	UpdateLastClientSeq(string, string, uint64)
}

// MessageBroker represents broker interface
type MessageBroker interface {
	Subscribe(string, string, uint64, chan *goch.Message) (func(), error)
	SubscribeNew(string, string, chan *goch.Message) (func(), error)
	Send(string, *goch.Message) error
}

type msgT int

const (
	chatMsg msgT = iota
	historyMsg
	errorMsg
	infoMsg
	historyReqMsg
)

const (
	maxHistoryCount uint64 = 512
)

type msg struct {
	Type  msgT        `json:"type"`
	Data  interface{} `json:"data,omitempty"`
	Error string      `json:"error,omitempty"`
}

// HandleConn handles websocket communication for requested chat/client
func (a *Agent) HandleConn(conn *websocket.Conn, req *initConReq) {
	a.conn = conn

	a.conn.SetCloseHandler(func(code int, text string) error {
		a.closed = true
		a.done <- struct{}{}
		return nil
	})

	ct, err := a.store.Get(req.Channel)
	if err != nil {
		writeFatal(a.conn, fmt.Sprintf("agent: unable to find chat: %v", err))
		return
	}

	// if ct == nil {
	// 	writeFatal(a.conn, "agent: this chat does not exist")
	// 	return
	// }

	user, err := ct.Join(req.UID, req.Secret)
	if err != nil {
		writeFatal(a.conn, fmt.Sprintf("agent: unable to join chat: %v", err))
		return
	}

	a.chat = ct
	a.setUser(user)

	mc := make(chan *goch.Message)
	{
		var close func()

		if req.LastSeq != nil {
			close, err = a.mb.Subscribe(req.Channel, user.UID, *req.LastSeq, mc)
		} else if seq, err := a.pushRecent(); err != nil {
			writeErr(a.conn, fmt.Sprintf("agent: unable to fetch chat history: %v", err))
			close, err = a.mb.SubscribeNew(req.Channel, user.UID, mc)
		} else {
			close, err = a.mb.Subscribe(req.Channel, user.UID, seq, mc)
		}

		if err != nil {
			writeFatal(a.conn, fmt.Sprintf("agent: unable to subscribe to chat updates due to: %v. closing connection", err))
			return
		}

		a.closeSub = close
	}

	a.loop(mc)
}

func (a *Agent) pushRecent() (uint64, error) {
	msgs, seq, err := a.store.GetRecent(a.chat.Name, 100)
	if err != nil {
		return 0, err
	}

	if msgs == nil {
		return 0, nil
	}

	a.store.UpdateLastClientSeq(a.uid, a.chat.Name, msgs[len(msgs)-1].Seq)

	return seq, a.conn.WriteJSON(msg{
		Type: historyMsg,
		Data: msgs,
	})

}

func (a *Agent) loop(mc chan *goch.Message) {
	go func() {
		for {
			if a.closed {
				return
			}

			_, r, err := a.conn.NextReader()
			if err != nil {
				writeErr(a.conn, err.Error())
				continue
			}

			a.handleClientMsg(r)
		}
	}()

	go func() {
		defer a.closeSub()
		defer a.conn.Close()
		for {
			select {
			case m := <-mc:
				a.conn.WriteJSON(msg{
					Type: chatMsg,
					Data: m,
				})

				a.store.UpdateLastClientSeq(a.uid, a.chat.Name, m.Seq)
			case <-a.done:
				return
			}
		}
	}()
}

func (a *Agent) handleClientMsg(r io.Reader) {
	var message struct {
		Type msgT            `json:"type"`
		Data json.RawMessage `json:"data,omitempty"`
	}

	err := json.NewDecoder(r).Decode(&message)
	if err != nil {
		writeErr(a.conn, fmt.Sprintf("invalid message format: %v", err))
		return
	}

	switch message.Type {
	case chatMsg:
		a.handleChatMsg(message.Data)
	case historyReqMsg:
		a.handleHistoryReqMsg(message.Data)
	}
}

type message struct {
	Meta map[string]string `json:"meta"`
	Seq  uint64            `json:"seq"`
	Text string            `json:"text"`
}

func (a *Agent) handleChatMsg(raw json.RawMessage) {
	var msg message

	err := json.Unmarshal(raw, &msg)
	if err != nil {
		writeErr(a.conn, fmt.Sprintf("invalid text message format: %v", err))
		return
	}

	if msg.Text == "" {
		writeErr(a.conn, "sent empty message")
		return
	}

	if len(msg.Text) > 1024 {
		writeErr(a.conn, "exceeded max message length of 1024 characters")
		return
	}

	err = a.mb.Send(a.chat.Name, &goch.Message{
		Meta:     msg.Meta,
		Text:     msg.Text,
		Seq:      msg.Seq,
		FromName: a.displayName,
		FromUID:  a.uid,
		Time:     time.Now().UnixNano(),
	})
	if err != nil {
		writeErr(a.conn, fmt.Sprintf("could not forward your message. try again: %v", err))
	}
}

func (a *Agent) handleHistoryReqMsg(raw json.RawMessage) {
	var req struct {
		To uint64 `json:"to"`
	}

	err := json.Unmarshal(raw, &req)
	if err != nil {
		writeErr(a.conn, fmt.Sprintf("invalid history request message format: %v", err))
		return
	}

	if req.To <= 0 {
		return
	}

	msgs, err := a.buildHistoryBatch(req.To)
	if err != nil {
		writeErr(a.conn, fmt.Sprintf("could not fetch chat history: %v", err))
		return
	}

	if err := a.conn.WriteJSON(msg{
		Type: historyMsg,
		Data: msgs,
	}); err != nil {
		writeErr(a.conn, fmt.Sprintf("could not write message: %v", err))
	}
}

func (a *Agent) buildHistoryBatch(to uint64) ([]*goch.Message, error) {
	var offset uint64

	if to >= maxHistoryCount {
		offset = to - maxHistoryCount
	}

	mc := make(chan *goch.Message)

	close, err := a.mb.Subscribe(a.chat.Name, "", offset, mc)
	if err != nil {
		return nil, err
	}

	defer close()

	var msgs []*goch.Message

	for {
		msg := <-mc
		if msg.Seq >= to {
			break
		}
		msgs = append(msgs, msg)
	}

	return msgs, nil
}

func writeErr(conn *websocket.Conn, err string) {
	conn.WriteJSON(msg{Error: err, Type: errorMsg})
}

func writeFatal(conn *websocket.Conn, err string) {
	conn.WriteJSON(msg{Error: err, Type: errorMsg})
	conn.Close()
}

func (a *Agent) setUser(u *goch.User) {
	a.uid = u.UID
	a.displayName = u.DisplayName
}
