package wasmtimex

func boolToU32(v bool) uint32 {
	if v {
		return 1
	}
	return 0
}

func boolToU64(v bool) uint64 {
	if v {
		return 1
	}
	return 0
}
