package sessionredis_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	sessionredis "github.com/mushroomsir/session-redis"
	"github.com/stretchr/testify/assert"
)

func TestRedisStore(t *testing.T) {

	cookiekey := "cookiekey"
	cookieNewKey := "cookiekeynew"
	t.Run("RedisStore with default options that should be", func(t *testing.T) {

		assert := assert.New(t)
		req, err := http.NewRequest("GET", "/", nil)
		recorder := httptest.NewRecorder()

		store := sessionredis.New()
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, err := store.Get(cookiekey, w, r)
			session.Values["name"] = "mushroom"
			session.Values["num"] = 99
			assert.Nil(err)
			err = session.Save()
			assert.Nil(err)

		})
		handler.ServeHTTP(recorder, req)

		//======reuse=====
		cookies, err := getCookie(cookiekey, recorder)
		assert.Nil(err)
		assert.NotNil(cookies.Value)
		req, err = http.NewRequest("GET", "/", nil)
		req.AddCookie(cookies)

		store = sessionredis.New()
		handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			session, err := store.Get(cookiekey, w, r)

			assert.Equal("mushroom", session.Values["name"])
			assert.Equal(float64(99), session.Values["num"])
			assert.Nil(err)
		})
		handler.ServeHTTP(recorder, req)

	})

	t.Run("RedisStore with custom options that should be", func(t *testing.T) {
		store := sessionredis.New(&sessionredis.Options{
			Expiration: 24 * time.Hour,
			DB:         0, // use default DB
			Addr:       "127.0.0.1:6379",
		})

		assert := assert.New(t)
		req, err := http.NewRequest("GET", "/", nil)
		recorder := httptest.NewRecorder()
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, err := store.Get(cookiekey, w, r)

			session.Values["name"] = "mushroom"
			session.Values["num"] = 99
			session.Save()

			assert.Nil(err)
			session, err = store.Get(cookiekey+"error", w, r)
			assert.Nil(err)
			assert.Equal(0, len(session.Values))
		})
		handler.ServeHTTP(recorder, req)

		//======reuse=====
		store = sessionredis.New(&sessionredis.Options{
			Expiration: 24 * time.Hour,
			DB:         0, // use default DB
			Addr:       "127.0.0.1:6379",
		})
		cookies, err := getCookie(cookiekey, recorder)
		assert.Nil(err)
		assert.NotNil(cookies.Value)
		t.Log(cookies.Value)
		req, err = http.NewRequest("GET", "/", nil)

		req.AddCookie(cookies)

		handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			session, err := store.Get(cookiekey, w, r)
			assert.Nil(err)
			assert.NotNil(session.Values)
			assert.Equal("mushroom", session.Values["name"])
			assert.Equal(float64(99), session.Values["num"])
		})
		handler.ServeHTTP(recorder, req)

	})

	t.Run("RedisStore donn't override old value when seting same value that should be", func(t *testing.T) {
		assert := assert.New(t)
		req, err := http.NewRequest("GET", "/", nil)
		recorder := httptest.NewRecorder()

		store := sessionredis.New()
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, err := store.Get(cookiekey, w, r)

			session.Values["name"] = "mushroom"
			session.Values["num"] = 99
			session.Save()
			assert.Nil(err)

		})
		handler.ServeHTTP(recorder, req)

		//======reuse=====
		cookies, err := getCookie(cookiekey, recorder)
		assert.Nil(err)
		assert.NotNil(cookies.Value)
		req, err = http.NewRequest("GET", "/", nil)
		req.AddCookie(cookies)

		store = sessionredis.New()
		handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, err := store.Get(cookiekey, w, r)
			session.Save()
			assert.Nil(err)
		})
		handler.ServeHTTP(recorder, req)
	})
	t.Run("RedisStore with sign session that should be", func(t *testing.T) {
		store := sessionredis.New(&sessionredis.Options{
			Keys: []string{"key"},
		})
		assert := assert.New(t)

		recorder := httptest.NewRecorder()

		req, _ := http.NewRequest("GET", "/", nil)
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, err := store.Get(cookiekey, w, r)
			session.Values["name"] = "mushroom"
			session.Values["num"] = 99
			session.Save()
			assert.Nil(err)

			session, err = store.Get(cookieNewKey, w, r)
			session.Values["name"] = "mushroom-n"
			session.Values["num"] = 100
			session.Save()
			assert.Nil(err)
		})
		handler.ServeHTTP(recorder, req)

		//======reuse=====
		store = sessionredis.New(&sessionredis.Options{
			Keys: []string{"key"},
		})
		req, _ = http.NewRequest("GET", "/", nil)
		cookies, _ := getCookie(cookiekey, recorder)
		req.AddCookie(cookies)
		cookies, _ = getCookie(cookiekey+".sig", recorder)
		req.AddCookie(cookies)

		cookies, _ = getCookie(cookieNewKey, recorder)
		req.AddCookie(cookies)
		cookies, _ = getCookie(cookieNewKey+".sig", recorder)
		req.AddCookie(cookies)

		store = sessionredis.New(&sessionredis.Options{
			Keys: []string{"key"},
		})
		handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, err := store.Get(cookiekey, w, r)

			assert.Nil(err)
			assert.Equal("mushroom", session.Values["name"])
			assert.Equal(float64(99), session.Values["num"])

			session, err = store.Get(cookieNewKey, w, r)
			assert.Nil(err)
			assert.Equal("mushroom-n", session.Values["name"])
			assert.Equal(float64(100), session.Values["num"])

			session, err = store.Get(cookieNewKey+"new", w, r)
			assert.Nil(err)
			assert.Equal(0, len(session.Values))

		})
		handler.ServeHTTP(recorder, req)

		//======reuse=====
		store = sessionredis.New(&sessionredis.Options{
			Keys: []string{"key"},
		})
		req, _ = http.NewRequest("GET", "/", nil)
		cookies, _ = getCookie(cookiekey, recorder)
		req.AddCookie(cookies)
		cookies, _ = getCookie(cookiekey+".sig", recorder)
		req.AddCookie(cookies)

		cookies, _ = getCookie(cookieNewKey, recorder)
		req.AddCookie(cookies)
		cookies, _ = getCookie(cookieNewKey+".sig", recorder)
		req.AddCookie(cookies)

		store = sessionredis.New(&sessionredis.Options{
			Keys: []string{"key"},
		})
		handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, err := store.Get(cookiekey, w, r)

			assert.Nil(err)
			assert.Equal("mushroom", session.Values["name"])
			assert.Equal(float64(99), session.Values["num"])

			session, err = store.Get(cookieNewKey, w, r)
			assert.Nil(err)
			assert.Equal("mushroom-n", session.Values["name"])
			assert.Equal(float64(100), session.Values["num"])

			session, err = store.Get(cookieNewKey+"new", w, r)
			assert.Nil(err)
			assert.Equal(0, len(session.Values))

		})
		handler.ServeHTTP(recorder, req)

	})
}
func getCookie(name string, recorder *httptest.ResponseRecorder) (*http.Cookie, error) {
	var err error
	res := &http.Response{Header: http.Header{"Set-Cookie": recorder.HeaderMap["Set-Cookie"]}}
	for _, val := range res.Cookies() {
		if val.Name == name {
			return val, nil
		}
	}
	return nil, err
}
