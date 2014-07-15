//Copyright (C) Mr.Pungle

package web

import (
	"github.com/pungle/dawn/logging"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

var (
	DEFAULT_LOG_FLAG  = logging.F_DATE
	DEFAULT_LOG_LEVEL = logging.L_TRACE
)

type Handler func(ctx *HttpContext)

type Resolver interface {
	AddHandler(string, Handler) error
	Resolve(string) (Handler, map[string]string)
}

type HttpConfig struct {
	addr     string
	tls      bool
	certfile string
	keyfile  string

	logFlag  int
	logLevel int
}

func NewConfig(addr string, logFlag int, logLevel int,
	tls bool, certfile string, keyfile string) *HttpConfig {
	return &HttpConfig{addr, tls, certfile, keyfile, logFlag, logLevel}
}

type loggedResponseWriter struct {
	http.ResponseWriter
	status        int
	contentLength int
}

func (self *loggedResponseWriter) WriteHeader(code int) {
	self.status = code
	self.ResponseWriter.WriteHeader(code)
}

func (self *loggedResponseWriter) Write(value []byte) (int, error) {
	self.contentLength += len(value)
	return self.ResponseWriter.Write(value)
}

type HttpServer struct {
	config     *HttpConfig
	resolvers  []Resolver
	sessionCtx *SessionContext
	logger     *logging.Logger
}

func NewServer(config *HttpConfig, sessionCtx *SessionContext, logHandler logging.Handler) *HttpServer {
	resolvers := []Resolver{
		NewMappingResolver(),
		NewPrefixResolver(),
		NewRegexpResolver(),
	}
	if logHandler == nil {
		logHandler = os.Stderr
	}
	logger := logging.NewLogger(logHandler, config.logFlag, config.logLevel)
	return &HttpServer{config, resolvers, sessionCtx, logger}
}

func (self *HttpServer) AddHandler(urlPattern string, handler Handler) (err error) {
	flag := urlPattern[0]
	var resolverIndex int = 0
	var pattern = urlPattern
	switch flag {
	case '=':
		pattern = urlPattern[1:]
	case '~':
		resolverIndex = 1
		pattern = urlPattern[1:]
	case '^':
		resolverIndex = 2
		pattern = urlPattern
	default:
		break
	}
	resolver := self.resolvers[resolverIndex]
	resolver.AddHandler(pattern, handler)
	return
}

func (self *HttpServer) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	loggedResp := &loggedResponseWriter{resp, http.StatusOK, 0}
	for _, resolver := range self.resolvers {
		handler, vars := resolver.Resolve(req.URL.Path)
		if handler != nil {
			ctx := NewHttpContext(loggedResp, req, self.sessionCtx, vars)
			loggedResp.Header().Set("Content-Type", "application/json")
			handler(ctx)
			goto logTime
		}
	}
	http.NotFound(loggedResp, req)

logTime:
	self.writeLog(loggedResp, req)
}

func (self *HttpServer) ListenAndServe() {
	go self.listen()
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGTSTP, syscall.SIGQUIT)
	<-c
}

func (self *HttpServer) listen() {
	logging.Notify("Listening %s", self.config.addr)
	logging.Notify("Https: %v", self.config.tls)
	config := self.config
	var err error
	if config.tls {
		err = http.ListenAndServeTLS(config.addr, config.certfile, config.keyfile, self)
	} else {
		err = http.ListenAndServe(config.addr, self)
	}
	if err != nil {
		logging.Error("Listen has an error: %s", err.Error())
	}
}

func (self *HttpServer) writeLog(resp *loggedResponseWriter, req *http.Request) {
	header := resp.Header()
	self.logger.Info(
		"[%s] %s%s %d %d '%s' '%s' '%s' '%s'",
		req.Method,
		req.Host,
		req.RequestURI,
		resp.status,
		resp.contentLength,
		header.Get("Content-Type"),
		req.RemoteAddr,
		req.UserAgent(),
		req.Referer(),
	)
}
