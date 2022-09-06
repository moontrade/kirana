package wasmtimex

import "testing"

func TestEngine(t *testing.T) {
	NewEngine().Close()
	NewEngineWithConfig(func(cfg *Config) {
		cfg.SetEpochInterruption(true)
	}).Close()
}
