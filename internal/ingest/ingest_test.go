package ingest_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/ribice/goch"

	"github.com/ribice/goch/internal/ingest"
)

var errTest = errors.New("error")

func TestChatIngest(t *testing.T) {
	cases := []struct {
		name    string
		chat    string
		n       int
		wantErr bool
	}{
		{
			name:    "empty queue",
			chat:    "general",
			n:       0,
			wantErr: false,
		},
		{
			name:    "100 messages",
			chat:    "general",
			n:       100,
			wantErr: false,
		},
		{
			name:    "1000 messages",
			chat:    "general",
			n:       1000,
			wantErr: false,
		},
		{
			name:    "test err",
			chat:    "general",
			n:       1000,
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			q := queue{err: tc.wantErr}
			s := store{}

			for i := 0; i < tc.n; i++ {
				msg := &goch.Message{
					Text: fmt.Sprintf("msg number %d", i),
				}
				bts, err := msg.Encode()
				if err != nil {
					t.Fatal(err)
				}
				q.data = append(
					q.data,
					struct {
						seq uint64
						msg []byte
					}{
						seq: uint64(i),
						msg: bts,
					},
				)
			}

			ig := ingest.New(
				&q,
				&s,
			)

			close, err := ig.Run(tc.chat)
			if (err != nil) != tc.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if err != nil {
				return
			}

			defer close()

			<-q.purged

			time.Sleep(100 * time.Millisecond)

			if len(s.data[tc.chat]) != tc.n {
				t.Fatalf("messages not received by ingester, want: %d, got: %d", tc.n, len(s.data[tc.chat]))
			}

			for i, m := range s.data[tc.chat] {
				if m.Text != fmt.Sprintf("msg number %d", i) && m.Seq != uint64(i) {
					t.Fatalf("message not received by ingester: %d", i)
				}
			}
		})
	}
}

func TestChatIngestDecodingErrs(t *testing.T) {
	cases := []struct {
		name string
		chat string
		n    int
	}{
		{
			name: "test 100",
			chat: "general",
			n:    100,
		},

		// TODO - test AppendMessage error handling
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			q := queue{}
			s := store{}

			for i := 0; i < tc.n; i++ {
				d, err := json.Marshal(goch.Message{Text: "foo bar"})
				if err != nil {
					t.Fatal(err)
				}
				q.data = append(
					q.data,
					struct {
						seq uint64
						msg []byte
					}{
						seq: uint64(i),
						msg: d,
					},
				)
			}

			ig := ingest.New(
				&q,
				&s,
			)

			close, err := ig.Run(tc.chat)
			if err != nil {
				t.Fatal(err)
			}

			defer close()

			<-q.purged

			time.Sleep(100 * time.Millisecond)

			if len(s.data[tc.chat]) != tc.n {
				t.Fatalf("messages not received by ingester, want: %d, got: %d", tc.n, len(s.data[tc.chat]))
			}

			for i, m := range s.data[tc.chat] {
				if m.Text != "ingest: message unavailable: decoding error" && m.Seq != uint64(i) {
					t.Fatalf("message not received by ingester: %d", i)
				}
			}
		})
	}
}

type store struct {
	data map[string][]*goch.Message
	err  bool
}

func (s *store) AppendMessage(id string, msg *goch.Message) error {
	if s.data == nil {
		s.data = make(map[string][]*goch.Message)
	}
	s.data[id] = append(s.data[id], msg)
	if s.err {
		return errTest
	}
	return nil
}

type queue struct {
	data []struct {
		seq uint64
		msg []byte
	}
	purged chan struct{}
	err    bool
}

func (q *queue) SubscribeQueue(id string, f func(uint64, []byte)) (io.Closer, error) {
	if q.err {
		return nil, errTest
	}
	q.purged = make(chan struct{})
	go func() {
		for _, m := range q.data {
			d := m.msg
			f(m.seq, d)
		}
		q.purged <- struct{}{}
	}()
	return &cl{}, nil
}

type cl struct{}

func (c *cl) Close() error { return nil }
