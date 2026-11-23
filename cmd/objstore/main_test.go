package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestHelpAndVersion(t *testing.T) {
	t.Parallel()
	var out, errb bytes.Buffer
	if code := run([]string{"--help"}, &out, &errb); code != 0 || !strings.Contains(out.String(), "usage:") {
		t.Fatalf("help: code %d, out %q", code, out.String())
	}
	out.Reset()
	if code := run([]string{"--version"}, &out, &errb); code != 0 || !strings.HasPrefix(out.String(), "objstore ") {
		t.Fatalf("version: code %d, out %q", code, out.String())
	}
}

func TestStubExitsNonZero(t *testing.T) {
	t.Parallel()
	var out, errb bytes.Buffer
	if code := run([]string{"verify"}, &out, &errb); code == 0 || !strings.Contains(errb.String(), "not implemented yet: see #7") {
		t.Fatalf("stub: code %d, err %q", code, errb.String())
	}
}
