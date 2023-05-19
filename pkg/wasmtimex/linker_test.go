package wasmtimex

import (
	"fmt"
	"testing"
)

func TestLinker(t *testing.T) {
	engine := NewEngine()
	defer engine.Delete()
	store := NewStore(engine, 0, nil)
	defer store.Delete()

	ctx := store.Context()

	// Compile two wasm modules where the first references the second
	wasm1, err := Wat2Wasm(`
	(module
	  (import "wasm2" "double" (func $double (param i32) (result i32)))
	  (func (export "double_and_add") (param i32 i32) (result i32)
	    local.get 0
	    call $double
	    local.get 1
	    i32.add
	  )
	)`)
	if err != nil {
		defer err.Delete()
		t.Fatal(err.Error())
	}
	defer wasm1.Delete()

	wasm2, err := Wat2Wasm(`
	(module
	  (func (export "double") (param i32) (result i32)
	    local.get 0
	    i32.const 2
	    i32.mul
	  )
	)`)
	if err != nil {
		defer err.Delete()
		t.Fatal(err.Error())
	}
	defer wasm2.Delete()

	// Next compile both modules
	module1, err := NewModule(engine, wasm1)
	if err != nil {
		defer err.Delete()
		t.Fatal(err.Error())
	}
	module2, err := NewModule(engine, wasm2)
	if err != nil {
		defer err.Delete()
		t.Fatal(err.Error())
	}

	linker := NewLinker(engine)

	// The second module is instantiated first since it has no imports, and
	// then we insert the instance back into the linker under the name
	// the first module expects.
	instance2, trap, err := linker.Instantiate(ctx, module2)
	if trap != nil {
		defer trap.Delete()
		t.Fatal(trap.Error())
	}
	if err != nil {
		defer err.Delete()
		t.Fatal(err.Error())
	}
	err = linker.DefineInstance(ctx, "wasm2", &instance2)
	if err != nil {
		defer err.Delete()
		t.Fatal(err.Error())
	}

	// And now we can instantiate our first module, executing the result
	// afterwards
	instance1, trap, err := linker.Instantiate(ctx, module1)
	if trap != nil {
		defer trap.Delete()
		t.Fatal(trap.Error())
	}
	if err != nil {
		defer err.Delete()
		t.Fatal(err.Error())
	}
	doubleAndAddExtern, ok := instance1.ExportNamed(ctx, "double_and_add")
	if !ok {
		t.Fatal("double_and_add not exported")
	}
	doubleAndAdd := doubleAndAddExtern.Func()

	params := make([]Val, 2)
	params[0].SetI32(2)
	params[1].SetI32(3)

	paramsAndResults := make([]ValRaw, 2)
	paramsAndResults[0].SetI32(2)
	paramsAndResults[1].SetI32(3)

	trap = doubleAndAdd.CallUnchecked(ctx, paramsAndResults)
	if trap != nil {
		defer trap.Delete()
		t.Fatal(trap.Error())
	}
	fmt.Println(paramsAndResults[0].I32())

	results := make([]Val, 1)
	trap, err = doubleAndAdd.Call(ctx, params, results)
	if trap != nil {
		defer trap.Delete()
		t.Fatal(trap.Error())
	}
	if err != nil {
		defer err.Delete()
		t.Fatal(err.Error())
	}
	fmt.Println(results[0].I32())
}

