package main

import (
	"fmt"
	"os"
	"time"

	"github.com/anishathalye/porcupine"
)

func saveVisualization(model porcupine.Model, info porcupine.LinearizationInfo, filename string) {
	f, _ := os.Create(filename)
	defer f.Close()
	porcupine.Visualize(model, info, f)
	fmt.Printf("- Visualization saved to %s\n", filename)
}

func main() {
	keys := []string{"x", "y", "z"}

	fmt.Println("--- Test 1: KVStore w/ locks + partition ---")

	_, info := runTest(TestConfig{
		NumClients:   5,
		OpsPerClient: 10,
		Keys:         keys,
		UsePartition: true,
	})

	saveVisualization(storeModel, info, "kvstore.html")
	fmt.Println()

	// ------------

	fmt.Println("--- Test 2: KVStore partition v/s naive ---")

	start := time.Now()
	_, info = runTest(TestConfig{
		NumClients:   20,
		OpsPerClient: 10,
		Keys:         keys,
		UsePartition: true,
	})
	partitionTime := time.Since(start)
	fmt.Printf("- Partition: %v\n", partitionTime)
	saveVisualization(storeModel, info, "partition.html")

	start = time.Now()
	_, info = runTest(TestConfig{
		NumClients:   20,
		OpsPerClient: 10,
		Keys:         keys,
		UsePartition: false,
	})
	naiveTime := time.Since(start)

	fmt.Printf("- Naive:     %v\n", naiveTime)
	saveVisualization(naiveModel, info, "naive.html")
}

// go build -o main
