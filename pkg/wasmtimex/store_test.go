package wasmtimex

import (
	"fmt"
	"testing"
)

func TestNewStore(t *testing.T) {
	engine := NewEngine()
	store := NewStore(engine, 0, nil)
	ctx := store.Context()
	fmt.Println(ctx)
	store.Delete()
}
