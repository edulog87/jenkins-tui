package jenkins

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/elogrono/jenkins-tui/internal/config"
	"github.com/elogrono/jenkins-tui/internal/jenkins/models"
)

func TestGetPipelineRunFromWFAPI(t *testing.T) {
	// Mock response matching the structure from your example
	mockResponse := models.WFAPIRun{
		ID:                  "221",
		Name:                "#221",
		Status:              "SUCCESS",
		StartTimeMillis:     1767194616671,
		EndTimeMillis:       1767194914425,
		DurationMillis:      297754,
		QueueDurationMillis: 5,
		PauseDurationMillis: 0,
		Stages: []models.WFAPIStage{
			{
				ID:                  "6",
				Name:                "Declarative: Checkout SCM",
				Status:              "SUCCESS",
				StartTimeMillis:     1767194617586,
				DurationMillis:      2448,
				PauseDurationMillis: 0,
			},
			{
				ID:                  "17",
				Name:                "A. Show setup",
				Status:              "SUCCESS",
				StartTimeMillis:     1767194620133,
				DurationMillis:      2022,
				PauseDurationMillis: 0,
			},
			{
				ID:                  "29",
				Name:                "B. Checkout code",
				Status:              "SUCCESS",
				StartTimeMillis:     1767194622168,
				DurationMillis:      2181,
				PauseDurationMillis: 0,
			},
		},
	}

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify correct endpoint
		if r.URL.Path != "/job/test-job/221/wfapi/describe" {
			t.Errorf("Expected path /job/test-job/221/wfapi/describe, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	// Create client with proper configuration
	cfg := &config.Config{
		Profile: config.Profile{
			BaseURL:      server.URL,
			Username:     "test",
			APIToken:     "test-token",
			RateLimitRPS: 10,
		},
	}
	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test GetPipelineRun (should use wfapi)
	run, err := client.GetPipelineRun(context.Background(), "test-job", 221)

	if err != nil {
		t.Fatalf("GetPipelineRun failed: %v", err)
	}

	if run == nil {
		t.Fatal("Expected non-nil run")
	}

	// Verify fields
	if run.ID != "221" {
		t.Errorf("Expected ID 221, got %s", run.ID)
	}

	if run.Status != "SUCCESS" {
		t.Errorf("Expected status SUCCESS, got %s", run.Status)
	}

	if len(run.Stages) != 3 {
		t.Errorf("Expected 3 stages, got %d", len(run.Stages))
	}

	// Verify first stage
	if run.Stages[0].Name != "Declarative: Checkout SCM" {
		t.Errorf("Expected first stage name 'Declarative: Checkout SCM', got %s", run.Stages[0].Name)
	}

	if run.Stages[0].Status != "SUCCESS" {
		t.Errorf("Expected first stage status SUCCESS, got %s", run.Stages[0].Status)
	}
}

func TestGetStageLog(t *testing.T) {
	mockLog := "Step 1: Checkout\nStep 2: Build\nStep 3: Test\n"

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify wfapi endpoint is used first
		expectedPath := "/job/test-job/221/execution/node/6/wfapi/log"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(mockLog))
	}))
	defer server.Close()

	// Create client with proper configuration
	cfg := &config.Config{
		Profile: config.Profile{
			BaseURL:      server.URL,
			Username:     "test",
			APIToken:     "test-token",
			RateLimitRPS: 10,
		},
	}
	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test GetStageLog
	log, err := client.GetStageLog(context.Background(), "test-job", 221, "6")
	if err != nil {
		t.Fatalf("GetStageLog failed: %v", err)
	}

	if log != mockLog {
		t.Errorf("Expected log %q, got %q", mockLog, log)
	}
}

func TestGetJob_WithStages(t *testing.T) {
	mockJob := models.JobDetail{
		Name:  "test-job",
		Color: "blue",
		Builds: []models.BuildRef{
			{Number: 1, Result: "SUCCESS"},
			{Number: 2, Result: "FAILURE"},
		},
	}

	mockWFRun := models.WFAPIRun{
		ID:     "1",
		Status: "SUCCESS",
		Stages: []models.WFAPIStage{
			{ID: "1", Name: "Stage 1", Status: "SUCCESS"},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Return job detail
		if r.URL.Path == "/job/test-job/api/json" {
			json.NewEncoder(w).Encode(mockJob)
			return
		}

		// Return wfapi run for build #1
		if r.URL.Path == "/job/test-job/1/wfapi/describe" {
			json.NewEncoder(w).Encode(mockWFRun)
			return
		}

		// Return 404 for build #2 (simulate no pipeline)
		if r.URL.Path == "/job/test-job/2/wfapi/describe" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}))
	defer server.Close()

	cfg := &config.Config{
		Profile: config.Profile{
			BaseURL:      server.URL,
			Username:     "test",
			APIToken:     "test-token",
			RateLimitRPS: 10,
		},
	}
	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	job, err := client.GetJob(context.Background(), "test-job")
	if err != nil {
		t.Fatalf("GetJob failed: %v", err)
	}

	if job == nil {
		t.Fatal("Expected non-nil job")
	}

	// Verify that build #1 has stages
	if len(job.Builds) < 1 {
		t.Fatal("Expected at least 1 build")
	}

	if len(job.Builds[0].Stages) != 1 {
		t.Errorf("Expected build #1 to have 1 stage, got %d", len(job.Builds[0].Stages))
	}
}
