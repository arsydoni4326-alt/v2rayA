package conf

import (
	"errors"
	"testing"
)

func TestIsGoTestFlagError(t *testing.T) {
	if !isGoTestFlagError(errors.New("unexpected word while parsing flags: '-test.paniconexit0'")) {
		t.Fatal("expected Go test flag error to be recognized")
	}
	if isGoTestFlagError(errors.New("unexpected word while parsing flags: '--address'")) {
		t.Fatal("application flag error must not be ignored")
	}
	if isGoTestFlagError(nil) {
		t.Fatal("nil error must not be recognized as a Go test flag error")
	}
}
