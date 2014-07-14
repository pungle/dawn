//Copyright (C) Mr.Pungle

package logging

import (
	"sync"
)

const PERSIST_SIZE = 512 //缓存l区阀值

type BufferHandler struct {
	handler  Handler
	buffer   []byte
	capacity int
	length   int

	lock sync.Mutex
}

func NewBuffHandler(handler Handler, capacity int) Handler {
	return &BufferHandler{
		handler:  handler,
		buffer:   make([]byte, 0, capacity),
		capacity: capacity + PERSIST_SIZE,
	}
}

// 将写入缓存区, 当缓存区大小缓存区容量阀值时写入handler
func (self *BufferHandler) Write(msg []byte) (cnt int, err error) {
	cnt = len(msg)
	self.lock.Lock()
	self.length += cnt
	self.buffer = append(self.buffer, msg...)
	if self.length+PERSIST_SIZE >= self.capacity {
		_, err = self.handler.Write(self.buffer)
		self.buffer = self.buffer[:0]
		self.length = 0
	}
	self.lock.Unlock()
	return
}

func (self *BufferHandler) Close() error {
	self.lock.Lock()
	self.handler.Write(self.buffer)
	self.lock.Unlock()
	return self.handler.Close()
}
