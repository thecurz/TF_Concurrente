// client/main.go
package main

import (
	"encoding/json"
	"log"
	"net"
	"recomendador/config"
	"recomendador/utils"
	"runtime"
)

func main() {
	log.Println("Client is starting...")

	// Load client configuration
	cfg := config.LoadClientConfig()
	log.Printf("Loaded client configuration: %+v\n", cfg)

	// Connect to the server
	conn, err := net.Dial("tcp", cfg.Server.Address)
	if err != nil {
		log.Fatalf("Error connecting to server: %v", err)
	}
	defer conn.Close()
	log.Println("Connected to server.")

	// Receive data from server
	decoder := json.NewDecoder(conn)
	var serverMsg utils.ServerMessage
	err = decoder.Decode(&serverMsg)
	if err != nil {
		log.Fatalf("Error decoding server message: %v", err)
	}
	log.Printf("Received message from server: %+v\n", serverMsg)

	if serverMsg.Message == "NO_MORE_WORK" {
		log.Println("No more work assigned by server")
		return
	}

	partition := serverMsg.Partition
	log.Printf("Received partition with %d reviews.\n", len(partition))

	// Perform computation
	log.Println("Starting computation...")
	results := utils.PerformComputation(partition)
	log.Println("Computation completed.")

	// Log memory usage
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	log.Printf("Memory Usage: Alloc = %v MiB", m.Alloc/1024/1024)

	// Send results back to server
	encoder := json.NewEncoder(conn)
	err = encoder.Encode(results)
	if err != nil {
		log.Fatalf("Error encoding results: %v", err)
	}
	log.Println("Results sent to server.")
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.CloseWrite()
	}
	conn.Close()
	log.Println("Connection closed.")
}
