package db

import (
	_ "embed"
	"sync"
)

var (
	//go:embed packages.txt
	dataGetPackages string
	//go:embed packages.txt.idx
	idxGetPackages []byte

	onceGetPackages sync.Once
	arrGetPackages  []string
)

func GetPackages() []string {
	onceGetPackages.Do(func() {
		arrGetPackages = make([]string, len(idxGetPackages))
		offset := 0
		for i, l := range idxGetPackages {
			strLen := int(l)
			arrGetPackages[i] = dataGetPackages[offset : offset+strLen]
			offset += strLen
		}
	})
	return arrGetPackages
}
