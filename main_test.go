package main

import (
	"bufio"
	"bytes"
	"context"
	"strings"
	"testing"
)

type FakeDbService struct {
	CreateDbCallCount int
	DropDbCallCount   int
}

func (svc *FakeDbService) CreateDb() error {
	svc.CreateDbCallCount++
	return nil
}

func (svc *FakeDbService) DropDb() error {
	svc.DropDbCallCount++
	return nil
}

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

func TestCreateDb(t *testing.T) {
	fake := &FakeDbService{}
	ctx := context.WithValue(context.Background(), "svc", fake)
	buf := bytes.NewBuffer([]byte{})
	w := bufio.NewWriter(buf)
	CreateDb(ctx, w)
	if fake.CreateDbCallCount != 1 {
		t.Errorf("Expected 1 calls to create, have %d", fake.CreateDbCallCount)
	}
}

func TestWrappedReader(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	w := bufio.NewWriter(buf)
	r := bufio.NewReader(buf)

	w.WriteString("testing\n")
	w.Flush()
	w.WriteString("again\n")
	w.Flush()
	textCh := WrappedReader(r)
	result, ok := <-textCh
	if !ok {
		t.Error("Channel read failed")
	}
	if result != "testing" {
		t.Errorf("Expected testing, have %s", result)
	}

	result, ok = <-textCh
	if !ok {
		t.Error("Channel is closed")
		return
	}
	if result != "again" {
		t.Errorf("Expected testing, have %s", result)
		return
	}
}
