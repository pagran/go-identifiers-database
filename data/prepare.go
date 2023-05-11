package main

import (
	"bufio"
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
	inputFlag   = flag.String("i", "", "Path to input file")
	outputFlag  = flag.String("o", "", "Path to output file (without extension)")
	packageFlag = flag.String("p", "", "Package name")
	methodFlag  = flag.String("m", "", "Getter method name")
)

func generate(input io.Reader, dataOutput io.Writer, indexOutput io.Writer) error {
	sc := bufio.NewScanner(input)
	for sc.Scan() {
		line := sc.Bytes()
		if len(line) >= math.MaxUint8 {
			return fmt.Errorf("line %s to big", sc.Text())
		}

		if _, err := dataOutput.Write(line); err != nil {
			return err
		}

		if _, err := indexOutput.Write([]byte{byte(len(line))}); err != nil {
			return err
		}
	}

	return nil
}

func writeGoFile(goFile io.Writer, dataFileName string, idxFileName string) error {
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
	templ := template.Must(template.New("src").Parse(srcCode))
	templ.Execute(goFile, map[string]interface{}{
		"package": *packageFlag,
		"method":  *methodFlag,
		"data":    dataFileName,
		"idx":     idxFileName,
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

	if err := generate(inputFile, dataFile, idxFile); err != nil {
		log.Fatalf("generate failed: %v", err)
	}

	if err := writeGoFile(goFile, dataFileName, idxFileName); err != nil {
		log.Fatalf("generate failed: %v", err)
	}
}
