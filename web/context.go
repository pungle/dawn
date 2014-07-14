//Copyright (C) Mr.Pungle

package web

import (
	"errors"
	"net/http"
)

var (
	ErrSessionNotSetup = errors.New("SessionDoesNotSetup")
)

type HttpContext struct {
	Request  *http.Request
	Response http.ResponseWriter
	vars     map[string]string

	sessionCtx *SessionContext

	curSession Session
}

func NewHttpContext(response http.ResponseWriter, request *http.Request,
	sessionCtx *SessionContext, vars map[string]string) *HttpContext {
	return &HttpContext{request, response, vars, sessionCtx, nil}
}

func (self *HttpContext) Session() Session {
	if self.sessionCtx == nil {
		panic(ErrSessionNotSetup)
	}
	if self.curSession != nil {
		return self.curSession
	}
	self.curSession = self.sessionCtx.Loads(self.Request)
	return self.curSession
}

func (self *HttpContext) NewSession() Session {
	if self.sessionCtx == nil {
		panic(ErrSessionNotSetup)
	}
	self.curSession = self.sessionCtx.New()
	return self.curSession
}

func (self *HttpContext) SaveSession() error {
	if self.sessionCtx == nil {
		panic(ErrSessionNotSetup)
	}
	return self.sessionCtx.Save(self.Response, self.curSession)
}

func (self *HttpContext) GetVar(key string) string {
	if self.vars == nil {
		return ""
	}
	return self.vars[key]
}