func BenchmarkLinker(b *testing.B) {
	engine := NewEngineWithConfigBuilder(func(cfg *Config) {
		cfg.SetEpochInterruption(true)
		cfg.SetStrategy(StrategyCranelift)
		cfg.SetCraneliftOptLevel(2)
		cfg.SetDebugInfo(false)
		cfg.SetWasmSIMD(true)
		cfg.SetWasmBulkMemory(true)
		cfg.SetWasmMultiMemory(true)
	})
	//engine := NewEngine()
	defer engine.Delete()
	store := NewStore(engine, 0, nil)
	defer store.Delete()

	store.Context().SetEpochDeadline(500)

	ctx := store.Context()

	// Compile two wasm modules where the first references the second
	wasm1, err := Wat2Wasm(`
	(module
	  (import "wasm2" "double" (func $double (param i32) (result i32)))
	  (func (export "double_and_add") (param i32 i32) (result i32)
	    local.get 0
	    call $double
	    local.get 1
	    i32.add
	  )
	)`)
	if err != nil {
		defer err.Delete()
		b.Fatal(err.Error())
	}
	defer wasm1.Delete()

	//wasm2, err := Wat2Wasm(`
	//(module
	//(func (export "double") (param i32) (result i32)
	//  local.get 0
	//  i32.const 2
	//  i32.mul
	//)
	//)`)
	wasm2, err := Wat2Wasm(`
	(module
	(func (export "double") (param i32) (result i32)
      i32.const 2)
	)`)
	if err != nil {
		defer err.Delete()
		b.Fatal(err.Error())
	}
	defer wasm2.Delete()

	// Next compile both modules
	module1, err := NewModule(engine, wasm1)
	if err != nil {
		defer err.Delete()
		b.Fatal(err.Error())
	}
	module2, err := NewModule(engine, wasm2)
	if err != nil {
		defer err.Delete()
		b.Fatal(err.Error())
	}

	instance, trap, err := NewInstance(store.Context(), module2)
	if trap != nil {
		defer trap.Delete()
		b.Fatal(trap.Error())
	}
	doubleFunc, ok := instance.ExportNamed(store.Context(), "double")
	if ok {
		fn := doubleFunc.Func()
		params := make([]Val, 1)
		params[0].SetI32(2)
		//params[1].SetI32(3)
		results := make([]Val, 1)
		trap, err = fn.Call(store.Context(), params, results)
		if trap != nil {
			defer trap.Delete()
			b.Fatal(trap.Error())
		}
		if err != nil {
			defer err.Delete()
			b.Fatal(err.Error())
		}

		paramsAndResults := make([]ValRaw, 2)
		paramsAndResults[0].SetI32(2)
		paramsAndResults[1].SetI32(3)

		trap = doubleFunc.Func().CallUnchecked(store.Context(), paramsAndResults)
		if trap != nil {
			defer trap.Delete()
			b.Fatal(trap.Error())
		}
		if err != nil {
			defer err.Delete()
			b.Fatal(err.Error())
		}
	}

	linker := NewLinker(engine)

	// The second module is instantiated first since it has no imports, and
	// then we insert the instance back into the linker under the name
	// the first module expects.
	instance2, trap, err := linker.Instantiate(ctx, module2)
	if trap != nil {
		defer trap.Delete()
		b.Fatal(trap.Error())
	}
	if err != nil {
		defer err.Delete()
		b.Fatal(err.Error())
	}
	err = linker.DefineInstance(ctx, "wasm2", &instance2)
	if err != nil {
		defer err.Delete()
		b.Fatal(err.Error())
	}

	// And now we can instantiate our first module, executing the result
	// afterwards
	instance1, trap, err := linker.Instantiate(ctx, module1)
	if trap != nil {
		defer trap.Delete()
		b.Fatal(trap.Error())
	}
	if err != nil {
		defer err.Delete()
		b.Fatal(err.Error())
	}

	doubleExtern, ok := instance2.ExportNamed(ctx, "double")
	if !ok {
		b.Fatal("double not exported")
	}
	defer doubleExtern.Delete()
	double := doubleExtern.Func()
	doubleAndAddExtern, ok := instance1.ExportNamed(ctx, "double_and_add")
	if !ok {
		b.Fatal("double_and_add not exported")
	}
	defer doubleAndAddExtern.Delete()
	doubleAndAdd := doubleAndAddExtern.Func()

	params := make([]Val, 2)
	params[0].SetI32(2)
	params[1].SetI32(3)

	paramsAndResults := make([]ValRaw, 2)
	paramsAndResults[0].SetI32(2)
	paramsAndResults[1].SetI32(3)
	//trap = doubleAndAdd.CallUnchecked(ctx, paramsAndResults)
	//if trap != nil {
	//	defer trap.Delete()
	//	b.Fatal(trap.Error())
	//}
	//res := paramsAndResults[0].I32()
	//fmt.Println(res)
	//
	results := make([]Val, 1)
	//trap, err = doubleAndAdd.Call(ctx, params, results)
	//if trap != nil {
	//	defer trap.Delete()
	//	b.Fatal(trap.Error())
	//}
	//if err != nil {
	//	defer err.Delete()
	//	b.Fatal(err.Error())
	//}

	b.Run("checked", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			trap, err = doubleAndAdd.Call(ctx, params, results)
			if trap != nil {
				msg := trap.Error()
				trap.Delete()
				b.Fatal(msg)
			}
			if err != nil {
				msg := err.Error()
				err.Delete()
				b.Fatal(msg)
			}
		}
	})

	b.Run("unchecked double", func(b *testing.B) {
		var arg [1]ValRaw
		args := arg[0:1]
		args[0].SetI32(2)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			trap = double.CallUnchecked(ctx, args)
			//if trap != nil {
			//	msg := trap.Error()
			//	trap.Delete()
			//	b.Fatal(msg)
			//}
		}
	})

	b.Run("unchecked", func(b *testing.B) {
		var arg [2]ValRaw
		args := arg[0:2]
		args[0].SetI32(2)
		args[1].SetI32(3)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			trap = doubleAndAdd.CallUnchecked(ctx, args)
			if trap != nil {
				msg := trap.Error()
				trap.Delete()
				b.Fatal(msg)
			}
		}
	})
}
