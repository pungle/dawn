//Copyright (C) Mr.Pungle

package web

import (
	"errors"
	"github.com/pungle/dawn/logging"
	"github.com/pungle/dawn/uuid"
	"net/http"
	"time"
)

type SessionDriver interface {
	Get(string) (interface{}, error)
	Set(string, interface{}, time.Duration) error
}

var (
	EmptySessionErr   = errors.New("session is empty")
	SessionNotInitErr = errors.New("session not init")
)

type Session interface {
	Get(string) (interface{}, error)
	Set(string, interface{}) error
	Values() (map[string]interface{}, error)
	ID() (string, error)
}

type httpSession struct {
	sid  string
	data map[string]interface{}
}

func (self *httpSession) Get(key string) (interface{}, error) {
	if self.data == nil {
		return nil, SessionNotInitErr
	}
	return self.data[key], nil
}

func (self *httpSession) Set(key string, value interface{}) error {
	if self.data == nil {
		return SessionNotInitErr
	}
	self.data[key] = value
	return nil
}

func (self *httpSession) Values() (map[string]interface{}, error) {
	if self.data == nil {
		return nil, SessionNotInitErr
	}
	return self.data, nil
}

func (self *httpSession) ID() (string, error) {
	if self.data == nil {
		return "", SessionNotInitErr
	}
	return self.sid, nil
}

type SessionContext struct {
	driver SessionDriver

	cookieName     string
	cookiePath     string
	cookieDomain   string
	cookieExpire   time.Duration
	cookieHttpOnly bool
	cookieSecure   bool

	sessionAge time.Duration
}

func NewSessionContext(driver SessionDriver, cookieName string,
	cookieDomain string, cookieExpire time.Duration,
	cookiePath string, cookieHttpOnly bool,
	cookieSecure bool, sessionAge time.Duration) *SessionContext {

	return &SessionContext{
		driver,
		cookieName,
		cookiePath,
		cookieDomain,
		cookieExpire,
		cookieHttpOnly,
		cookieSecure,
		sessionAge,
	}
}

func (self *SessionContext) New() Session {
	session := &httpSession{
		uuid.NewUUID().Base64(),
		make(map[string]interface{}),
	}
	return session
}

func (self *SessionContext) Loads(req *http.Request) Session {
	cookie, _ := req.Cookie(self.cookieName)

	if cookie != nil {
		data, err := self.driver.Get(cookie.Value)
		if err != nil {
			logging.Error("LoadSession error: %s, key: %s", err.Error(), cookie.Value)
			return nil
		}
		if data == nil {
			return nil
		}
		sessionData := data.(map[string]interface{})
		session := &httpSession{
			cookie.Value,
			sessionData,
		}
		return session
	}
	return nil
}

func (self *SessionContext) Save(resp http.ResponseWriter, session Session) error {
	data, err := session.Values()
	if err != nil {
		return err
	}
	if len(data) == 0 {
		logging.Warn("Saving an empty session.")
		return EmptySessionErr
	}

	sid, err := session.ID()
	if err != nil {
		return err
	}

	err = self.driver.Set(sid, data, self.sessionAge)

	if err != nil {
		logging.Error("SaveSession error: %s, sid: %s", err.Error(), sid)
		return err
	}
	cookieTime := time.Now().Add(self.cookieExpire)
	cookie := &http.Cookie{
		Name:     self.cookieName,
		Value:    sid,
		Path:     self.cookiePath,
		Domain:   self.cookieDomain,
		Expires:  cookieTime,
		MaxAge:   int(self.cookieExpire / time.Second),
		HttpOnly: self.cookieHttpOnly,
		Secure:   self.cookieSecure,
	}
	http.SetCookie(resp, cookie)
	return nil
}
