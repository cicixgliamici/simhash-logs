package tokenize

import (
	"reflect"
	"testing"
)

func TestSimple_KeepsPlaceholders(t *testing.T) {
	in := "sshd failed from <ip> port <num> ssh2"
	got := Simple(in)
	want := []string{"sshd", "failed", "from", "<ip>", "port", "<num>", "ssh2"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got=%v want=%v", got, want)
	}
}

func TestSimple_SplitsOnPunctuation(t *testing.T) {
	in := "kernel: [123] eth0: link-up!"
	got := Simple(in)
	want := []string{"kernel", "123", "eth0", "link", "up"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got=%v want=%v", got, want)
	}
}

func TestSimple_EmptyInput(t *testing.T) {
	if got := Simple("   "); got != nil {
		t.Fatalf("expected nil for empty input, got=%v", got)
	}
}