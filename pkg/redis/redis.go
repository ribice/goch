package redis

import (
	"fmt"
	"strconv"

	"github.com/vmihailenco/msgpack"

	"github.com/go-redis/redis"
	"github.com/ribice/goch"
)

const (
	chanListKey             = "channel.list"
	historyPrefix           = "history"
	chatPrefix              = "chat"
	chatLastSeqPrefix       = "last_seq"
	chatClientLastSeqPrefix = "client.last_seq"

	maxHistorySize int64 = 1000
)

// Client represents Redis client
type Client struct {
	cl *redis.Client
}

// New instantiates new Redis client
func New(addr, pass string, port int) (*Client, error) {
	opts := redis.Options{
		Addr: addr + ":" + strconv.Itoa(port),
	}
	if pass != "" {
		opts.Password = pass
	}

	client := redis.NewClient(&opts)

	_, err := client.Ping().Result()
	if err != nil {
		return nil, fmt.Errorf("Cannot connect to Redis Addr %v, Port %v Reason %v", addr, port, err)
	}
	return &Client{cl: client}, nil
}

// Get retrieves chat from Client
func (s *Client) Get(id string) (*goch.Chat, error) {
	val, err := s.cl.Get(chatID(id)).Result()
	if err != nil {
		return nil, err
	}

	return goch.DecodeChat(val)
}

// GetRecent returns list of recent messages, and sequence until last message
func (s *Client) GetRecent(id string, n int64) ([]goch.Message, uint64, error) {
	cmd := s.cl.LRange(chatHistoryID(id), -n, -1)
	if cmd.Err() != nil {
		return nil, 0, cmd.Err()
	}

	data, err := cmd.Result()
	if err != nil {
		return nil, 0, err
	}

	if data == nil || len(data) == 0 {
		return nil, 0, nil
	}

	var seq uint64
	msgs := make([]goch.Message, len(data))

	for i, m := range data {
		msg, err := goch.DecodeMsg([]byte(m))
		if err != nil {
			msg.Text = "message unavailable!"
		} else {
			seq = msgs[i].Seq
		}
	}

	return msgs, (seq + 1), nil
}

// AppendMessage adds new message
func (s *Client) AppendMessage(id string, m *goch.Message) error {
	data, err := m.Encode()
	if err != nil {
		data, _ = msgpack.Marshal([]byte(`{"text":"message unavailable, unable to encode","from":"goch/client"}`))
	}

	key := chatHistoryID(id)

	if err := s.cl.RPush(key, data).Err(); err != nil {
		return err
	}

	s.updateChannelSeq(id, m.Seq)

	return s.cl.LTrim(key, -maxHistorySize, -1).Err()
}

func (s *Client) updateChannelSeq(id string, seq uint64) {
	var currSeq int64

	val, err := s.cl.Get(chatLastSeqID(id)).Result()
	if err != nil {
		if err != redis.Nil {
			return
		}
		val = "0"
	}

	currSeq, _ = strconv.ParseInt(val, 10, 64)

	if uint64(currSeq) >= seq {
		return
	}

	s.cl.Set(chatLastSeqID(id), seq, 0)
}

// UpdateLastClientSeq updates client's last seen message
func (s *Client) UpdateLastClientSeq(uid string, id string, seq uint64) {
	var currSeq int64

	val, err := s.cl.Get(chatClientLastSeqID(uid, id)).Result()
	if err != nil {
		if err != redis.Nil {
			return
		}
		val = "0"
	}

	currSeq, _ = strconv.ParseInt(val, 10, 64)

	if uint64(currSeq) >= seq {
		return
	}

	s.cl.Set(chatClientLastSeqID(uid, id), seq, 0)
}

// GetUnreadCount returns number of unread messages
func (s *Client) GetUnreadCount(uid string, id string) uint64 {
	val, err := s.cl.Get(chatClientLastSeqID(uid, id)).Result()
	if err != nil {
		if err != redis.Nil {
			return 0
		}
		val = "0"
	}

	useq, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0
	}

	val, err = s.cl.Get(chatLastSeqID(id)).Result()
	if err != nil {
		return 0
	}

	cseq, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0
	}

	delta := cseq - useq

	if delta <= 0 {
		return 0
	}

	return uint64(delta)
}

// Save saves new chat
func (s *Client) Save(ct *goch.Chat) error {
	data, err := ct.Encode()
	if err != nil {
		return err
	}

	pipe := s.cl.TxPipeline()
	pipe.Set(chatID(ct.Name), data, 0)

	// Save only public channels
	if ct.Secret == "" {
		pipe.SAdd(chanListKey, ct.Name)
	}

	_, err = pipe.Exec()
	return err
}

// ListChannels returns list of all channels
func (s *Client) ListChannels() ([]string, error) {
	return s.cl.SMembers(chanListKey).Result()
}

func chatID(id string) string {
	return fmt.Sprintf("%s.%s", chatPrefix, id)
}

func chatHistoryID(id string) string {
	return fmt.Sprintf("%s.%s.%s", historyPrefix, chatPrefix, id)
}

func chatLastSeqID(id string) string {
	return fmt.Sprintf("%s.%s.%s", chatLastSeqPrefix, chatPrefix, id)
}

func chatClientLastSeqID(uid, id string) string {
	return fmt.Sprintf("%s.%s.%s", chatClientLastSeqPrefix, uid, id)
}
