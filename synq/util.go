package main

import (
	"encoding/json"
	"log"
	"os"
)

func writeSchema(schema json.RawMessage, filename string) {
	payload, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal schema: %v", err)
	}

	err = os.WriteFile(filename, payload, 0644)
	if err != nil {
		log.Fatalf("Failed to write schema to file: %v", err)
	}
}
