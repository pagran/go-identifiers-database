package db

import (
	_ "embed"
	"sync"
)

var (
	//go:embed filenames.txt
	dataGetFilenames string
	//go:embed filenames.txt.idx
	idxGetFilenames []byte

	onceGetFilenames sync.Once
	arrGetFilenames  []string
)

func GetFilenames() []string {
	onceGetFilenames.Do(func() {
		arrGetFilenames = make([]string, len(idxGetFilenames))
		offset := 0
		for i, l := range idxGetFilenames {
			strLen := int(l)
			arrGetFilenames[i] = dataGetFilenames[offset : offset+strLen]
			offset += strLen
		}
	})
	return arrGetFilenames
}
