package main

import (
	"bufio"
	"bytes"
	"strings"
	"testing"
)

type FakeDbService struct {
	CreateDbCallCount    int
	DropDbCallCount      int
	CreateTableCallCount int
	DropTableCallCount   int
	InsertTextCallCount  int
}

func (svc *FakeDbService) CreateDb() error {
	svc.CreateDbCallCount++
	return nil
}

func (svc *FakeDbService) DropDb() error {
	svc.DropDbCallCount++
	return nil
}

func (svc *FakeDbService) CreateTable() error {
	svc.CreateTableCallCount++
	return nil
}

func (svc *FakeDbService) DropTable() error {
	svc.DropTableCallCount++
	return nil
}

func (svc *FakeDbService) InsertText(text string) error {
	svc.InsertTextCallCount++
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

	CreateDb(fake)
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
	if result[0] != "testing" {
		t.Errorf("Expected testing, have %s", result)
	}

	result, ok = <-textCh
	if !ok {
		t.Error("Channel is closed")
		return
	}
	if result[0] != "again" {
		t.Errorf("Expected testing, have %s", result)
		return
	}
}
