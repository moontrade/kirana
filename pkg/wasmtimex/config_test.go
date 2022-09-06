package wasmtimex

import "testing"

func TestConfig(t *testing.T) {
	cfg := NewConfig()
	defer cfg.Close()
	cfg.SetDebugInfo(true)
	cfg.SetWasmThreads(true)
	cfg.SetWasmReferenceTypes(true)
	cfg.SetWasmSIMD(true)
	cfg.SetWasmBulkMemory(true)
	cfg.SetWasmMultiValue(true)
	cfg.SetWasmMultiMemory(true)
	cfg.SetConsumeFuel(true)
	cfg.SetStrategy(StrategyAuto)
	cfg.SetStrategy(StrategyCranelift)
	cfg.SetCraneliftDebugVerifier(true)
	cfg.SetCraneliftOptLevel(OptLevelNone)
	cfg.SetCraneliftOptLevel(OptLevelSpeed)
	cfg.SetCraneliftOptLevel(OptLevelSpeedAndSize)
	cfg.SetProfiler(ProfilingStrategyNone)
	err := cfg.CacheConfigLoadDefault()
	if err != nil {
		panic(err)
	}
	err = cfg.CacheConfigLoad("nonexistent.toml")
	if err == nil {
		panic("expected an error")
	}
}
