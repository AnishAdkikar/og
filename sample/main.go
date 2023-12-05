package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	hnsw "github.com/AnishAdkikar/og"
)

func main() {
	pythonScript := "../scripts/embeddings.py"
	inputCSVFile := "../scripts/data.csv"
	outputEmbeddingsFile := "../scripts/embeddings.txt"

	if err := runPythonScript(pythonScript, inputCSVFile, outputEmbeddingsFile); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Python file execution complete")
	const (
		M              = 3
		efConstruction = 4
		efSearch       = 2
		K              = 1
	)

	h := hnsw.New(M, efConstruction)
	if err := readEmbeddingsFiles(outputEmbeddingsFile, "../scripts/text_data.txt", h.Add); err != nil {
		log.Fatal(err)
	}
	fmt.Println("hnsw addition execution complete")

	cmd := exec.Command("python", "../scripts/search.py", "tiered bad working tired")
	output, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	var resultArray []float32
	if err := json.Unmarshal(output, &resultArray); err != nil {
		log.Fatal(err)
	}
	query := resultArray
	fmt.Printf("Now searching with HNSW...\n")
	result := h.Search(query, efSearch, K)
	fmt.Println(result)
}

func runPythonScript(pythonScript, inputCSVFile, outputEmbeddingsFile string) error {
	cmd := exec.Command("python", pythonScript, inputCSVFile, outputEmbeddingsFile)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}
func readEmbeddingsFiles(dataFilePath string, stringFilePath string, addFunc func(hnsw.Point, uint32, string)) error {
	dataFile, err := os.Open(dataFilePath)
	if err != nil {
		return err
	}
	defer dataFile.Close()

	stringFile, err := os.Open(stringFilePath)
	if err != nil {
		return err
	}
	defer stringFile.Close()

	scannerData := bufio.NewScanner(dataFile)
	scannerString := bufio.NewScanner(stringFile)

	lineNumber := 1 
	for scannerData.Scan() {
		dataLine := scannerData.Text()
		dataValues := strings.Fields(dataLine)
		var data []float32
		for _, valStr := range dataValues {
			val, err := strconv.ParseFloat(valStr, 32)
			if err != nil {
				return err
			}
			data = append(data, float32(val))
		}
		point := hnsw.Point(data)

		if scannerString.Scan() {
			stringValue := scannerString.Text()

			addFunc(point, uint32(lineNumber), stringValue)

			lineNumber++
		} else {
			return errors.New("second file has fewer lines than the first file")
		}
	}

	if err := scannerData.Err(); err != nil {
		return err
	}

	if err := scannerString.Err(); err != nil {
		return err
	}

	return nil
}


