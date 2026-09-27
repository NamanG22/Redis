package core

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/NamanG22/Redis/config"
)

func withTempAOF(t *testing.T) func() {
	t.Helper()
	origFile := config.AOFFile
	origStore := store
	config.AOFFile = filepath.Join(t.TempDir(), "master.aof")
	store = make(map[string]*Obj)
	return func() {
		config.AOFFile = origFile
		store = origStore
	}
}

func TestDumpAllAOFReplacesExistingFile(t *testing.T) {
	defer withTempAOF(t)()

	if err := os.WriteFile(config.AOFFile, []byte("OLD HISTORY\nSET k old\n"), 0644); err != nil {
		t.Fatal(err)
	}
	Put("k", NewObj("v", -1))
	Put("k1", NewObj("v1", -1))

	if err := DumpAllAOF(); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(config.AOFFile)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(data, []byte("OLD HISTORY")) {
		t.Fatalf("rewrite appended instead of replacing file:\n%s", data)
	}

	got, err := Decode(data)
	if err != nil {
		t.Fatalf("rewritten AOF is not valid RESP: %v\n%s", err, data)
	}
	if len(got) != 2 {
		t.Fatalf("rewritten AOF has %d commands, want 2:\n%s", len(got), data)
	}
}

func TestDumpAllAOFDoesNotGrowOnSecondRewrite(t *testing.T) {
	defer withTempAOF(t)()

	Put("k", NewObj("v", -1))
	if err := DumpAllAOF(); err != nil {
		t.Fatal(err)
	}
	first, err := os.ReadFile(config.AOFFile)
	if err != nil {
		t.Fatal(err)
	}

	if err := DumpAllAOF(); err != nil {
		t.Fatal(err)
	}
	second, err := os.ReadFile(config.AOFFile)
	if err != nil {
		t.Fatal(err)
	}
	if len(second) != len(first) {
		t.Fatalf("second rewrite grew the file from %d to %d bytes", len(first), len(second))
	}
}

func TestDumpAllAOFOmitsDeletedKeys(t *testing.T) {
	defer withTempAOF(t)()

	Put("keep", NewObj("v", -1))
	Put("gone", NewObj("v", -1))
	if err := DumpAllAOF(); err != nil {
		t.Fatal(err)
	}
	Delete("gone")
	if err := DumpAllAOF(); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(config.AOFFile)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(data, []byte("gone")) {
		t.Fatalf("deleted key still present after rewrite:\n%s", data)
	}
	if !bytes.Contains(data, []byte("keep")) {
		t.Fatalf("live key missing after rewrite:\n%s", data)
	}
}

func TestEncodeStringArray(t *testing.T) {
	got := Encode([]string{"SET", "k", "v"}, false)
	want := []byte("*3\r\n$3\r\nSET\r\n$1\r\nk\r\n$1\r\nv\r\n")
	if !bytes.Equal(got, want) {
		t.Fatalf("Encode([]string) = %q, want %q", got, want)
	}
}
