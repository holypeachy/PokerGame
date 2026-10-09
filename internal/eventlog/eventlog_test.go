package eventlog

import (
	"errors"
	"io"
	"os"
	"testing"

	"pokergame"
)

//TODO: Check cases

type failingWriter struct{ err error }

func (w failingWriter) Write([]byte) (int, error) { return 0, w.err }

func TestOnEventReportsWriterFailure(t *testing.T) {
	// Stderr is process-global, so this test must not run in parallel.
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	oldStderr := os.Stderr
	os.Stderr = writer
	t.Cleanup(func() { os.Stderr = oldStderr; reader.Close(); writer.Close() })
	l := New(failingWriter{err: errors.New("write failure")})
	l.OnEvent(pokergame.Event{Type: pokergame.HoleCardsDealt, HandNumber: 2, Street: pokergame.Preflop})
	os.Stderr = oldStderr
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	got, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	want := "event log failed: write failure\n2-Preflop: HoleCardsDealt\n\n"
	if string(got) != want {
		t.Errorf("expected stderr fallback %q, got %q", want, got)
	}
}
