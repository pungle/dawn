//Copyright (C) Mr.Pungle

package logging

import (
	"sync"
	"testing"
)

type fakeHandler struct {
	alloc   [10240]byte
	content []byte
	isclose bool
}

func (self *fakeHandler) Write(content []byte) (int, error) {
	if self.content == nil {
		self.content = self.alloc[:0]
	}
	self.content = append(self.content, content...)
	return len(content), nil
}

func (self *fakeHandler) Close() error {
	self.isclose = true
	return nil
}

func TestBuffWriter(t *testing.T) {
	fake := &fakeHandler{}
	data := []byte("1234567890abcde")
	buffSize := 512
	buff := NewBuffHandler(fake, buffSize)
	var wg sync.WaitGroup
	wg.Add(1000)
	for n := 0; n < 1000; n++ {
		go func() {
			for i := 0; i < 1000; i++ {
				buff.Write(data)
			}
			wg.Done()
		}()
	}
	wg.Wait()
	buff.Close()

	if !fake.isclose {
		t.Error("handler have not close.")
	}

	if len(fake.content) != len(data)*1000*1000 {
		t.Error("Write Size not match", len(fake.content), len(data)*1000*1000)
	}
}

func BenchmarkWriterSingleChar(t *testing.B) {
	fake := &fakeHandler{}
	data := []byte("1234")
	buffSize := 1024 * 1024
	buff := NewBuffHandler(fake, buffSize)
	var wg sync.WaitGroup
	wg.Add(1000)
	for n := 0; n < 1000; n++ {
		go func() {
			for i := 0; i < t.N; i++ {
				buff.Write(data)
			}
			wg.Done()
		}()
	}
	wg.Wait()
	buff.Close()
}
