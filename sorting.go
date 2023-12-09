package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"
	
	"github.com/gorilla/mux"
	"github.com/gorilla/handlers"
)

type Payload struct {
	ToSort [][]int `json:"to_sort"`
}

type Response struct {
	SortedArrays [][]int `json:"sorted_arrays"`
	TimeNS       int64   `json:"time_ns"`
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/process-single", processSingle).Methods("POST")
	router.HandleFunc("/process-concurrent", processConcurrent).Methods("POST")

	// Enable CORS for all routes
	corsHandler := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),           // Allow requests from any origin
		handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type"}),
		handlers.ExposedHeaders([]string{"Content-Type"}),
		handlers.MaxAge(86400),                            // Cache CORS preflight results for 1 day
		handlers.AllowCredentials(),
	)

	// Wrap the router with the CORS handler
	http.Handle("/", corsHandler(router))

	port := 8000
	serverAddr := fmt.Sprintf(":%d", port)

	// Start the server
	fmt.Printf("Server listening on port %d...\n", port)
	err := http.ListenAndServe(serverAddr, nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}

func processSingle(w http.ResponseWriter, r *http.Request) {

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	
	var payload Payload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	startTime := time.Now()

	// Sort each sub-array sequentially
	var results [][]int
	for _, subArray := range payload.ToSort {
		sortedArray := performSort("Sequential", subArray)
		results = append(results, sortedArray)
	}

	duration := time.Since(startTime)

	response := Response{
		SortedArrays: results,
		TimeNS:       duration.Nanoseconds(),
	}

	// Return JSON response
	writeJSONResponse(w, response)
}

func processConcurrent(w http.ResponseWriter, r *http.Request) {

if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	
		// Set Content-Type header for JSON
	w.Header().Set("Content-Type", "application/json")
	
	var payload Payload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	startTime := time.Now()

	var wg sync.WaitGroup
	resultCh := make(chan []int, len(payload.ToSort))

	for _, subArray := range payload.ToSort {
		wg.Add(1)
		go func(arr []int) {
			defer wg.Done()
			resultCh <- performSort("Concurrent", arr)
		}(subArray)
	}

	// Close the channel after all goroutines finish
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// Collect results
	var results [][]int
	for result := range resultCh {
		results = append(results, result)
	}

	duration := time.Since(startTime)

	response := Response{
		SortedArrays: results,
		TimeNS:       duration.Nanoseconds(),
	}

	// Return JSON response
	writeJSONResponse(w, response)
}

func performSort(taskName string, arr []int) []int {
	// Simulate sorting a sub-array
	sort.Ints(arr)
	time.Sleep(1 * time.Second)
	fmt.Printf("Task %s sorted: %v\n", taskName, arr)
	return arr
}

func writeJSONResponse(w http.ResponseWriter, response Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
