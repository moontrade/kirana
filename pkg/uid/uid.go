package uid

import (
	crand "crypto/rand"
	"encoding/binary"
	"github.com/moontrade/wormhole/pkg/wyhash"
	"math/big"
)

func UID() uint64 {
	return wyhash.Next()
}

// CryptoUID generates a random string for cryptographic usage.
func CryptoUID(n int, runes string) (string, error) {
	letters := []rune(runes)
	b := make([]rune, n)
	for i := range b {
		v, err := crand.Int(crand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		b[i] = letters[v.Int64()]
	}
	return string(b), nil
}

// CryptoU64 returns cryptographic random uint64.
func CryptoU64() (uint64, error) {
	var v uint64
	if err := binary.Read(crand.Reader, binary.LittleEndian, &v); err != nil {
		return 0, err
	}
	return v, nil
}
