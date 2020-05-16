package geebolt

import "os"

type DB struct {
	data []byte
	file *os.File
}

const maxMapSize = 1 << 31

func (db *DB) mmap(sz int) error {
	b, err := syscall.Mmap()
}

func Open(path string) {

}
