package sessionredis

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-http-utils/cookie"
	"github.com/go-http-utils/cookie-session"

	"time"

	redis "gopkg.in/redis.v5"
)

// Options ...
type Options struct {
	Addr       string
	Keys       []string
	Expiration time.Duration
	DB         int
	Password   string
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
		Keys:       nil,
		DB:         0, // use default DB
		Addr:       "127.0.0.1:6379",
	}
	if len(options) > 0 {
		options := options[0]
		if options.Expiration > time.Second {
			opts.Expiration = options.Expiration
		}
		opts.Keys = options.Keys
		opts.DB = options.DB
		opts.Password = options.Password
		if options.Addr != "" {
			opts.Addr = options.Addr
		}
	}
	store = &RedisStore{opts: opts}

	store.client = redis.NewClient(&redis.Options{
		Addr:     opts.Addr,
		Password: opts.Password,
		DB:       opts.DB,
	})

	return
}

// Get existed session from Request's cookies
func (c *RedisStore) Get(name string, w http.ResponseWriter, r *http.Request, signed ...bool) (session *sessions.Session, err error) {
	b := false
	if len(signed) > 0 {
		b = signed[0]
	}
	var cookies *cookie.Cookies
	if len(c.opts.Keys) > 0 && len(c.opts.Keys[0]) > 0 {
		cookies = cookie.New(w, r, c.opts.Keys)
	} else {
		cookies = cookie.New(w, r)
	}
	session = sessions.NewSession(name, c, cookies, b)
	sid, _ := cookies.Get(name, b)
	if sid != "" {
		val, rediserror := c.client.Get(sid).Result()
		if err != nil {
			return nil, rediserror
		}
		b, decodeerror := base64.StdEncoding.DecodeString(val)
		if decodeerror != nil {
			return nil, decodeerror
		}
		err = json.Unmarshal(b, &session.Values)
	} else {
		sid, _ = NewUUID()
	}
	session.SID = sid
	session.SetCache(session.Values)
	return
}

// Save session to Response's cookie
func (c *RedisStore) Save(session *sessions.Session) (err error) {
	if session.IsCache() {
		return
	}
	b, err := json.Marshal(session.Values)
	if err != nil {
		return
	}
	val := base64.StdEncoding.EncodeToString(b)
	err = c.client.Set(session.SID, val, c.opts.Expiration).Err()
	if err != nil {
		return
	}
	opts := &cookie.Options{
		Path:     "/",
		HTTPOnly: true,
		Signed:   session.IsSigned(),
		MaxAge:   int(c.opts.Expiration / time.Second),
	}
	session.GetCookie().Set(session.Name(), session.SID, opts)
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
