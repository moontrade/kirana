package wasmtimex

import (
	"testing"
)

func TestEngine(t *testing.T) {
	NewEngine().Delete()
	NewEngineWithConfigBuilder(func(cfg *Config) {
		cfg.SetEpochInterruption(true)
	}).Delete()
}
