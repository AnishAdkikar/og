package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	hnsw "github.com/AnishAdkikar/og/hnsw"
	"github.com/dustinxie/lockfree"
)

var (
	userHnswMap      = lockfree.NewHashMap()
	persistenceMutex sync.Mutex
	dataDir          = "user_data"
)

func getUserFilePath(userID string) string {
	return filepath.Join(dataDir, userID+".json")
}

func saveHNSWToFile(userID string, h *hnsw.Hnsw) error {
	persistenceMutex.Lock()
	defer persistenceMutex.Unlock()
	filePath := getUserFilePath(userID)

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Serialize and save HNSW to file
	encoder := json.NewEncoder(file)
	err = encoder.Encode(h)
	if err != nil {
		return err
	}

	return nil
}

func loadHNSWFromFile(userID string) (*hnsw.Hnsw, error) {
	persistenceMutex.Lock()
	defer persistenceMutex.Unlock()
	filePath := getUserFilePath(userID)

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Deserialize HNSW from file
	decoder := json.NewDecoder(file)
	var h hnsw.Hnsw
	err = decoder.Decode(&h)
	if err != nil {
		return nil, err
	}

	return &h, nil
}

func handleConnection(w http.ResponseWriter, r *http.Request) {
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
	h, ok := userHnswMap.Get(userID)
	if !ok {
		M_int, _ := strconv.Atoi(M)
		efConstruction_int, _ := strconv.Atoi(efConstruction)

		// Check if there is a saved HNSW for the user and load it
		h, err := loadHNSWFromFile(userID)
		if err != nil {
			// If not found, create a new HNSW
			h = hnsw.New(M_int, efConstruction_int)
			fmt.Println("Created new entry")
		}

		userHnswMap.Set(userID, h)
	} else {
		// If there's already an instance in the map, use it
		existingH, ok := h.(*hnsw.Hnsw)
		if !ok {
			http.Error(w, "Invalid HNSW instance in the map", http.StatusInternalServerError)
			return
		}
		h = existingH
		fmt.Println("loaded existing entry")
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Connection established successfully")
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
	fmt.Println(err)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error decoding JSON data: %s", err), http.StatusBadRequest)
		return
	}
	fmt.Println("Json data decoded")

	userID := requestData.UserID


	hInterface, ok := userHnswMap.Get(userID)
	if !ok {
		http.Error(w, "HNSW instance not found for the user", http.StatusBadRequest)
		return
	}

	h, ok := hInterface.(*hnsw.Hnsw)
	if !ok {
		http.Error(w, "HNSW instance not found for the user", http.StatusBadRequest)
		return
	}

	for text, vector := range requestData.Data {
		h.Add(vector, uint32(h.Size), text)
		fmt.Println(h.Size)
		h.Size++
	}
	err = saveHNSWToFile(userID, h)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error saving HNSW to file: %s", err), http.StatusInternalServerError)
		return
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
	hInterface, ok := userHnswMap.Get(userID)
	if !ok {
		http.Error(w, "HNSW instance not found for the user", http.StatusBadRequest)
		return
	}

	h, ok := hInterface.(*hnsw.Hnsw)
	if !ok {
		http.Error(w, "HNSW instance not found for the user", http.StatusBadRequest)
		return
	}

	res := h.Search(data, requestData.Ef, requestData.K)
	jsonResponse, err := json.Marshal(map[string][]string{"res": res})
	if err != nil {
		http.Error(w, fmt.Sprintf("Error encoding JSON response: %s", err), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}
func main() {
	err := os.MkdirAll(dataDir, os.ModePerm)
	if err != nil {
		fmt.Println("Error creating data directory:", err)
		return
	}

	http.HandleFunc("/connection", handleConnection)
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
