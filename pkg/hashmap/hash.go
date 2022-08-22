package hashmap

import "github.com/moontrade/wormhole/pkg/wyhash"

type HasherFunc[K any] func(key K) uint64

func HashInt(key int) uint64 {
	return wyhash.I64(int64(key))
}

func HashInt64(key int64) uint64 {
	return wyhash.I64(key)
}

func HashString(key string) uint64 {
	return wyhash.String(key)
}

func HashUint64(key uint64) uint64 {
	return wyhash.U64(key)
}
