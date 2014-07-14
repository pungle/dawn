//Copyright (C) Mr.Pungle

package web_test

import (
	"net/http"
	"testing"
)

type fakeResp struct {
	Data string
}

func (rf *fakeResp) Write(bytes []byte) (int, error) {
	rf.Data += string(bytes)
	return len(bytes), nil
}

func (rf *fakeResp) WriteHeader(code int) {
	//TODO:..
}

func (rf *fakeResp) Header() http.Header {
	return nil
}

func fakeHandler1(resp http.ResponseWriter, req *http.Request) {
	resp.Write([]byte("1"))
}

func fakeHandler2(resp http.ResponseWriter, req *http.Request) {
	resp.Write([]byte("2"))
}

func fakeHandler3(resp http.ResponseWriter, req *http.Request) {
	resp.Write([]byte("3"))
}

func fakeHandler4(resp http.ResponseWriter, req *http.Request) {
	resp.Write([]byte("4"))
}

func TestMappingResolver(t *testing.T) {
	resolver := NewMappingResolver()
	if resolver == nil {
		t.Error("Create MappingResolver Error")
	}
	uri := "abcdefa"
	resolver.AddHandler(uri, fakeHandler1)
	resolver.AddHandler(uri+"2", fakeHandler2)
	resolver.AddHandler(uri+"3", fakeHandler3)
	resolver.AddHandler(uri+"4", fakeHandler4)

	hand, err := resolver.Resolve(uri + "5")
	if hand != nil {
		t.Error("Resolve error: Found the error handler.")
	}
	if err != nil {
		t.Error("Resolve occurred error: ", err.Error())
	}

	hand, err = resolver.Resolve(uri + "3")
	if hand == nil {
		t.Error("Resolve error: Can not found the real handler.")
	}
	if err != nil {
		t.Error("Resolve occurred error: ", err.Error())
	}

	resp := &fakeResp{}
	hand(resp, nil)
	if resp.Data != "3" {
		t.Error("Resolve error: Found the error handler.", resp.Data)
	}
}

func TestPrefixResolver(t *testing.T) {
	resolver := NewPrefixResolver()
	if resolver == nil {
		t.Error("Create PrefixResolver Error")
	}
}

func TestRegexpResolver(t *testing.T) {
	resolver := NewRegexpResolver()
	if resolver == nil {
		t.Error("Create RegexpResolver Error")
	}
}
