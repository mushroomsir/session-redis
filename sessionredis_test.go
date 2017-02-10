package sessionredis_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-http-utils/cookie-session"
	sessionredis "github.com/mushroomsir/session-redis"
	"github.com/stretchr/testify/assert"
)

func TestRedisStore(t *testing.T) {
	store := sessionredis.New()
	cookiekey := "teambition"
	t.Run("RedisStore with default options that should be", func(t *testing.T) {
		assert := assert.New(t)
		req, err := http.NewRequest("GET", "/", nil)
		recorder := httptest.NewRecorder()
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, _ := sessions.New(cookiekey, store, w, r)
			session.Values["name"] = "mushroom"
			session.Values["num"] = 99
			session.Save()

		})
		handler.ServeHTTP(recorder, req)

		//======reuse=====
		cookies, err := getCookie(cookiekey, recorder)
		assert.Nil(err)
		assert.NotNil(cookies.Value)
		t.Log(cookies.Value)
		req, err = http.NewRequest("GET", "/", nil)

		req.AddCookie(cookies)

		handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			session, _ := sessions.New(cookiekey, store, w, r)

			assert.Equal("mushroom", session.Values["name"])
			assert.Equal(float64(99), session.Values["num"])
		})
		handler.ServeHTTP(recorder, req)

		//======reuse=====
		req, err = http.NewRequest("GET", "/", nil)
		req.AddCookie(cookies)
		handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			session, _ := sessions.New(cookiekey, store, w, r)

			assert.Equal("mushroom", session.Values["name"])
			assert.Equal(float64(99), session.Values["num"])
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
