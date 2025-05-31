package main

import (
	"fmt"
	"mycelium/internal/function"
)

func main() {
	fmt.Println("=== Memory Registry Example ===")

	// Create a new memory registry
	registry := &function.MemoryRegistry{}

	// Define some example function metadata
	functions := []function.FunctionMeta{
		{
			Name:    "echo-function",
			Type:    "builtin",
			Version: "1.0.0",
			Config: map[string]string{
				"description": "Echoes the input event data",
				"timeout":     "30s",
				"memory":      "128MB",
			},
		},
		{
			Name:    "transform-function",
			Type:    "builtin",
			Version: "2.1.0",
			Config: map[string]string{
				"description":   "Transforms data from one format to another",
				"input_format":  "json",
				"output_format": "xml",
				"timeout":       "60s",
			},
		},
		{
			Name:    "notify-function",
			Type:    "builtin",
			Version: "1.5.2",
			Config: map[string]string{
				"description": "Sends notifications via email or SMS",
				"smtp_host":   "smtp.example.com",
				"sms_api":     "https://api.sms.example.com",
			},
		},
	}

	// Store functions in the registry
	fmt.Println("\nStoring functions in registry:")
	for _, meta := range functions {
		// Mock binary data (in reality this would be actual function code)
		mockBinary := []byte(fmt.Sprintf("binary-data-for-%s", meta.Name))

		err := registry.StoreFunction(meta, mockBinary)
		if err != nil {
			fmt.Printf("Error storing function %s: %v\n", meta.Name, err)
			continue
		}
		fmt.Printf("✓ Stored function: %s v%s\n", meta.Name, meta.Version)
	}

	// List all functions
	fmt.Println("\nListing all functions:")
	allFunctions, err := registry.ListFunctions()
	if err != nil {
		fmt.Printf("Error listing functions: %v\n", err)
		return
	}

	for i, meta := range allFunctions {
		fmt.Printf("%d. %s v%s (%s)\n", i+1, meta.Name, meta.Version, meta.Type)
		if desc, ok := meta.Config["description"]; ok {
			fmt.Printf("   Description: %s\n", desc)
		}
		if len(meta.Config) > 1 { // More than just description
			fmt.Printf("   Config: %+v\n", meta.Config)
		}
		fmt.Println()
	}

	// Retrieve a specific function
	fmt.Println("Retrieving specific function:")
	functionName := "transform-function"
	meta, binary, err := registry.GetFunction(functionName)
	if err != nil {
		fmt.Printf("Error retrieving function %s: %v\n", functionName, err)
		return
	}

	fmt.Printf("Retrieved function: %s\n", meta.Name)
	fmt.Printf("  Version: %s\n", meta.Version)
	fmt.Printf("  Type: %s\n", meta.Type)
	if desc, ok := meta.Config["description"]; ok {
		fmt.Printf("  Description: %s\n", desc)
	}
	fmt.Printf("  Binary size: %d bytes\n", len(binary))
	fmt.Printf("  Config: %+v\n", meta.Config)

	// Try to retrieve a non-existent function
	fmt.Println("\nTrying to retrieve non-existent function:")
	_, _, err = registry.GetFunction("non-existent-function")
	if err != nil {
		fmt.Printf("Expected error: %v\n", err)
	}

	// Delete a function
	fmt.Println("\nDeleting a function:")
	deleteFunction := "echo-function"
	err = registry.DeleteFunction(deleteFunction)
	if err != nil {
		fmt.Printf("Error deleting function %s: %v\n", deleteFunction, err)
	} else {
		fmt.Printf("✓ Deleted function: %s\n", deleteFunction)
	}

	// List functions after deletion
	fmt.Println("\nFunctions after deletion:")
	remainingFunctions, err := registry.ListFunctions()
	if err != nil {
		fmt.Printf("Error listing functions: %v\n", err)
		return
	}

	fmt.Printf("Remaining functions: %d\n", len(remainingFunctions))
	for _, meta := range remainingFunctions {
		fmt.Printf("  - %s v%s\n", meta.Name, meta.Version)
	}

	// Demonstrate updating a function (store with same name)
	fmt.Println("\nUpdating a function:")
	updatedMeta := function.FunctionMeta{
		Name:    "transform-function",
		Type:    "builtin",
		Version: "2.2.0", // New version
		Config: map[string]string{
			"description":   "Enhanced data transformation function with XML support",
			"input_format":  "json",
			"output_format": "xml",
			"timeout":       "90s", // Increased timeout
			"xml_schema":    "https://schemas.example.com/v2",
		},
	}

	err = registry.StoreFunction(updatedMeta, []byte("updated-binary-data"))
	if err != nil {
		fmt.Printf("Error updating function: %v\n", err)
	} else {
		fmt.Printf("✓ Updated function: %s to v%s\n", updatedMeta.Name, updatedMeta.Version)
	}

	// Verify the update
	meta, _, err = registry.GetFunction("transform-function")
	if err != nil {
		fmt.Printf("Error retrieving updated function: %v\n", err)
	} else {
		fmt.Printf("Verified update: %s v%s\n", meta.Name, meta.Version)
		fmt.Printf("New timeout: %s\n", meta.Config["timeout"])
	}
}
