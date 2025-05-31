package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/nats-io/nats.go"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <command>")
		fmt.Println("Commands:")
		fmt.Println("  discover    - Discover all available services")
		fmt.Println("  info <name> - Get detailed information about a service")
		fmt.Println("  stats <name>- Get statistics for a service")
		fmt.Println("  ping        - Ping all services")
		os.Exit(1)
	}

	command := os.Args[1]

	// Connect to NATS
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer nc.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	switch command {
	case "discover", "ping":
		discoverServices(nc, ctx)
	case "info":
		if len(os.Args) < 3 {
			fmt.Println("Usage: go run main.go info <service-name>")
			os.Exit(1)
		}
		getServiceInfo(nc, ctx, os.Args[2])
	case "stats":
		if len(os.Args) < 3 {
			fmt.Println("Usage: go run main.go stats <service-name>")
			os.Exit(1)
		}
		getServiceStats(nc, ctx, os.Args[2])
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}

func discoverServices(nc *nats.Conn, ctx context.Context) {
	fmt.Println("🔍 Discovering NATS services...")

	// Send PING to discover services
	responses := make([]string, 0)

	// Use request-many to get responses from all services
	subject := "$SRV.PING"
	replies, err := nc.RequestWithContext(ctx, subject, nil)
	if err != nil {
		log.Printf("Error discovering services: %v", err)
		return
	}

	// Parse the response
	var pingResp struct {
		Name     string            `json:"name"`
		ID       string            `json:"id"`
		Version  string            `json:"version"`
		Metadata map[string]string `json:"metadata"`
		Type     string            `json:"type"`
	}

	if err := json.Unmarshal(replies.Data, &pingResp); err != nil {
		log.Printf("Error parsing response: %v", err)
		return
	}

	responses = append(responses, pingResp.Name)

	if len(responses) == 0 {
		fmt.Println("❌ No services found")
		return
	}

	fmt.Printf("✅ Found %d service(s):\n", len(responses))
	for i, serviceName := range responses {
		fmt.Printf("%d. %s (ID: %s, Version: %s)\n", i+1, serviceName, pingResp.ID, pingResp.Version)
	}

	fmt.Println("\n💡 To get more details about a service, run:")
	fmt.Printf("   go run main.go info %s\n", pingResp.Name)
	fmt.Printf("   go run main.go stats %s\n", pingResp.Name)
}

func getServiceInfo(nc *nats.Conn, ctx context.Context, serviceName string) {
	fmt.Printf("📋 Getting information for service: %s\n", serviceName)

	subject := fmt.Sprintf("$SRV.INFO.%s", serviceName)
	resp, err := nc.RequestWithContext(ctx, subject, nil)
	if err != nil {
		log.Printf("Error getting service info: %v", err)
		return
	}

	var info struct {
		Name        string `json:"name"`
		ID          string `json:"id"`
		Version     string `json:"version"`
		Description string `json:"description"`
		Endpoints   []struct {
			Name       string            `json:"name"`
			Subject    string            `json:"subject"`
			QueueGroup string            `json:"queue_group"`
			Metadata   map[string]string `json:"metadata"`
		} `json:"endpoints"`
		Metadata map[string]string `json:"metadata"`
		Type     string            `json:"type"`
	}

	if err := json.Unmarshal(resp.Data, &info); err != nil {
		log.Printf("Error parsing service info: %v", err)
		return
	}

	fmt.Printf("✅ Service Information:\n")
	fmt.Printf("   Name: %s\n", info.Name)
	fmt.Printf("   ID: %s\n", info.ID)
	fmt.Printf("   Version: %s\n", info.Version)
	fmt.Printf("   Description: %s\n", info.Description)

	if len(info.Endpoints) > 0 {
		fmt.Printf("   Endpoints:\n")
		for _, endpoint := range info.Endpoints {
			fmt.Printf("     • %s: %s (queue: %s)\n", endpoint.Name, endpoint.Subject, endpoint.QueueGroup)
			for key, value := range endpoint.Metadata {
				fmt.Printf("       %s: %s\n", key, value)
			}
		}
	}

	if len(info.Metadata) > 0 {
		fmt.Printf("   Metadata:\n")
		for key, value := range info.Metadata {
			fmt.Printf("     %s: %s\n", key, value)
		}
	}
}

func getServiceStats(nc *nats.Conn, ctx context.Context, serviceName string) {
	fmt.Printf("📊 Getting statistics for service: %s\n", serviceName)

	subject := fmt.Sprintf("$SRV.STATS.%s", serviceName)
	resp, err := nc.RequestWithContext(ctx, subject, nil)
	if err != nil {
		log.Printf("Error getting service stats: %v", err)
		return
	}

	var stats struct {
		Name      string    `json:"name"`
		ID        string    `json:"id"`
		Version   string    `json:"version"`
		Started   time.Time `json:"started"`
		Endpoints []struct {
			Name                  string        `json:"name"`
			Subject               string        `json:"subject"`
			QueueGroup            string        `json:"queue_group"`
			NumRequests           int64         `json:"num_requests"`
			NumErrors             int64         `json:"num_errors"`
			LastError             string        `json:"last_error"`
			ProcessingTime        time.Duration `json:"processing_time"`
			AverageProcessingTime time.Duration `json:"average_processing_time"`
		} `json:"endpoints"`
		Type string `json:"type"`
	}

	if err := json.Unmarshal(resp.Data, &stats); err != nil {
		log.Printf("Error parsing service stats: %v", err)
		return
	}

	fmt.Printf("✅ Service Statistics:\n")
	fmt.Printf("   Name: %s\n", stats.Name)
	fmt.Printf("   ID: %s\n", stats.ID)
	fmt.Printf("   Version: %s\n", stats.Version)
	fmt.Printf("   Started: %s\n", stats.Started.Format(time.RFC3339))
	fmt.Printf("   Uptime: %s\n", time.Since(stats.Started).Round(time.Second))

	if len(stats.Endpoints) > 0 {
		fmt.Printf("   Endpoint Statistics:\n")
		for _, endpoint := range stats.Endpoints {
			fmt.Printf("     • %s (%s):\n", endpoint.Name, endpoint.Subject)
			fmt.Printf("       Requests: %d\n", endpoint.NumRequests)
			fmt.Printf("       Errors: %d\n", endpoint.NumErrors)
			if endpoint.NumRequests > 0 {
				errorRate := float64(endpoint.NumErrors) / float64(endpoint.NumRequests) * 100
				fmt.Printf("       Error Rate: %.2f%%\n", errorRate)
				fmt.Printf("       Average Processing Time: %s\n", endpoint.AverageProcessingTime)
				fmt.Printf("       Total Processing Time: %s\n", endpoint.ProcessingTime)
			}
			if endpoint.LastError != "" {
				fmt.Printf("       Last Error: %s\n", endpoint.LastError)
			}
		}
	}
}
