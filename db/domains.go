package db

import (
	_ "embed"
	"sync"
)

var (
	//go:embed domains.txt
	dataGetDomains string
	//go:embed domains.txt.idx
	idxGetDomains []byte

	onceGetDomains sync.Once
	arrGetDomains  []string
)

func GetDomains() []string {
	onceGetDomains.Do(func() {
		arrGetDomains = make([]string, len(idxGetDomains))
		offset := 0
		for i, l := range idxGetDomains {
			strLen := int(l)
			arrGetDomains[i] = dataGetDomains[offset : offset+strLen]
			offset += strLen
		}
	})
	return arrGetDomains
}
