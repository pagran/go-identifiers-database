package db

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"reflect"
	"sync"
	"unsafe"
)

var (
	//go:embed filenames.txt.gz
	dataGetFilenames []byte
	//go:embed filenames.txt.gz.idx
	idxGetFilenames []byte

	onceGetFilenames     sync.Once
	unpackedGetFilenames string
	arrGetFilenames      []string
)

func GetFilenames() []string {
	onceGetFilenames.Do(func() {
		var buf bytes.Buffer
		buf.Grow(1461159)

		reader, err := gzip.NewReader(bytes.NewReader(dataGetFilenames))
		if err != nil {
			panic(err)
		}

		if _, err := buf.ReadFrom(reader); err != nil {
			panic(err)
		}

		b := buf.Bytes()

		slice := (*reflect.SliceHeader)(unsafe.Pointer(&b))
		str := (*reflect.StringHeader)(unsafe.Pointer(&unpackedGetFilenames))
		str.Data = slice.Data
		str.Len = slice.Len

		arrGetFilenames = make([]string, len(idxGetFilenames))
		offset := 0
		for i, l := range idxGetFilenames {
			strLen := int(l)
			arrGetFilenames[i] = unpackedGetFilenames[offset : offset+strLen]
			offset += strLen
		}
	})
	return arrGetFilenames
}
