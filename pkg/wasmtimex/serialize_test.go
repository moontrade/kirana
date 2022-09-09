package wasmtimex

import (
	"fmt"
	"testing"
	"time"
)

func serialize(engine *Engine) []byte {
	// Compile the wasm module into an in-memory instance of a `Module`.
	fmt.Println("Compiling module...")
	wasm, err := Wat2Wasm(`
	(module
	  (func $hello (import "" "hello"))
	  (func (export "run") (call $hello))
	)`)
	if err != nil {
		panic(err)
	}
	module, err := NewModule(engine, wasm)
	if err != nil {
		panic(err)
	}
	serialized, err := module.Serialize()
	if err != nil {
		panic(err)
	}
	defer serialized.Delete()

	fmt.Println("Serialized.")
	return serialized.Bytes()
}

func deserialize(engine *Engine, encoded []byte) {
	// Configure the initial compilation environment.
	fmt.Println("Initializing...")
	store := NewStore(engine, 0, nil)
	ctx := store.Context()
	ctx.SetEpochDeadline(NewEpochDeadline(time.Second))

	// Deserialize the compiled module.
	fmt.Println("Deserialize module...")
	module, err := Deserialize(engine, encoded)
	if err != nil {
		defer err.Delete()
		panic(err.Error())
	}

	// Here we handle the imports of the module, which in this case is our
	// `helloFunc` callback.
	fmt.Println("Creating callback...")
	//helloFunc := WrapFunc(store, func() {
	//	fmt.Println("Calling back...")
	//	fmt.Println("> Hello World!")
	//})

	helloFunc := NewFunc(ctx, NewFuncTypeZeroZero(), CallbackStub, 0, nil)

	// Once we've got that all set up we can then move to the instantiation
	// phase, pairing together a compiled module as well as a set of imports.
	// Note that this is where the wasm `start` function, if any, would run.
	fmt.Println("Instantiating module...")
	instance, trap, err := NewInstance(ctx, module, helloFunc.AsExtern())
	if err != nil {
		defer err.Delete()
		panic(err.Error())
	}
	if trap != nil {
		defer trap.Delete()
		panic(trap.Error())
	}

	// Next we poke around a bit to extract the `run` function from the module.
	fmt.Println("Extracting export...")
	runExport, ok := instance.ExportNamed(store.Context(), "run")
	if !ok {
		panic("func named 'run' not found")
	}
	run := runExport.Func()
	if run == nil {
		panic("Failed to find function export `run`")
	}

	// And last but not least we can call it!
	fmt.Println("Calling export...")
	trap, err = run.Call(store.Context(), nil, nil)
	if err != nil {
		defer err.Delete()
		panic(err.Error())
	}
	if trap != nil {
		defer trap.Delete()
		panic(trap.Error())
	}

	fmt.Println("Done.")
}

func TestSerialize(t *testing.T) {
	// Configure the initial compilation environment.
	fmt.Println("Initializing...")
	engine := NewEngine()
	defer engine.Delete()
	bytes := serialize(engine)
	deserialize(engine, bytes)

	// Output:
	// Initializing...
	// Compiling module...
	// Serialized.
	// Initializing...
	// Deserialize module...
	// Creating callback...
	// Instantiating module...
	// Extracting export...
	// Calling export...
	// Calling back...
	// > Hello World!
	// Done.
}
