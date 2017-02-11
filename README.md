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
go run examples/main.go
```
##Usage
```go
    store := sessionredis.New()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	    session, _ := store.Get(sessionkey, w, r)
		if val, ok := session.Values["name"]; ok {
			println(val)
		} else {
			session.Values["name"] = "mushroom"
		}
		session.Save()
	})