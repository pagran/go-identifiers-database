package db

import (
	_ "embed"
	"sync"
)

var (
	//go:embed identifiers.txt
	dataGetIdentifiers string
	//go:embed identifiers.txt.idx
	idxGetIdentifiers []byte

	onceGetIdentifiers sync.Once
	arrGetIdentifiers  []string
)

func GetIdentifiers() []string {
	onceGetIdentifiers.Do(func() {
		arrGetIdentifiers = make([]string, len(idxGetIdentifiers))
		offset := 0
		for i, l := range idxGetIdentifiers {
			strLen := int(l)
			arrGetIdentifiers[i] = dataGetIdentifiers[offset : offset+strLen]
			offset += strLen
		}
	})
	return arrGetIdentifiers
}
