package wasmtimex

import (
	"fmt"
	"testing"
)

func TestValType(t *testing.T) {
	NewValType(KindI32).Delete()
	NewValType(KindI64).Delete()
	NewValType(KindF32).Delete()
	NewValType(KindF64).Delete()
	NewValType(KindExternref).Delete()
	NewValType(KindFuncref).Delete()
}

func TestValTypeKind(t *testing.T) {
	if NewValType(KindI32).Kind() != KindI32 {
		t.Fatalf("wrong kind")
	}
	if NewValType(KindI64).Kind() != KindI64 {
		t.Fatalf("wrong kind")
	}
	if NewValType(KindF32).Kind() != KindF32 {
		t.Fatalf("wrong kind")
	}
	if NewValType(KindF64).Kind() != KindF64 {
		t.Fatalf("wrong kind")
	}
	if NewValType(KindExternref).Kind() != KindExternref {
		t.Fatalf("wrong kind")
	}
	if NewValType(KindFuncref).Kind() != KindFuncref {
		t.Fatalf("wrong kind")
	}
	if KindI32 == KindI64 {
		t.Fatalf("unequal kinds equal")
	}
	if KindI32 != KindI32 {
		t.Fatalf("equal kinds unequal")
	}
}

func TestValTypeVec(t *testing.T) {
	fmt.Println("sizeof", SizeofvalTypeVec())
	vec := NewValTypeVec(2)
	//defer vec.Delete()

	var (
		a = NewValType(KindI64)
		b = NewValType(KindF64)
	)
	vec2 := NewValTypeVecOf([]*ValType{a.Clone(), b.Clone()})
	vec.Set(0, a)
	vec.Set(1, b)

	slice := vec.Unsafe()
	_ = slice

	fmt.Println("a = ", a.Kind())
	fmt.Println("b = ", b.Kind())

	var (
		aa = vec.Get(0)
		bb = vec.Get(1)
	)
	fmt.Println("[0] = ", aa.Kind())
	fmt.Println("[1] = ", bb.Kind())

	fmt.Println("[0] = ", vec2.Get(0).Kind())
	fmt.Println("[1] = ", vec2.Get(1).Kind())
}
