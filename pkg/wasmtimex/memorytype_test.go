package wasmtimex

import "testing"

func TestNewMemoryType(t *testing.T) {
	ty := NewMemoryType(1, true, 4, false)
	if ty.Minimum() != 1 {
		t.Error("expected minimum to be 1 instead found ", ty.Minimum())
	}
	if ty.Maximum() != 4 {
		t.Error("expected maximum to be 4 instead found ", ty.Maximum())
	}
}
