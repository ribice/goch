// Package ingest provides functionality for
// updating per chat read models (recent history)
package ingest

import (
	"fmt"
	"io"
	"time"

	"github.com/ribice/goch"
)

// New creates new ingest instance
func New(mq MQ, s ChatStore) *Ingest {
	return &Ingest{
		mq:    mq,
		store: s,
	}
}

// Ingest represents chat ingester
type Ingest struct {
	mq    MQ
	store ChatStore
}

// MQ represents ingest message queue interface
type MQ interface {
	SubscribeQueue(string, func(uint64, []byte)) (io.Closer, error)
}

// ChatStore represents chat store interface
type ChatStore interface {
	AppendMessage(string, *goch.Message) error
}

// Run subscribes to ingest queue group and updates chat read model
func (i *Ingest) Run(id string) (func(), error) {
	closer, err := i.mq.SubscribeQueue(
		"chat."+id,
		func(seq uint64, data []byte) {
			msg, err := goch.DecodeMsg(data)
			if err != nil {
				msg = &goch.Message{
					FromUID: "ingest",
					Text:    "ingest: message unavailable: decoding error",
					Time:    time.Now().UnixNano(),
				}
			}

			msg.Seq = seq
			// TODO: Handle error via ACK
			i.store.AppendMessage(id, msg)
		},
	)

	if err != nil {
		return nil, fmt.Errorf("ingest: couldn't subscribe: %v", err)
	}

	return func() { closer.Close() }, nil
}
