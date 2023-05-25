package main

import (
	"bufio"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/template"
)

var (
	inputFlag   = flag.String("i", "", "Path to input file")
	outputFlag  = flag.String("o", "", "Path to output file (without extension)")
	packageFlag = flag.String("p", "", "Package name")
	methodFlag  = flag.String("m", "", "Getter method name")
	enumFlag    = flag.String("e", "", "Enum name")
)

const srcCode = `package {{ .package }}

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"reflect"
	"sync"
	"unsafe"
)

var (
	//go:embed {{ .data }}
	data{{ .method }} []byte
	//go:embed {{ .idx }}
	idx{{ .method }} []byte

	once{{ .method }} sync.Once
	unpacked{{ .method }} string
	arr{{ .method }}  []string
)

func {{ .method }}(typ {{ .enum }}) []string {
	once{{ .method }}.Do(func() {
		var buf bytes.Buffer
		buf.Grow({{ .size }})

		reader, err := gzip.NewReader(bytes.NewReader(data{{ .method }}))
		if err != nil {
			panic(err)
		}

		if _, err := buf.ReadFrom(reader); err != nil {
			panic(err)
		}

		b := buf.Bytes()

		slice := (*reflect.SliceHeader)(unsafe.Pointer(&b))
		str := (*reflect.StringHeader)(unsafe.Pointer(&unpacked{{ .method }}))
		str.Data = slice.Data
		str.Len = slice.Len

		arr{{ .method }} = make([]string, len(idx{{ .method }}))
		offset := 0
		for i, l := range idx{{ .method }} {
			strLen := int(l)
			arr{{ .method }}[i] = unpacked{{ .method }}[offset : offset+strLen]
			offset += strLen
		}
	})
{{range .indexes }}
	if typ == {{ $.enum }}({{.Type}}) {
		return arr{{ $.method }}[{{.From}}:{{.To}}]
	}
{{end}}
	return arr{{ .method }}
}
`

type NameIndex struct {
	Type, From, To int
}

func generate(input io.Reader, dataOutput io.Writer, indexOutput io.Writer) (int, []NameIndex, error) {
	dataOutput = gzip.NewWriter(dataOutput)
	defer dataOutput.(io.Closer).Close()

	names := make(map[int][]string)

	total := 0
	sc := bufio.NewScanner(input)
	for sc.Scan() {
		parts := strings.SplitN(sc.Text(), "\t", 2)
		if len(parts) != 2 {
			return 0, nil, fmt.Errorf("line %s invalid format", sc.Text())
		}

		nameType, err := strconv.Atoi(parts[0])
		if err != nil {
			return 0, nil, err
		}

		line := parts[1]
		if len(line) >= math.MaxUint8 {
			return 0, nil, fmt.Errorf("line %s to big", line)
		}

		total += len(line)
		names[nameType] = append(names[nameType], line)
	}

	var nameIndexes []NameIndex

	for nameType, n := range names {
		nameIdx := NameIndex{Type: nameType}
		if len(nameIndexes) != 0 {
			nameIdx.From = nameIndexes[len(nameIndexes)-1].To
		}
		nameIdx.To = nameIdx.From + len(n)
		nameIndexes = append(nameIndexes, nameIdx)

		sort.Strings(n)
		for _, s := range n {
			buf := []byte(s)
			if _, err := dataOutput.Write(buf); err != nil {
				return 0, nil, err
			}
			if _, err := indexOutput.Write([]byte{byte(len(buf))}); err != nil {
				return 0, nil, err
			}
		}
	}

	return total, nameIndexes, nil
}

func writeGoFile(goFile io.Writer, dataFileName string, idxFileName string, size int, indexes []NameIndex) error {
	templ := template.Must(template.New("src").Parse(srcCode))
	err := templ.Execute(goFile, map[string]interface{}{
		"package": *packageFlag,
		"method":  *methodFlag,
		"enum":    *enumFlag,
		"data":    dataFileName,
		"idx":     idxFileName,
		"size":    size,
		"indexes": indexes,
	})
	if err != nil {
		return err
	}
	return nil
}

func main() {
	flag.Parse()

	inputFile, err := os.Open(*inputFlag)
	if err != nil {
		log.Fatalf("open input file failed: %v", err)
	}
	defer inputFile.Close()

	outputDirectory := filepath.Dir(*outputFlag)
	baseOutputName := filepath.Base(*outputFlag)

	goFile, err := os.OpenFile(filepath.Join(outputDirectory, baseOutputName+".go"), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o777)
	if err != nil {
		log.Fatalf("create output file go file failed: %v", err)
	}
	defer goFile.Close()

	dataFileName := baseOutputName + ".txt.gz"
	dataFile, err := os.OpenFile(filepath.Join(outputDirectory, dataFileName), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o777)
	if err != nil {
		log.Fatalf("create output file go file failed: %v", err)
	}
	defer dataFile.Close()

	idxFileName := dataFileName + ".idx"

	idxFile, err := os.OpenFile(filepath.Join(outputDirectory, idxFileName), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o777)
	if err != nil {
		log.Fatalf("create output file go file failed: %v", err)
	}
	defer idxFile.Close()

	size, nameIndexes, err := generate(inputFile, dataFile, idxFile)
	if err != nil {
		log.Fatalf("generate failed: %v", err)
	}

	if err := writeGoFile(goFile, dataFileName, idxFileName, size, nameIndexes); err != nil {
		log.Fatalf("generate failed: %v", err)
	}
}
