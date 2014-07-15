//Copyright (C) Mr.Pungle

package logging

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	L_TRACE = iota
	L_DEBUG
	L_INFO
	L_WARN
	L_ERROR
	L_NOTIFY
)

const (
	F_DATE = 1 << iota
	F_SHORT_FILE
	F_LONG_FILE
	F_LEVEL

	F_FILE = F_SHORT_FILE | F_LONG_FILE
)

const (
	LOG_TIME_FORMAT = "2006/01/02 15:04:05 -0700"
)

var LevelName = map[int]string{
	L_TRACE:  "TRACE",
	L_DEBUG:  "DEBUG",
	L_INFO:   "INFO",
	L_WARN:   "WARN",
	L_ERROR:  "ERROR",
	L_NOTIFY: "NOTIFY",
}

var (
	ErrInvalidLevel   = errors.New("InvalidLogLevel")
	ErrStdLogNotOpen  = errors.New("StdLogNotOpen")
	ErrLoggerIsClosed = errors.New("LoggerIsClosed")
)

type Handler interface {
	io.Writer
}

type Logger struct {
	handler Handler
	bufSize int
	flags   int
	level   int
	mq      chan []byte
	quit    chan bool

	waitGroup sync.WaitGroup
	lock      sync.Mutex
}

func NewLogger(handler Handler, flags int, level int) *Logger {
	if level < L_TRACE || level > L_NOTIFY {
		level = L_TRACE
	}
	mq := make(chan []byte, 1024)
	quit := make(chan bool)
	logger := &Logger{
		handler: handler,
		bufSize: 1024,
		flags:   flags,
		level:   level,
		mq:      mq,
		quit:    quit,
	}
	logger.waitGroup.Add(1)
	go logger.listen()
	go logger.onExit()
	return logger
}

func (self *Logger) onExit() {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGTSTP, syscall.SIGQUIT)
	<-c
	self.Close()
}

func (self *Logger) listen() {
	for {
		select {
		case msg := <-self.mq:
			self.handler.Write(msg)
		case <-self.quit:
			self.waitGroup.Done()
			return
		}
	}
}

func (self *Logger) Close() {
	if self.quit == nil || self.mq == nil {
		return
	}
	close(self.quit)
	self.waitGroup.Wait()
	self.quit = nil
	self.mq = nil
}

func (self *Logger) isClosed() bool {
	return self.quit == nil || self.mq == nil
}

func (self *Logger) alloc(msg string, level int) []byte {
	self.lock.Lock()
	flags := self.flags
	self.lock.Unlock()
	showDate := flags&F_DATE > 0
	size := len(msg) + 1 //多分配1字节给\n
	if showDate {
		size += 30 //预留日期时间空间
	}
	showFile := flags&F_FILE > 0
	if showFile {
		size += 300 // 预留文件名和行号的空间
	}
	showLevel := flags&F_LEVEL > 0
	if showLevel {
		size += 10 // 预留LEVEL需要空间
	}
	buf := make([]byte, 0, size)
	if showDate {
		buf = append(buf, '[')
		buf = append(buf, time.Now().Format(LOG_TIME_FORMAT)...)
		buf = append(buf, ']', ' ')
	}

	if showFile {
		_, fileName, line, ok := runtime.Caller(2)
		if ok {
			if flags&F_SHORT_FILE > 0 {
				idx := strings.LastIndex(fileName, "/")
				if idx >= 0 {
					fileName = fileName[idx+1:]
				}
			}
			buf = append(buf, fileName...)
			buf = append(buf, ':')
			buf = append(buf, fmt.Sprintf("%d", line)...)
			buf = append(buf, ' ')
		}
	}
	if showLevel {
		buf = append(buf, '[')
		buf = append(buf, LevelName[level]...)
		buf = append(buf, ']', ' ')
	}
	buf = append(buf, msg...)
	buf = append(buf, '\n')
	return buf
}

func (self *Logger) SetLevel(l int) error {
	if l < L_TRACE || l > L_NOTIFY {
		return ErrInvalidLevel
	}
	self.lock.Lock()
	self.level = l
	self.lock.Unlock()
	return nil
}

func (self *Logger) SetFlags(flags int) error {
	self.lock.Lock()
	self.flags = flags
	self.lock.Unlock()
	return nil
}

func (self *Logger) Log(level int, format string, v ...interface{}) error {
	if self.isClosed() {
		return ErrLoggerIsClosed
	}
	self.lock.Lock()
	l := self.level
	self.lock.Unlock()
	if level < l {
		return nil
	}
	self.mq <- self.alloc(fmt.Sprintf(format, v...), level)
	return nil
}

func (self *Logger) Trace(format string, v ...interface{}) error {
	return self.Log(L_TRACE, format, v...)
}

func (self *Logger) Debug(format string, v ...interface{}) error {
	return self.Log(L_DEBUG, format, v...)
}

func (self *Logger) Info(format string, v ...interface{}) error {
	return self.Log(L_INFO, format, v...)
}

func (self *Logger) Warn(format string, v ...interface{}) error {
	return self.Log(L_WARN, format, v...)
}

func (self *Logger) Error(format string, v ...interface{}) error {
	return self.Log(L_ERROR, format, v...)
}

func (self *Logger) Notify(format string, v ...interface{}) error {
	return self.Log(L_NOTIFY, format, v...)
}

// ---------- for std logger -------

var std *Logger = NewLogger(os.Stderr, F_DATE|F_LEVEL, L_INFO)

func SetHandler(handler Handler) error {
	std.lock.Lock()
	std.handler = handler
	std.lock.Unlock()
	return nil
}

func SetFlags(flags int) error {
	return std.SetFlags(flags)
}

func Flags() (int, error) {
	std.lock.Lock()
	flags := std.flags
	std.lock.Unlock()
	return flags, nil
}

func SetLevel(l int) error {
	return std.SetLevel(l)
}

func Level() (int, error) {
	std.lock.Lock()
	l := std.level
	std.lock.Unlock()
	return l, nil
}

func Trace(format string, v ...interface{}) error {
	return std.Log(L_TRACE, format, v...)
}

func Debug(format string, v ...interface{}) error {
	return std.Log(L_DEBUG, format, v...)
}

func Info(format string, v ...interface{}) error {
	return std.Log(L_INFO, format, v...)
}

func Warn(format string, v ...interface{}) error {
	return std.Log(L_WARN, format, v...)
}

func Error(format string, v ...interface{}) error {
	return std.Log(L_ERROR, format, v...)
}

func Notify(format string, v ...interface{}) error {
	return std.Log(L_NOTIFY, format, v...)
}
