package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

type RegisterServiceDTO struct {
	Name          string `json:"name"`
	URL           string `json:"url"`
	CheckInterval int    `json:"check_interval"`
}

func main() {
	apiURL := flag.String("api", "http://localhost:8080/api/v1", "Base URL of the Health Checker API")
	numServices := flag.Int("services", 1000, "Number of services to register")
	concurrency := flag.Int("concurrency", 20, "Number of concurrent workers")
	mockPort := flag.Int("mock-port", 9090, "Port for the mock service to listen on")
	durationFlag := flag.Int("duration", 30, "Duration in seconds to wait and observe health checks")
	flag.Parse()

	// 1. Start Mock Server
	go startMockServer(*mockPort)
	fmt.Printf("Started mock target server on port %d\n", *mockPort)

	// 2. Register Services
	fmt.Printf("Starting load test: Registering %d services with %d workers...\n", *numServices, *concurrency)
	start := time.Now()

	var wg sync.WaitGroup
	jobs := make(chan int, *numServices)
	var successCount int64
	var failCount int64

	// Start workers
	for i := 0; i < *concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for id := range jobs {
				service := RegisterServiceDTO{
					Name:          fmt.Sprintf("load-test-service-%d", id),
					URL:           fmt.Sprintf("http://host.docker.internal:%d/health/%d", *mockPort, id), // Use host.docker.internal if running app in docker
					CheckInterval: 10 + (id % 50),                                                         // Spread out intervals
				}

				// If running locally (not docker), use localhost
				// But for now let's assume localhost for the URL registered,
				// assuming the worker running inside docker can reach the host or
				// if everything is local.
				// Let's stick to localhost for the URL field if we assume the app is running locally too.
				// If the app is in docker, it needs to reach this loadtest script.
				// Let's use a flag or just localhost and assume network mode host or similar.
				// Actually, simpler: just use http://localhost:... and assume the user handles networking.
				service.URL = fmt.Sprintf("http://localhost:%d/health/%d", *mockPort, id)

				if err := registerService(*apiURL, service); err != nil {
					// log.Printf("Failed to register service %d: %v", id, err)
					atomic.AddInt64(&failCount, 1)
				} else {
					atomic.AddInt64(&successCount, 1)
				}
			}
		}()
	}

	// Enqueue jobs
	for i := 0; i < *numServices; i++ {
		jobs <- i
	}
	close(jobs)

	wg.Wait()
	duration := time.Since(start)

	fmt.Printf("\nLoad Test Completed in %v\n", duration)
	fmt.Printf("Successful Registrations: %d\n", successCount)
	fmt.Printf("Failed Registrations:     %d\n", failCount)
	fmt.Printf("Throughput:               %.2f req/sec\n", float64(*numServices)/duration.Seconds())

	if failCount > 0 {
		fmt.Printf("WARNING: %d requests failed. Check server logs.\n", failCount)
	} else {
		fmt.Println("SUCCESS: All services registered.")
	}

	// 3. Wait and observe (Optional)
	fmt.Printf("\nWaiting for %d seconds to receive health checks from the system...\n", *durationFlag)
	time.Sleep(time.Duration(*durationFlag) * time.Second)
	fmt.Println("Load test finished.")
}

func startMockServer(port int) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Simulate some latency
		// time.Sleep(10 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

func registerService(apiBase string, dto RegisterServiceDTO) error {
	jsonData, err := json.Marshal(dto)
	if err != nil {
		return err
	}

	resp, err := http.Post(fmt.Sprintf("%s/services", apiBase), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
