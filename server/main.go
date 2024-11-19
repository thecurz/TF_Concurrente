package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"recomendador/config"
	"recomendador/utils"

	"github.com/gorilla/websocket"
)

var (
	partitions        [][]utils.Review    // Data partitions
	partitionIndex    int                 // Index for next partition
	partitionMutex    sync.Mutex          // Mutex for partition index
	aggregatedResults []utils.ResultData  // Collected client results
	resultsMutex      sync.Mutex          // Mutex for aggregated results
	recommendations   map[string][]string // Map of categories to recommendations
	recommendationsMu sync.RWMutex        // Mutex for recommendations map
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	// Load server configuration
	cfg := config.LoadServerConfig()

	// Load and partition dataset
	data := utils.LoadData(cfg.Dataset.Path)
	partitions = utils.SplitData(data, cfg.Dataset.Partitions)
	partitionIndex = 0

	var wg sync.WaitGroup

	// Start the TCP server for client communication
	wg.Add(1)
	go func() {
		defer wg.Done()
		startTCPServer(cfg)
	}()

	// Wait for all clients to finish processing and recommendations to be ready
	wg.Wait()
	processAggregatedResults()

	// Start the CLI for user interaction
	go startCLI()

	// Start the WebSocket server for UI interaction
	startWebSocketServer()
}

func startWebSocketServer() {
	http.HandleFunc("/ws", handleWebSocket)
	fmt.Println("WebSocket Server running on ws://localhost:8080/ws")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v\n", err)
		return
	}
	defer conn.Close()

	for {
		var req struct {
			Category string `json:"category"`
		}
		err := conn.ReadJSON(&req)
		if err != nil {
			log.Printf("Error reading JSON: %v\n", err)
			break
		}

		recommendationsMu.RLock()
		recs, exists := recommendations[req.Category]
		recommendationsMu.RUnlock()

		if !exists {
			recs = []string{}
		}

		resp := struct {
			Recommendations []string `json:"recommendations"`
		}{Recommendations: recs}
		err = conn.WriteJSON(resp)
		if err != nil {
			log.Printf("Error writing JSON: %v\n", err)
			break
		}
	}
}

func startTCPServer(cfg config.ServerConfig) {
	ln, err := net.Listen("tcp", ":"+cfg.Server.Port)
	ln.(*net.TCPListener).SetDeadline(
		time.Now().Add(time.Minute * 1),
	) // TODO: change timeout to 10min?
	if err != nil {
		fmt.Println("Error starting TCP server:", err)
		return
	}
	defer ln.Close()
	fmt.Println("TCP Server listening on port", cfg.Server.Port)

	var wg sync.WaitGroup

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		wg.Add(1)
		go handleClient(conn, &wg)

		// Exit condition: all partitions have been assigned
		partitionMutex.Lock()
		done := partitionIndex >= len(partitions)
		partitionMutex.Unlock()
		if done {
			break
		}
	}

	// Wait for all clients to finish
	clientWg.Wait()
	wg.Wait()
}

func startCLI() {
	fmt.Println(
		"Recommendations are ready. You can now enter product categories to get recommendations.",
	)
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Enter product category (type 'categories' to list, 'exit' to quit): ")
		if !scanner.Scan() {
			break
		}
		category := scanner.Text()
		category = strings.ToLower(strings.TrimSpace(category))

		if category == "exit" {
			fmt.Println("Exiting CLI.")
			break
		} else if category == "categories" {
			displayAvailableCategories()
			continue
		}

		displayRecommendations(category)
	}
}

func displayAvailableCategories() {
	recommendationsMu.RLock()
	defer recommendationsMu.RUnlock()
	if len(recommendations) == 0 {
		fmt.Println("No categories available.")
		return
	}

	fmt.Println("Available categories:")
	for category := range recommendations {
		fmt.Printf("- %s\n", category)
	}
}

