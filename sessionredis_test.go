package sessionredis_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-http-utils/cookie"
	"github.com/go-http-utils/cookie-session"
	sessionredis "github.com/mushroomsir/session-redis"
	"github.com/stretchr/testify/assert"
)

// Session ...
type Session struct {
	*sessions.Meta `json:"-"`
	UserID         string `json:"userId"`
	Name           string `json:"name"`
	Age            int64  `json:"authed"`
}

// Save ...
func (s *Session) Save() error {
	return s.SaveIt(s)
}

func TestRedisStore(t *testing.T) {

	SessionName := "Sess"
	SessionKeys := []string{"keyxxx"}
	NewSessionName := "NewSess"

	t.Run("RedisStore with default options that should be", func(t *testing.T) {

		assert := assert.New(t)
		req, err := http.NewRequest("GET", "/", nil)
		recorder := httptest.NewRecorder()

		store := sessionredis.New()
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			session := &Session{Meta: &sessions.Meta{}}
			store.Load(SessionName, session, cookie.New(w, r, SessionKeys))
			if session.UserID == "" {
				session.UserID = "123465"
				session.Name = "mushroom"
				session.Age = 18
			}
			err = session.Save()
			assert.Nil(err)
		})
		handler.ServeHTTP(recorder, req)

		//====== reuse session =====
		req, err = http.NewRequest("GET", "/", nil)
		migrateCookies(recorder, req)

		store = sessionredis.New()
		handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			session := &Session{Meta: &sessions.Meta{}}
			store.Load(SessionName, session, cookie.New(w, r, SessionKeys))

			assert.Equal("123465", session.UserID)
			assert.Equal("mushroom", session.Name)
			assert.Equal(int64(18), session.Age)
		})
		handler.ServeHTTP(recorder, req)
	})

	t.Run("RedisStore with custom options that should be", func(t *testing.T) {
		assert := assert.New(t)

		store := sessionredis.New(&sessionredis.Options{
			Expiration: 24 * time.Hour,
			DB:         0, // use default DB
			Addr:       "127.0.0.1:6379",
		})

		req, _ := http.NewRequest("GET", "/", nil)
		recorder := httptest.NewRecorder()

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session := &Session{Meta: &sessions.Meta{}}
			store.Load(SessionName, session, cookie.New(w, r, SessionKeys))
			if session.UserID == "" {
				session.UserID = "1234654"
				session.Name = "mushroom"
				session.Age = 19
			}
			session.Save()

			session = &Session{Meta: &sessions.Meta{}}
			store.Load(SessionName+"error", session, cookie.New(w, r, SessionKeys))

			assert.Equal(int64(0), session.Age)
			assert.Equal("", session.UserID)
			assert.Equal("", session.Name)
		})
		handler.ServeHTTP(recorder, req)

		//====== reuse session =====
		req, _ = http.NewRequest("GET", "/", nil)
		migrateCookies(recorder, req)

		store = sessionredis.New(&sessionredis.Options{
			Expiration: 24 * time.Hour,
			DB:         0, // use default DB
			Addr:       "127.0.0.1:6379",
		})
		handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			session := &Session{Meta: &sessions.Meta{}}
			store.Load(SessionName, session, cookie.New(w, r, SessionKeys))

			assert.Equal("1234654", session.UserID)
			assert.Equal("mushroom", session.Name)
			assert.Equal(int64(19), session.Age)
		})
		handler.ServeHTTP(recorder, req)

	})

	t.Run("RedisStore donn't override old value when seting same value that should be", func(t *testing.T) {
		assert := assert.New(t)
		req, _ := http.NewRequest("GET", "/", nil)
		recorder := httptest.NewRecorder()

		store := sessionredis.New()
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session := &Session{Meta: &sessions.Meta{}}
			store.Load(SessionName, session, cookie.New(w, r, SessionKeys))

			session.UserID = "1234654"
			session.Name = "mushroom"
			session.Age = 19

			session.Save()
		})
		handler.ServeHTTP(recorder, req)

		//====== reuse session =====
		req, _ = http.NewRequest("GET", "/", nil)
		migrateCookies(recorder, req)

		store = sessionredis.New()
		handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			session := &Session{Meta: &sessions.Meta{}}
			store.Load(SessionName, session, cookie.New(w, r, SessionKeys))

			assert.Equal("1234654", session.UserID)
			assert.Equal("mushroom", session.Name)
			assert.Equal(int64(19), session.Age)
			session.UserID = "1234654"
			session.Name = "mushroom"
			session.Age = 19

			session.Save()
		})
		handler.ServeHTTP(recorder, req)
	})
	t.Run("RedisStore with sign session that should be", func(t *testing.T) {
		assert := assert.New(t)

		recorder := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)

		store := sessionredis.New()
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			session := &Session{Meta: &sessions.Meta{}}
			store.Load(SessionName, session, cookie.New(w, r, SessionKeys))

			session.UserID = "1234654"
			session.Name = "mushroom"
			session.Age = 19
			session.Save()

			session = &Session{Meta: &sessions.Meta{}}
			store.Load(NewSessionName, session, cookie.New(w, r, SessionKeys))

			session.UserID = "12346543"
			session.Name = "mushrooma"
			session.Age = 20
			session.Save()
		})
		handler.ServeHTTP(recorder, req)

		//====== reuse session =====

		req, _ = http.NewRequest("GET", "/", nil)
		migrateCookies(recorder, req)

		store = sessionredis.New()
		handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session := &Session{Meta: &sessions.Meta{}}
			store.Load(SessionName, session, cookie.New(w, r, SessionKeys))

			assert.Equal("1234654", session.UserID)
			assert.Equal("mushroom", session.Name)
			assert.Equal(int64(19), session.Age)

			session = &Session{Meta: &sessions.Meta{}}
			store.Load(NewSessionName, session, cookie.New(w, r, SessionKeys))

			assert.Equal("12346543", session.UserID)
			assert.Equal("mushrooma", session.Name)
			assert.Equal(int64(20), session.Age)

			session = &Session{Meta: &sessions.Meta{}}
			store.Load(SessionName+"new", session, cookie.New(w, r, SessionKeys))

			assert.Equal(int64(0), session.Age)
			assert.Equal("", session.UserID)
			assert.Equal("", session.Name)

		})
		handler.ServeHTTP(recorder, req)

		//====== reuse session =====
		req, _ = http.NewRequest("GET", "/", nil)
		migrateCookies(recorder, req)

		store = sessionredis.New()

		handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session := &Session{Meta: &sessions.Meta{}}
			store.Load(SessionName, session, cookie.New(w, r, SessionKeys))

			assert.Equal("1234654", session.UserID)
			assert.Equal("mushroom", session.Name)
			assert.Equal(int64(19), session.Age)

			session = &Session{Meta: &sessions.Meta{}}
			store.Load(NewSessionName, session, cookie.New(w, r, SessionKeys))

			assert.Equal("12346543", session.UserID)
			assert.Equal("mushrooma", session.Name)
			assert.Equal(int64(20), session.Age)

			session = &Session{Meta: &sessions.Meta{}}
			store.Load(SessionName+"new", session, cookie.New(w, r, SessionKeys))

			assert.Equal(int64(0), session.Age)
			assert.Equal("", session.UserID)
			assert.Equal("", session.Name)
		})
		handler.ServeHTTP(recorder, req)
	})
}
func migrateCookies(recorder *httptest.ResponseRecorder, req *http.Request) {
	for _, cookie := range recorder.Result().Cookies() {
		req.AddCookie(cookie)
	}
}
