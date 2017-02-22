package main

import (
	"net/http"
	"net/http/httptest"

	"github.com/go-http-utils/cookie"
	"github.com/go-http-utils/cookie-session"
	sessionredis "github.com/mushroomsir/session-redis"
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

func main() {

	SessionName := "Sess"
	SessionKeys := []string{"keyxxx"}

	recorder := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)

	store := sessionredis.New()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := &Session{Meta: &sessions.Meta{}}
		store.Load(SessionName, session, cookie.New(w, r, SessionKeys))
		if session.UserID == "" {
			session.UserID = "x"
			session.Name = "y"
			session.Age = 18
		}
		session.Save()

	})
	handler.ServeHTTP(recorder, req)
	//======reuse=====
	req, _ = http.NewRequest("GET", "/", nil)
	migrateCookies(recorder, req)

	store = sessionredis.New()
	handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := &Session{Meta: &sessions.Meta{}}
		store.Load(SessionName, session, cookie.New(w, r, SessionKeys))

		println(session.UserID)
		println(session.Name)
		println(session.Age)
	})
	handler.ServeHTTP(recorder, req)
}
func migrateCookies(recorder *httptest.ResponseRecorder, req *http.Request) {
	for _, cookie := range recorder.Result().Cookies() {
		req.AddCookie(cookie)
	}
}
