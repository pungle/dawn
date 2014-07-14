//Copyright (C) Mr.Pungle

package web

import (
	"regexp"
	"strings"
)

//------------------ MappingResolver ------------------

type MappingResolver struct {
	handlers map[string]Handler
	isEmpty  bool
}

func NewMappingResolver() Resolver {
	handlers := make(map[string]Handler)
	return &MappingResolver{handlers, true}
}

func (self *MappingResolver) AddHandler(uri string, handler Handler) error {
	self.handlers[uri] = handler
	self.isEmpty = false
	return nil
}

func (self *MappingResolver) Resolve(uri string) (Handler, map[string]string) {
	if self.isEmpty {
		return nil, nil
	}
	return self.handlers[uri], nil
}

//------------------ PrefixResolver ------------------

type prefixHandler struct {
	prefix  string
	handler Handler
}

type PrefixResolver struct {
	handlers []*prefixHandler
	isEmpty  bool
}

func NewPrefixResolver() Resolver {
	handlers := make([]*prefixHandler, 0, 10)
	return &PrefixResolver{handlers, true}
}

func fixPrefix(prefix string) string {
	lastIdx := len(prefix) - 1
	if prefix[lastIdx] != '/' {
		prefix += "/"
	}
	return prefix
}

func (self *PrefixResolver) AddHandler(prefix string, handler Handler) error {
	prefix = fixPrefix(prefix)
	hand := &prefixHandler{prefix, handler}
	self.handlers = append(self.handlers, hand)
	self.isEmpty = false
	return nil
}

func (self *PrefixResolver) Resolve(uri string) (Handler, map[string]string) {
	if self.isEmpty {
		return nil, nil
	}
	handlers := self.handlers
	for _, hand := range handlers {
		if strings.HasPrefix(uri, hand.prefix) {
			return hand.handler, nil
		}
	}
	return nil, nil
}

//------------------ RegexpResolver ------------------

type regexpHandler struct {
	re      *regexp.Regexp
	handler Handler
}

type RegexpResolver struct {
	handlers []*regexpHandler
	isEmpty  bool
}

func NewRegexpResolver() Resolver {
	handlers := make([]*regexpHandler, 0, 10)
	return &RegexpResolver{handlers, true}
}

func (self *RegexpResolver) AddHandler(patternStr string, handler Handler) error {
	if patternStr[0] != '^' {
		patternStr = "^" + patternStr
	}

	re := regexp.MustCompile("\\{(\\w+)(\\s?:\\s?)(.*?)*\\}\\$")
	patternStr = re.ReplaceAllString(patternStr, "(?P<$1>$3)")
	re = regexp.MustCompile(patternStr)
	hand := &regexpHandler{re, handler}
	self.handlers = append(self.handlers, hand)
	self.isEmpty = false
	return nil
}

func (self *RegexpResolver) Resolve(uri string) (Handler, map[string]string) {
	if self.isEmpty {
		return nil, nil
	}
	handlers := self.handlers
	for _, hand := range handlers {
		matchs := hand.re.FindStringSubmatch(uri)
		s := len(matchs)
		if s > 0 {
			if s == 1 {
				return hand.handler, nil
			}
			vars := make(map[string]string)
			names := hand.re.SubexpNames()
			for idx, name := range names {
				if idx == 0 {
					continue
				}
				vars[name] = matchs[idx]
			}
			return hand.handler, vars
		}
	}
	return nil, nil
}
