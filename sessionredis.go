package sessionredis

import (
	"crypto/rand"
	"fmt"
	"io"
	"time"

	"github.com/go-http-utils/cookie"
	"github.com/go-http-utils/cookie-session"
	redis "gopkg.in/redis.v5"
)

// Options ...
type Options struct {
	Addr       string
	Expiration time.Duration
	DB         int
	Password   string
	opts       *cookie.Options
}

//RedisStore backend for cookie-session
type RedisStore struct {
	opts   *Options
	client *redis.Client
}

// New an CookieStore instance
func New(options ...*Options) (store *RedisStore) {
	opts := &Options{
		Expiration: 24 * time.Hour,
		DB:         0, // use default DB
		Addr:       "127.0.0.1:6379",
	}
	if len(options) > 0 {
		options := options[0]
		if options.Expiration > time.Second {
			opts.Expiration = options.Expiration
		}
		opts.DB = options.DB
		opts.Password = options.Password
		if options.Addr != "" {
			opts.Addr = options.Addr
		}
	}
	opts.opts = &cookie.Options{
		Path:     "/",
		HTTPOnly: true,
		Signed:   true,
		MaxAge:   int(opts.Expiration / time.Second),
	}
	store = &RedisStore{opts: opts}
	store.client = redis.NewClient(&redis.Options{
		Addr:     opts.Addr,
		Password: opts.Password,
		DB:       opts.DB,
	})
	return
}

// Load existed session from Request's cookies
func (c *RedisStore) Load(name string, session sessions.Sessions, cookie *cookie.Cookies) (err error) {
	sid, err := cookie.Get(name, c.opts.opts.Signed)
	var result string
	if sid != "" {
		result, err = c.client.Get(sid).Result()
		if err == nil && result != "" {
			err = sessions.Decode(result, &session)
		}
	}
	session.Init(name, sid, cookie, c, result)
	return err
}

// Save session to Response's cookie
func (c *RedisStore) Save(session sessions.Sessions) (err error) {
	val, err := sessions.Encode(session)
	if err != nil || !session.IsChanged(val) {
		return
	}
	sid := session.GetSID()
	if sid == "" {
		sid, _ = NewUUID()
	}
	err = c.client.Set(sid, val, c.opts.Expiration).Err()
	if err == nil {
		session.GetCookie().Set(session.GetName(), sid, c.opts.opts)
	}
	return
}

// NewUUID generates a random UUID according to RFC 4122
func NewUUID() (string, error) {
	uuid := make([]byte, 16)
	io.ReadFull(rand.Reader, uuid)
	uuid[8] = uuid[8]&^0xc0 | 0x80
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x%x%x%x%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:]), nil
}
