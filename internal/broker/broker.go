// Package broker provides chat broker functionality
package broker

import (
	"fmt"
	"io"
	"time"

	"github.com/ribice/goch"
)

// New creates new chat broker instance
func New(mq MQ, store ChatStore, ig Ingester) *Broker {
	return &Broker{mq: mq, store: store, ig: ig}
}

// Broker represents chat broker
type Broker struct {
	mq    MQ
	ig    Ingester
	store ChatStore
}

// MQ represents message broker interface
type MQ interface {
	Send(string, []byte) error
	SubscribeSeq(string, string, uint64, func(uint64, []byte)) (io.Closer, error)
	SubscribeTimestamp(string, string, time.Time, func(uint64, []byte)) (io.Closer, error)
}

// Ingester represents chat history read model ingester
type Ingester interface {
	Run(string) (func(), error)
}

// ChatStore represents chat store interface
type ChatStore interface {
	UpdateLastClientSeq(string, string, uint64)
}

// Subscribe subscribes to provided chat id at start sequence
// Returns close subscription func, or an error.
func (b *Broker) Subscribe(chatID, uid string, start uint64, c chan *goch.Message) (func(), error) {
	closer, err := b.mq.SubscribeSeq("chat."+chatID, uid, start, func(seq uint64, data []byte) {
		msg, err := goch.DecodeMsg(data)
		if err != nil {
			msg = &goch.Message{
				FromUID: "broker",
				Text:    "broker: message unavailable: decoding error",
				Time:    time.Now().UnixNano(),
			}
		}

		msg.Seq = seq

		if msg.FromUID != uid {
			c <- msg
		} else {
			b.store.UpdateLastClientSeq(msg.FromUID, chatID, seq)
		}
	})

	if err != nil {
		return nil, err
	}

	cleanup, err := b.ig.Run(chatID)
	if err != nil {
		closer.Close()
		return nil, fmt.Errorf("broker: unable to run ingest for chat: %v", err)
	}

	return func() { closer.Close(); cleanup() }, nil
}

// SubscribeNew subscribes to provided chat id subject starting from time.Now()
// Returns close subscription func, or an error.
func (b *Broker) SubscribeNew(chatID, uid string, c chan *goch.Message) (func(), error) {
	closer, err := b.mq.SubscribeTimestamp("chat."+chatID, uid, time.Now(), func(seq uint64, data []byte) {
		msg, err := goch.DecodeMsg(data)
		if err != nil {
			msg = &goch.Message{
				FromUID: "broker",
				Text:    "broker: message unavailable: decoding error",
				Time:    time.Now().UnixNano(),
			}
		}

		msg.Seq = seq

		if msg.FromUID != uid {
			c <- msg
		}
	})

	if err != nil {
		return nil, err
	}

	cleanup, err := b.ig.Run(chatID)
	if err != nil {
		closer.Close()
		return nil, fmt.Errorf("broker: unable to run ingest for chat: %v", err)
	}

	return func() { closer.Close(); cleanup() }, nil
}

// Send sends new message to a given chat
func (b *Broker) Send(chatID string, msg *goch.Message) error {
	data, err := msg.Encode()
	if err != nil {
		return err
	}

	return b.mq.Send("chat."+chatID, data)
}
