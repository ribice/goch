package nats

import (
	"fmt"
	"io"
	"time"

	stan "github.com/nats-io/go-nats-streaming"
)

// Client represents NATS client
type Client struct {
	cn stan.Conn
}

// New initializes a connection to NATS server
func New(clusterID, clientID, url string) (*Client, error) {
	conn, err := stan.Connect(clusterID, clientID, stan.NatsURL(url))
	if err != nil {
		return nil, fmt.Errorf("error connecting to NATS: %v", err)
	}
	return &Client{cn: conn}, nil
}

// SubscribeQueue subscribers to a message queue
func (c *Client) SubscribeQueue(subj string, f func(uint64, []byte)) (io.Closer, error) {
	return c.cn.QueueSubscribe(
		subj,
		"ingest",
		func(m *stan.Msg) {
			f(m.Sequence, m.Data)
		},
		stan.SetManualAckMode(),
	)
}

// SubscribeSeq subscribers to a message queue from received sequence
func (c *Client) SubscribeSeq(id string, nick string, start uint64, f func(uint64, []byte)) (io.Closer, error) {
	return c.cn.Subscribe(
		id,
		func(m *stan.Msg) {
			f(m.Sequence, m.Data)
		},
		stan.StartAtSequence(start),
		stan.SetManualAckMode(),
	)
}

// SubscribeTimestamp subscribers to a message queue from received time.Time
func (c *Client) SubscribeTimestamp(id string, nick string, t time.Time, f func(uint64, []byte)) (io.Closer, error) {
	return c.cn.Subscribe(
		id,
		func(m *stan.Msg) {
			f(m.Sequence, m.Data)
		},
		stan.StartAtTime(t),
		stan.SetManualAckMode(),
	)
}

// Send publishes new message
func (c *Client) Send(id string, msg []byte) error {
	return c.cn.Publish(id, msg)
}
