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
	"text/template"
)

var (
	inputFlag    = flag.String("i", "", "Path to input file")
	outputFlag   = flag.String("o", "", "Path to output file (without extension)")
	packageFlag  = flag.String("p", "", "Package name")
	methodFlag   = flag.String("m", "", "Getter method name")
	compressFlag = flag.Bool("c", false, "Add data file compression")
)

const srcCode = `package {{ .package }}

import (
	_ "embed"
	"sync"
)

var (
	//go:embed {{ .data }}
	data{{ .method }} string
	//go:embed {{ .idx }}
	idx{{ .method }} []byte

	once{{ .method }} sync.Once
	arr{{ .method }}  []string
)

func {{ .method }}() []string {
	once{{ .method }}.Do(func() {
		arr{{ .method }} = make([]string, len(idx{{ .method }}))
		offset := 0
		for i, l := range idx{{ .method }} {
			strLen := int(l)
			arr{{ .method }}[i] = data{{ .method }}[offset : offset+strLen]
			offset += strLen
		}
	})
	return arr{{ .method }}
}
`

const srcCodeWithCompression = `package {{ .package }}

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

func {{ .method }}() []string {
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
	return arr{{ .method }}
}
`

func generate(input io.Reader, dataOutput io.Writer, indexOutput io.Writer) (int, error) {
	if *compressFlag {
		dataOutput = gzip.NewWriter(dataOutput)
		defer dataOutput.(io.Closer).Close()
	}
	total := 0
	sc := bufio.NewScanner(input)
	for sc.Scan() {
		line := sc.Bytes()
		if len(line) >= math.MaxUint8 {
			return 0, fmt.Errorf("line %s to big", sc.Text())
		}

		total += len(line)
		if _, err := dataOutput.Write(line); err != nil {
			return 0, err
		}

		if _, err := indexOutput.Write([]byte{byte(len(line))}); err != nil {
			return 0, err
		}
	}

	return total, nil
}

func writeGoFile(goFile io.Writer, dataFileName string, idxFileName string, size int) error {
	src := srcCode
	if *compressFlag {
		src = srcCodeWithCompression
	}
	templ := template.Must(template.New("src").Parse(src))
	templ.Execute(goFile, map[string]interface{}{
		"package": *packageFlag,
		"method":  *methodFlag,
		"data":    dataFileName,
		"idx":     idxFileName,
		"size":    size,
	})
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

	dataFileName := baseOutputName + ".txt"
	if *compressFlag {
		dataFileName += ".gz"
	}

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

	size, err := generate(inputFile, dataFile, idxFile)
	if err != nil {
		log.Fatalf("generate failed: %v", err)
	}

	if err := writeGoFile(goFile, dataFileName, idxFileName, size); err != nil {
		log.Fatalf("generate failed: %v", err)
	}
}
