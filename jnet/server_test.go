package jnet

import "testing"

func TestServerV1(t *testing.T) {
	s := NewServer("[Jinx Server Start] V1")
	s.Serve()
}
