package wasmtimex

import (
	"fmt"
	"testing"
)

func TestNewMemory(t *testing.T) {
	engine := NewEngine()
	defer engine.Delete()
	store := NewStore(engine, 0, nil)
	ctx := store.Context()
	ty := NewMemoryType(1, true, 4, false)
	memory, err := NewMemory(ctx, ty)
	if err != nil {
		defer err.Delete()
		t.Fatal(err.Error())
	}

	ty2 := memory.Type(ctx)

	_ = ty2
	fmt.Println("size", memory.Size(ctx))

	data := memory.Data(ctx)
	fmt.Println("data size", len(data))

	prevSize, err := memory.Grow(ctx, 2)
	if err != nil {
		defer err.Delete()
		t.Fatal(err.Error())
	}

	data = memory.Data(ctx)

	fmt.Println("prev size", prevSize)
	fmt.Println("new data size", len(data))
}
