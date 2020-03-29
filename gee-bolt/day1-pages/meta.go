package geebolt

import (
	"errors"
	"hash/fnv"
	"unsafe"
)

// Represent a marker value to indicate that a file is a gee-bolt DB
const magic uint32 = 0xED0CDAED

type meta struct {
	magic    uint32
	pageSize uint32
	pgid     uint64
	checksum uint64
}

func (m *meta) sum64() uint64 {
	var h = fnv.New64a()
	_, _ = h.Write((*[unsafe.Offsetof(meta{}.checksum)]byte)(unsafe.Pointer(m))[:])
	return h.Sum64()
}

func (m *meta) validate() error {
	if m.magic != magic {
		return errors.New("invalid magic number")
	}
	if m.checksum != m.sum64() {
		return errors.New("invalid checksum")
	}
	return nil
}
