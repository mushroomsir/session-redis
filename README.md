# session-redis


[![Build Status](https://travis-ci.org/mushroomsir/session-redis.svg?branch=master)](https://travis-ci.org/mushroomsir/session-redis)
[![Coverage Status](http://img.shields.io/coveralls/mushroomsir/session-redis.svg?style=flat-square)](https://coveralls.io/r/mushroomsir/session-redis)
[![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/mushroomsir/sessionredis/master/LICENSE)
[![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](http://godoc.org/github.com/mushroomsir/sessionredis)

A session store backend for [cookie-session](https://github.com/go-http-utils/cookie-session)
## Installation
```go
go get github.com/mushroomsir/session-redis
```
##Examples
```go
go run example/main.go
```
##Usage
```go
    SessionName := "Sess"
	SessionKeys := []string{"keyxxx"}

    store := sessionredis.New()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	    session := &Session{Meta: &sessions.Meta{}}
		store.Load(SessionName, session, cookie.New(w, r, SessionKeys...))
		if session.UserID == "" {
			session.UserID = "x"
			session.Name = "y"
			session.Authed = 1
		}
		session.Save()
	})
```	