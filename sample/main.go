package main

import (
	"bufio"
	"encoding/json"
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
		M              = 32
		efConstruction = 400
		efSearch       = 100
		K              = 2
	)

	h := hnsw.New(M, efConstruction)

	if err := readEmbeddingsFile(outputEmbeddingsFile, h.Add); err != nil {
		log.Fatal(err)
	}
	fmt.Println("hnsw addition execution complete")

	cmd := exec.Command("python", "../scripts/search.py", "good ")
	output, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	var resultArray []float32
	if err := json.Unmarshal(output, &resultArray); err != nil {
		log.Fatal(err)
	}
	fmt.Println(resultArray)
	query := resultArray
	fmt.Printf("Now searching with HNSW...\n")
	result := h.Search(query, efSearch, K)
	fmt.Println(result)

	// Print data from the text file based on line numbers
	if err := printLinesByNumbers("../scripts/text_data.txt", result); err != nil {
		log.Fatal(err)
	}
	
}

func runPythonScript(pythonScript, inputCSVFile, outputEmbeddingsFile string) error {
	cmd := exec.Command("python", pythonScript, inputCSVFile, outputEmbeddingsFile)

	// Redirect standard output and standard error to Go's standard output and error
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command
	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func readEmbeddingsFile(filePath string, addFunc func(hnsw.Point, uint32, string)) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNumber := 1 // Start line numbering from 1
	for scanner.Scan() {
		line := scanner.Text()
		// Split the line into space-separated values
		values := strings.Fields(line)

		// Extract values and convert to appropriate types
		var data []float32
		for _, valStr := range values {
			val, err := strconv.ParseFloat(valStr, 32)
			if err != nil {
				return err
			}
			data = append(data, float32(val))
		}
		point := hnsw.Point(data)

		// Assuming the value is a string (adjust as needed)
		stringValue := "example" // Replace with the actual string value

		// Call the hnsw.Add function with the extracted values and line number
		addFunc(point, uint32(lineNumber), stringValue)

		// Increment line number for the next iteration
		lineNumber++
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}


func printLinesByNumbers(filePath string, lineNumbers []uint32) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	currentLineNumber := 1

	for scanner.Scan() {
		if contains(lineNumbers, currentLineNumber) {
			line := scanner.Text()
			fmt.Printf("%s\n", line)
		}

		currentLineNumber++
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func contains(arr []uint32, num int) bool {
	for _, v := range arr {
		if v == uint32(num) {
			return true
		}
	}
	return false
}