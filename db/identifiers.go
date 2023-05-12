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
	//go:embed identifiers.txt.gz
	dataGetIdentifiers []byte
	//go:embed identifiers.txt.gz.idx
	idxGetIdentifiers []byte

	onceGetIdentifiers     sync.Once
	unpackedGetIdentifiers string
	arrGetIdentifiers      []string
)

func GetIdentifiers() []string {
	onceGetIdentifiers.Do(func() {
		var buf bytes.Buffer
		buf.Grow(21021613)

		reader, err := gzip.NewReader(bytes.NewReader(dataGetIdentifiers))
		if err != nil {
			panic(err)
		}

		if _, err := buf.ReadFrom(reader); err != nil {
			panic(err)
		}

		b := buf.Bytes()

		slice := (*reflect.SliceHeader)(unsafe.Pointer(&b))
		str := (*reflect.StringHeader)(unsafe.Pointer(&unpackedGetIdentifiers))
		str.Data = slice.Data
		str.Len = slice.Len

		arrGetIdentifiers = make([]string, len(idxGetIdentifiers))
		offset := 0
		for i, l := range idxGetIdentifiers {
			strLen := int(l)
			arrGetIdentifiers[i] = unpackedGetIdentifiers[offset : offset+strLen]
			offset += strLen
		}
	})
	return arrGetIdentifiers
}