func displayRecommendations(category string) {
	recommendationsMu.RLock()
	defer recommendationsMu.RUnlock()
	if recommendations == nil {
		fmt.Println("Recommendations are not ready yet. Please try again later.")
		return
	}

	recs, exists := recommendations[category]
	if !exists || len(recs) == 0 {
		fmt.Printf("No recommendations found for the category '%s'.\n", category)
		return
	}

	// Limit to top 10 recommendations
	limit := 10
	if len(recs) < limit {
		limit = len(recs)
	}
	topRecs := recs[:limit]

	// Display recommendations
	fmt.Printf("Top %d recommendations for category '%s':\n", limit, category)
	for i, productID := range topRecs {
		fmt.Printf("%d. %s\n", i+1, productID)
	}
}

var clientWg sync.WaitGroup

func handleClient(conn net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()
	defer conn.Close()

	clientWg.Add(1)
	defer clientWg.Done()

	clientAddr := conn.RemoteAddr().String()
	fmt.Printf("Client connected: %s\n", clientAddr)

	partition := getNextPartition()
	encoder := json.NewEncoder(conn)

	if partition == nil {
		// No more partitions to assign
		fmt.Println("No more partitions to assign to client")
		noWorkMessage := utils.ServerMessage{Message: "NO_MORE_WORK"}
		err := encoder.Encode(noWorkMessage)
		if err != nil {
			fmt.Printf("Error sending NO_MORE_WORK message to client %s: %v\n", clientAddr, err)
		}
		return
	}

	// Send partition data to client
	serverMessage := utils.ServerMessage{Partition: partition}
	err := encoder.Encode(serverMessage)
	if err != nil {
		fmt.Printf("Error sending data to client %s: %v\n", clientAddr, err)
		return
	}
	fmt.Printf("Sent partition data to client %s\n", clientAddr)

	// Receive results from client
	decoder := json.NewDecoder(conn)
	var results utils.ResultData
	err = decoder.Decode(&results)
	if err != nil {
		if err == io.EOF {
			fmt.Printf("Client %s closed the connection.\n", clientAddr)
		} else {
			fmt.Printf("Error decoding client results from %s: %v\n", clientAddr, err)
		}
		return
	}
	fmt.Printf("Received results from client %s\n", clientAddr)

	// Aggregate results
	aggregateResults(results)
}

func getNextPartition() []utils.Review {
	partitionMutex.Lock()
	defer partitionMutex.Unlock()

	if partitionIndex >= len(partitions) {
		return nil
	}

	partition := partitions[partitionIndex]
	partitionIndex++
	return partition
}

func aggregateResults(results utils.ResultData) {
	resultsMutex.Lock()
	defer resultsMutex.Unlock()

	aggregatedResults = append(aggregatedResults, results)
}

func processAggregatedResults() {
	fmt.Println("Processing aggregated results...")

	combinedRecommendations := make(map[string]map[string]int) // Map[category]Map[productID]score

	for _, result := range aggregatedResults {
		for _, recs := range result.Recommendations {
			for _, productID := range recs {
				category := utils.GetProductCategory(productID)
				if category == "" {
					continue
				}
				if _, exists := combinedRecommendations[category]; !exists {
					combinedRecommendations[category] = make(map[string]int)
				}
				combinedRecommendations[category][productID] += 1 // Increase score
			}
		}
	}

	recommendationsMu.Lock()
	recommendations = make(map[string][]string)
	for category, productsMap := range combinedRecommendations {
		// Create a slice of product-score pairs
		type productScore struct {
			ProductID string
			Score     int
		}
		var productScores []productScore
		for productID, score := range productsMap {
			productScores = append(productScores, productScore{ProductID: productID, Score: score})
		}
		// Sort products by score descending
		sort.Slice(productScores, func(i, j int) bool {
			return productScores[i].Score > productScores[j].Score
		})
		// Extract sorted product IDs
		products := make([]string, len(productScores))
		for i, ps := range productScores {
			products[i] = ps.ProductID
		}
		recommendations[category] = products
	}
	recommendationsMu.Unlock()

	fmt.Println("Recommendations processing completed.")

	recommendationsMu.RLock()
	defer recommendationsMu.RUnlock()
	fmt.Println("Generated Recommendations:")
	if len(recommendations) == 0 {
		fmt.Println("No recommendations were generated.")
	} else {
		for category, recs := range recommendations {
			fmt.Printf("Category: '%s', Recommendations: %v\n", category, recs)
		}
	}
}
