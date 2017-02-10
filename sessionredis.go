package sessionredis

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-http-utils/cookie"

	"time"

	redis "gopkg.in/redis.v5"
)

// Options ...
type Options struct {
	addr       string
	keys       []string
	Expiration time.Duration
	DB         int
	Password   string
}

//RedisStore ...
type RedisStore struct {
	opts   *Options
	client *redis.Client
	cookie *cookie.Cookies
	signed bool
}

// New an CookieStore instance
func New(options ...*Options) (store *RedisStore) {
	opts := &Options{
		Expiration: 24 * time.Hour,
		keys:       nil,
		DB:         0, // use default DB
		addr:       "127.0.0.1:6379",
	}
	if len(options) > 0 {
		opts = options[0]
	}
	store = &RedisStore{opts: opts}

	store.client = redis.NewClient(&redis.Options{
		Addr:     opts.addr,
		Password: opts.Password,
		DB:       opts.DB,
	})

	return
}

// Init an CookieStore instance
func (c *RedisStore) Init(w http.ResponseWriter, r *http.Request, signed bool) {

	if len(c.opts.keys) > 0 && len(c.opts.keys[0]) > 0 {
		c.cookie = cookie.New(w, r, c.opts.keys)
	} else {
		c.cookie = cookie.New(w, r)
	}
	c.signed = signed
	return
}

// Get existed session from Request's cookies
func (c *RedisStore) Get(name string) (data map[string]interface{}, err error) {
	sid, err := c.cookie.Get(name, c.signed)
	if err != nil {
		return
	}
	val, err := c.client.Get(sid).Result()
	if err != nil {
		return
	}
	b, err := base64.StdEncoding.DecodeString(val)
	if err != nil {
		return
	}
	err = json.Unmarshal(b, &data)
	return
}

// Save session to Response's cookie
func (c *RedisStore) Save(name string, data map[string]interface{}) (err error) {
	sid, err := NewUUID()
	if err != nil {
		return
	}
	b, err := json.Marshal(data)
	if err != nil {
		return
	}
	val := base64.StdEncoding.EncodeToString(b)
	cmd := c.client.Set(sid, val, c.opts.Expiration)
	err = cmd.Err()
	if err != nil {
		return
	}
	opts := &cookie.Options{
		Path:     "/",
		HTTPOnly: true,
		Signed:   c.signed,
		MaxAge:   int(c.opts.Expiration / time.Second),
	}
	c.cookie.Set(name, sid, opts)
	return
}

// NewUUID generates a random UUID according to RFC 4122
func NewUUID() (string, error) {
	uuid := make([]byte, 16)
	n, err := io.ReadFull(rand.Reader, uuid)
	if n != len(uuid) || err != nil {
		return "", err
	}
	// variant bits; see section 4.1.1
	uuid[8] = uuid[8]&^0xc0 | 0x80
	// version 4 (pseudo-random); see section 4.1.3
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x%x%x%x%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:]), nil
}