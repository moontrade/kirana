package aof

func Open(name string, geometry Geometry, recovery Recovery) (aof *AOF, err error) {
	return instance.Open(name, geometry, recovery)
}
