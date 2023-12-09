package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"
)

type Payload struct {
	ToSort [][]int `json:"to_sort"`
}

type Response struct {
	SortedArrays [][]int `json:"sorted_arrays"`
	TimeNS       int64   `json:"time_ns"`
}

func main() {
	// Set up the server
	http.HandleFunc("/process-single", processSingle)
	http.HandleFunc("/process-concurrent", processConcurrent)
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
