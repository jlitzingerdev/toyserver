package main

import (
	"bufio"
	"bytes"
	"strings"
	"testing"
)

// Input provided to WriteAndFlush is written and flushed to
// the underlying buffer
func TestWriteAndFlush(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	w := bufio.NewWriter(buf)
	WriteAndFlush(w, "foo\n")
	if w.Buffered() != 0 {
		t.Errorf("Expect 0 bytes buffered, have=%d", w.Available())
	}
	if r, _ := buf.ReadBytes('\n'); strings.TrimSpace(string(r)) != "foo" {
		t.Errorf("Expect foo, got=%s", strings.TrimSpace(string(r)))
	}
}
