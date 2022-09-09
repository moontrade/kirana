package wasmtimex

import (
	"testing"
)

func TestTrap(t *testing.T) {
	trap := NewTrap("message")
	defer trap.Delete()
	m := trap.Message()
	if m.ToOwned() != "message" {
		panic("wrong message")
	}
}

func TestTrapFrames(t *testing.T) {
	engine := NewEngine()
	defer engine.Delete()
	store := NewStore(engine, 0, nil)
	wasm, err := Wat2Wasm(`
	  (func call $foo)
	  (func $foo call $bar)
	  (func $bar unreachable)
	  (start 0)
	`)

	if err != nil {
		defer err.Delete()
		t.Fatal(err.Error())
	}
	module, err := NewModule(engine, wasm)
	defer module.Delete()
	if err != nil {
		defer err.Delete()
		t.Fatal(err.Error())
	}

	_, trap, err := NewInstance(store.Context(), module)
	if err != nil {
		defer err.Delete()
		t.Fatal(err.Error())
	}
	if trap == nil {
		panic("expected trap")
	}
	defer trap.Delete()

	frames_ := trap.Frames()
	defer frames_.Delete()
	frames := frames_.Unsafe()
	if len(frames) != 3 {
		panic("expected 3 frames")
	}
	if frames[0].FuncName() != "bar" {
		panic("bad function name")
	}
	if frames[1].FuncName() != "foo" {
		panic("bad function name")
	}
	if frames[2].FuncName() != "" {
		panic("bad function name")
	}
	if frames[0].FuncIndex() != 2 {
		panic("bad function index")
	}
	if frames[1].FuncIndex() != 1 {
		panic("bad function index")
	}
	if frames[2].FuncIndex() != 0 {
		panic("bad function index")
	}

	expected := `wasm trap: wasm ` + "`unreachable`" + ` instruction executed
wasm backtrace:
    0:   0x26 - <unknown>!bar
    1:   0x21 - <unknown>!foo
    2:   0x1c - <unknown>!<wasm function 0>
`
	actual := trap.Error()
	if actual != expected {
		t.Fatalf("expected\n%s\ngot\n%s", expected, actual)
	}
	expCode := UnreachableCodeReached
	if code, ok := trap.Code(); !ok && code != expCode {
		t.Fatalf("expected %v got %v", expCode, code)
	}
}

//
//func TestTrapModuleName(t *testing.T) {
//	store := NewStore(NewEngine())
//	wasm, err := Wat2Wasm(`(module $f
//	  (func unreachable)
//	  (start 0)
//	)`)
//	assertNoError(err)
//	module, err := NewModule(store.Engine, wasm)
//	assertNoError(err)
//
//	i, err := NewInstance(store, module, []AsExtern{})
//	if i != nil {
//		panic("expected failure")
//	}
//	if err == nil {
//		panic("expected failure")
//	}
//	trap := err.(*Trap)
//	frames := trap.Unsafe()
//	if len(frames) != 1 {
//		panic("expected 3 frames")
//	}
//	if *frames[0].ModuleName() != "f" {
//		panic("bad module name")
//	}
//}
