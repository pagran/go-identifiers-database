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
	//go:embed dataset.txt.gz
	dataGetNames []byte
	//go:embed dataset.txt.gz.idx
	idxGetNames []byte

	onceGetNames     sync.Once
	unpackedGetNames string
	arrGetNames      []string
)

func GetNames(typ NameType) []string {
	onceGetNames.Do(func() {
		var buf bytes.Buffer
		buf.Grow(58636055)

		reader, err := gzip.NewReader(bytes.NewReader(dataGetNames))
		if err != nil {
			panic(err)
		}

		if _, err := buf.ReadFrom(reader); err != nil {
			panic(err)
		}

		b := buf.Bytes()

		slice := (*reflect.SliceHeader)(unsafe.Pointer(&b))
		str := (*reflect.StringHeader)(unsafe.Pointer(&unpackedGetNames))
		str.Data = slice.Data
		str.Len = slice.Len

		arrGetNames = make([]string, len(idxGetNames))
		offset := 0
		for i, l := range idxGetNames {
			strLen := int(l)
			arrGetNames[i] = unpackedGetNames[offset : offset+strLen]
			offset += strLen
		}
	})

	if typ == NameType(4) {
		return arrGetNames[0:321571]
	}

	if typ == NameType(5) {
		return arrGetNames[321571:821285]
	}

	if typ == NameType(6) {
		return arrGetNames[821285:1339606]
	}

	if typ == NameType(7) {
		return arrGetNames[1339606:1630952]
	}

	if typ == NameType(1) {
		return arrGetNames[1630952:1757731]
	}

	if typ == NameType(2) {
		return arrGetNames[1757731:1792966]
	}

	if typ == NameType(3) {
		return arrGetNames[1792966:2956371]
	}

	return arrGetNames
}
