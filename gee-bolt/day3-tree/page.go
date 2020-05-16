package geebolt

import (
	"fmt"
	"reflect"
	"unsafe"
)

const pageHeaderSize = unsafe.Sizeof(page{})
const branchPageElementSize = unsafe.Sizeof(branchPageElement{})
const leafPageElementSize = unsafe.Sizeof(leafPageElement{})
const maxKeysPerPage = 1024

const (
	branchPageFlag uint16 = iota
	leafPageFlag
	metaPageFlag
	freelistPageFlag
)

type page struct {
	id       uint64
	flags    uint16
	count    uint16
	overflow uint32
}

type leafPageElement struct {
	pos   uint32
	ksize uint32
	vsize uint32
}

type branchPageElement struct {
	pos   uint32
	ksize uint32
	pgid  uint64
}

func (p *page) typ() string {
	switch p.flags {
	case branchPageFlag:
		return "branch"
	case leafPageFlag:
		return "leaf"
	case metaPageFlag:
		return "meta"
	case freelistPageFlag:
		return "freelist"
	}
	return fmt.Sprintf("unknown<%02x>", p.flags)
}

func (p *page) meta() *meta {
	return (*meta)(unsafe.Pointer(uintptr(unsafe.Pointer(p)) + pageHeaderSize))
}

func (p *page) dataPtr() unsafe.Pointer {
	return unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(p)) + pageHeaderSize,
		Len:  int(p.count),
		Cap:  int(p.count),
	})
}

func (p *page) leafPageElement(index uint16) *leafPageElement {
	off := pageHeaderSize + uintptr(index)*leafPageElementSize
	return (*leafPageElement)(unsafe.Pointer(uintptr(unsafe.Pointer(p)) + off))
}

func (p *page) leafPageElements() []leafPageElement {
	if p.count == 0 {
		return nil
	}
	return *(*[]leafPageElement)(p.dataPtr())
}

func (p *page) branchPageElement(index uint16) *branchPageElement {
	off := pageHeaderSize + uintptr(index)*branchPageElementSize
	return (*branchPageElement)(unsafe.Pointer(uintptr(unsafe.Pointer(p)) + off))
}

func (p *page) branchPageElements() []branchPageElement {
	if p.count == 0 {
		return nil
	}
	return *(*[]branchPageElement)(p.dataPtr())
}
