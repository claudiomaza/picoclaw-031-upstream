package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	a2a "github.com/sipeed/picoclaw/pkg/extensibility/a2a"
)

func main() {
	var req a2a.TurnRequest
	if os.Getenv("PICOCLAW_PROFILE_HTTP_ADDR") == "" {
		if err := json.NewDecoder(os.Stdin).Decode(&req); err != nil {
			emitError(err)
			return
		}
	}
	path := os.Getenv("PICOCLAW_CONFIG")
	if path == "" {
		path = os.ExpandEnv("$HOME/.picoclaw/config.json")
	}
	runner, err := a2a.NewRunner(path)
	if err != nil {
		emitError(err)
		return
	}
	if req.Operation == "list_agents" {
		_ = json.NewEncoder(os.Stdout).Encode(map[string]any{"status": "COMPLETED", "agents": runner.ListAgents()})
		return
	}
	if addr := os.Getenv("PICOCLAW_PROFILE_HTTP_ADDR"); addr != "" {
		log.Printf("profile API listening on %s", addr)
		if err := http.ListenAndServe(addr, runner.ProfileHandler()); err != nil {
			emitError(err)
		}
		return
	}
	result, err := runner.Execute(context.Background(), req)
	if err != nil {
		emitError(err)
		return
	}
	_ = json.NewEncoder(os.Stdout).Encode(result)
}

func emitError(err error) {
	if err == io.EOF {
		err = fmt.Errorf("request JSON is required")
	}
	_ = json.NewEncoder(os.Stdout).Encode(map[string]any{"error": err.Error()})
	os.Exit(1)
}
