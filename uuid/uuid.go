//Copyright (C) Mr.Pungle

package uuid

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

type UUID [16]byte

func NewUUID() *UUID {
	u := &UUID{}
	_, err := rand.Read(u[:16])
	if err != nil {
		panic(err) // TODO: panic??
	}
	u[8] = (u[8] | 0x80) & 0xBf
	u[6] = (u[6] | 0x40) & 0x4f
	return u
}

func (self *UUID) String() string {
	return fmt.Sprintf("%x-%x-%x-%x-%x", self[:4], self[4:6], self[6:8], self[8:10], self[10:])
}

func (self *UUID) Base64() string {
	return base64.URLEncoding.EncodeToString(self[:16])
}

func (self *UUID) HexString() string {
	return fmt.Sprintf("%x%x%x%x%x", self[:4], self[4:6], self[6:8], self[8:10], self[10:])
}
