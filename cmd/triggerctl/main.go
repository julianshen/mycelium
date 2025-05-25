package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"mycelium/internal/trigger"

	"github.com/nats-io/nats.go"
)

func main() {
	// Parse command line flags
	natsURL := flag.String("nats-url", "nats://localhost:4222", "NATS server URL")
	streamName := flag.String("stream", "config-stream", "NATS stream name")
	flag.Parse()

	// Get subcommand
	args := flag.Args()
	if len(args) == 0 {
		fmt.Println("Usage: triggerctl <command> [options]")
		fmt.Println("\nCommands:")
		fmt.Println("  add <yaml-file>    Add a trigger from YAML file")
		fmt.Println("  list               List all triggers")
		fmt.Println("  delete <id>        Delete a trigger by ID")
		fmt.Println("  examples           Generate example trigger definitions")
		os.Exit(1)
	}

	// Connect to NATS
	nc, err := nats.Connect(*natsURL)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer nc.Close()

	// Create NATS store
	store, err := trigger.NewNATSStore(nc, *streamName)
	if err != nil {
		log.Fatalf("Failed to create trigger store: %v", err)
	}
	defer store.Close()

	// Load existing triggers
	ctx := context.Background()
	if err := store.LoadAll(ctx); err != nil {
		log.Fatalf("Failed to load triggers: %v", err)
	}

	// Handle commands
	switch args[0] {
	case "add":
		if len(args) != 2 {
			log.Fatal("Usage: triggerctl add <yaml-file>")
		}
		if err := addTrigger(ctx, store, args[1]); err != nil {
			log.Fatalf("Failed to add trigger: %v", err)
		}
		fmt.Println("Trigger added successfully")

	case "list":
		triggers := store.GetAllTriggers()
		if len(triggers) == 0 {
			fmt.Println("No triggers found")
			return
		}
		for _, t := range triggers {
			fmt.Printf("\nTrigger: %s\n", t.Name)
			fmt.Printf("  ID: %s\n", t.ID)
			fmt.Printf("  Namespaces: %v\n", t.Namespaces)
			fmt.Printf("  Event Type: %s\n", t.EventType)
			fmt.Printf("  Object Type: %s\n", t.ObjectType)
			fmt.Printf("  Criteria: %s\n", t.Criteria)
			fmt.Printf("  Action: %s\n", t.Action)
			fmt.Printf("  Enabled: %v\n", t.Enabled)
		}

	case "delete":
		if len(args) != 2 {
			log.Fatal("Usage: triggerctl delete <id>")
		}
		if err := store.DeleteTrigger(ctx, "default", args[1]); err != nil {
			log.Fatalf("Failed to delete trigger: %v", err)
		}
		fmt.Println("Trigger deleted successfully")

	case "examples":
		generateExamples()

	default:
		log.Fatalf("Unknown command: %s", args[0])
	}
}

func addTrigger(ctx context.Context, store *trigger.NATSStore, yamlFile string) error {
	// Read YAML file
	data, err := os.ReadFile(yamlFile)
	if err != nil {
		return fmt.Errorf("failed to read YAML file: %w", err)
	}

	// Parse trigger
	var t trigger.Trigger
	if err := t.FromYAML(data); err != nil {
		return fmt.Errorf("failed to parse trigger: %w", err)
	}

	// Save trigger
	return store.SaveTrigger(ctx, "default", t.ID, &t)
}

func generateExamples() {
	examples := []string{
		`# Example 1: Basic config update notification
id: config-update
name: Config Update Notification
namespaces: ["default"]
object_type: Config
event_type: config.updated
criteria: event.payload.after.critical == true
enabled: true
action: notify
description: Notifies when a critical config is updated`,

		`# Example 2: User role change detection
id: role-change
name: User Role Change Detection
namespaces: ["*"]
object_type: User
event_type: user.updated
criteria: event.payload.before.role != event.payload.after.role
enabled: true
action: audit
description: Detects when a user's role is changed`,

		`# Example 3: Resource usage alert
id: resource-alert
name: High Resource Usage Alert
namespaces: ["prod"]
object_type: Resource
event_type: resource.updated
criteria: event.payload.after.usage > 90
enabled: true
action: alert
description: Alerts when resource usage exceeds 90%`,

		`# Example 4: Complex condition with multiple fields
id: security-breach
name: Security Breach Detection
namespaces: ["*"]
object_type: Security
event_type: security.alert
criteria: |
  event.payload.after.severity == "high" &&
  event.payload.after.source_ip != "" &&
  has(event.payload.after, "attack_type")
enabled: true
action: security-response
description: Detects potential security breaches with high severity`,
	}

	for i, example := range examples {
		filename := fmt.Sprintf("trigger-example-%d.yaml", i+1)
		if err := os.WriteFile(filename, []byte(example), 0644); err != nil {
			log.Printf("Failed to write example %d: %v", i+1, err)
			continue
		}
		fmt.Printf("Generated %s\n", filename)
	}
}
