package main
import (
	"fmt"
	"math/rand"
	"github.com/casibase/go-hnsw"
)

func main() {

	const (
		M              = 32
		efConstruction = 400
		efSearch       = 100
		K              = 10
	)

	// var zero hnsw.Point = make([]float32, 128)

	h := hnsw.New(M, efConstruction)

	for i := 0; i < 1000; i++ {
		h.Add(randomPoint(), uint32(i))
	}

	query := randomPoint()

	fmt.Printf("Now searching with HNSW...\n")
	result := h.Search(query, efSearch, K)

	results := result.Items()
	for i := range results{
		println(results[i].ID)
	}



}

func randomPoint() hnsw.Point {
	var v hnsw.Point = make([]float32, 128)
	for i := range v {
		v[i] = rand.Float32()
	}
	return v
}
