package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	hnsw "github.com/AnishAdkikar/og"
)

var (
	userHnswMap = make(map[string]*hnsw.Hnsw)
	mu          sync.Mutex
)

func handleNewConnection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed", http.StatusMethodNotAllowed)
		return
	}
	var requestData struct {
		UserID         string `json:"userID"`
		M              string `json:"M"`
		EfConstruction string `json:"efConstruction"`
	}

	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error decoding JSON data: %s", err), http.StatusBadRequest)
		return
	}

	userID := requestData.UserID
	M := requestData.M
	efConstruction := requestData.EfConstruction
	// fmt.Println(userID, M, efConstruction)
	mu.Lock()
	defer mu.Unlock()

	h, ok := userHnswMap[userID]
	if !ok {
		fmt.Println("Creating new entry")
		M_int, _ := strconv.Atoi(M)
		efConstruction_int, _ := strconv.Atoi(efConstruction)
		h = hnsw.New(M_int, efConstruction_int)
		userHnswMap[userID] = h
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Data received successfully")
}

func handleAddData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed", http.StatusMethodNotAllowed)
		return
	}

	var requestData struct {
		UserID string               `json:"userID"`
		Data   map[string][]float32 `json:"data"`
	}

	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error decoding JSON data: %s", err), http.StatusBadRequest)
		return
	}

	userID := requestData.UserID

	mu.Lock()
	defer mu.Unlock()

	h, ok := userHnswMap[userID]
	if !ok {
		http.Error(w, "HNSW instance not found for the user", http.StatusBadRequest)
		return
	}

	for text, vector := range requestData.Data {
		h.Add(vector, uint32(h.Size), text)
		fmt.Println(text,vector,h.Size)
		h.Size++
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Data added successfully")
}

func handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed", http.StatusMethodNotAllowed)
		return
	}

	var requestData struct {
		UserID string    `json:"userID"`
		Ef     int       `json:"ef"`
		K      int       `json:"K"`
		Data   []float32 `json:"data"`
	}

	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error decoding JSON data: %s", err), http.StatusBadRequest)
		return
	}

	userID := requestData.UserID
	data := requestData.Data

	mu.Lock()
	defer mu.Unlock()

	h, ok := userHnswMap[userID]
	if !ok {
		http.Error(w, "HNSW instance not found for the user", http.StatusBadRequest)
		return
	}

	res := h.Search(data, requestData.Ef, requestData.K)

	w.WriteHeader(http.StatusOK)
	// fmt.Fprint(w, "Data added successfully")
	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error encoding JSON response: %s", err), http.StatusInternalServerError)
		return
	}
}
func main() {

	http.HandleFunc("/connection", handleNewConnection)
	http.HandleFunc("/add-data", handleAddData)
	http.HandleFunc("/search", handleSearch)
	go func() {
		if err := http.ListenAndServe(":8080", nil); err != nil {
			fmt.Println(err)
		}
	}()
	fmt.Println("Server listening on http://127.0.0.1:8080")
	select {}

}
