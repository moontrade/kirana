package util

// Sets the bit at pos in the integer n.
func SetBit64(n uint64, pos uint64) uint64 {
	n |= (1 << pos)
	return n
}

// Clears the bit at pos in n.
func ClearBit64(n uint64, pos uint64) uint64 {
	mask := ^(uint64(1) << pos)
	n &= mask
	return n
}

func HasBit64(n uint64, pos uint64) bool {
	val := n & (1 << pos)
	return (val > 0)
}

// Sets the bit at pos in the integer n.
func SetBit16(n uint16, pos uint16) uint16 {
	n |= (1 << pos)
	return n
}

// Clears the bit at pos in n.
func ClearBit16(n uint16, pos uint16) uint16 {
	mask := ^(uint16(1) << pos)
	n &= mask
	return n
}

func HasBit16(n uint16, pos uint16) bool {
	val := n & (1 << pos)
	return (val > 0)
}

// Sets the bit at pos in the integer n.
func SetBit32(n uint32, pos uint32) uint32 {
	n |= (1 << pos)
	return n
}

// Clears the bit at pos in n.
func ClearBit32(n uint32, pos uint32) uint32 {
	mask := ^(uint32(1) << pos)
	n &= mask
	return n
}

func HasBit32(n uint32, pos uint32) bool {
	val := n & (1 << pos)
	return (val > 0)
}
