//Copyright (C) Mr.Pungle

package uuid

import (
	"testing"
)

func TestNewUUID(t *testing.T) {
	u := NewUUID()
	t.Log(u.Base64())
	t.Log(u.HexString())
	t.Log(u.String())
}

func BenchmarkNewUUID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewUUID().Base64()
	}
}

func BenchmarkNewStringUUID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewUUID().String()
	}
}

func BenchmarkNewHexStringUUID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewUUID().HexString()
	}
}
