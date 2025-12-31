package jenkins

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/elogrono/jenkins-tui/internal/config"
)

func testConfig(url string) *config.Config {
	return &config.Config{
		Profile: config.Profile{
			BaseURL:        url,
			Username:       "testuser",
			APIToken:       "testtoken",
			TimeoutSeconds: 5,
			RateLimitRPS:   10,
		},
	}
}

func TestNewClient(t *testing.T) {
	cfg := testConfig("https://jenkins.example.com")
	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	if client.baseURL != "https://jenkins.example.com" {
		t.Errorf("expected baseURL https://jenkins.example.com, got %s", client.baseURL)
	}
	if client.username != "testuser" {
		t.Errorf("expected username testuser, got %s", client.username)
	}
}

func TestNewClientTrimsTrailingSlash(t *testing.T) {
	cfg := testConfig("https://jenkins.example.com/")
	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	if client.baseURL != "https://jenkins.example.com" {
		t.Errorf("expected baseURL without trailing slash, got %s", client.baseURL)
	}
}

func TestNewClientNilConfig(t *testing.T) {
	_, err := NewClient(nil)
	if err == nil {
		t.Error("expected error for nil config")
	}
}

func TestGetRootInfo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check auth header
		user, pass, ok := r.BasicAuth()
		if !ok || user != "testuser" || pass != "testtoken" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Check path
		if r.URL.Path != "/api/json" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		response := map[string]interface{}{
			"mode":            "NORMAL",
			"nodeDescription": "the master Jenkins node",
			"nodeName":        "",
			"numExecutors":    2,
			"useCrumbs":       true,
			"useSecurity":     true,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	cfg := testConfig(server.URL)
	client, _ := NewClient(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	info, err := client.GetRootInfo(ctx)
	if err != nil {
		t.Fatalf("GetRootInfo failed: %v", err)
	}

	if info.Mode != "NORMAL" {
		t.Errorf("expected mode NORMAL, got %s", info.Mode)
	}
	if info.NumExecutors != 2 {
		t.Errorf("expected 2 executors, got %d", info.NumExecutors)
	}
	if !info.UseCrumbs {
		t.Error("expected useCrumbs to be true")
	}
}

func TestGetViews(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, _, ok := r.BasicAuth()
		if !ok || user != "testuser" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		response := map[string]interface{}{
			"views": []map[string]string{
				{"name": "All", "url": "https://jenkins.example.com/"},
				{"name": "Production", "url": "https://jenkins.example.com/view/Production/"},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	cfg := testConfig(server.URL)
	client, _ := NewClient(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	views, err := client.GetViews(ctx)
	if err != nil {
		t.Fatalf("GetViews failed: %v", err)
	}

	if len(views) != 2 {
		t.Fatalf("expected 2 views, got %d", len(views))
	}
	if views[0].Name != "All" {
		t.Errorf("expected first view name 'All', got %s", views[0].Name)
	}
	if views[1].Name != "Production" {
		t.Errorf("expected second view name 'Production', got %s", views[1].Name)
	}
}

func TestGetNodes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, _, ok := r.BasicAuth()
		if !ok || user != "testuser" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		response := map[string]interface{}{
			"computer": []map[string]interface{}{
				{
					"displayName":  "master",
					"offline":      false,
					"numExecutors": 2,
					"executors": []map[string]interface{}{
						{"currentExecutable": nil},
						{"currentExecutable": nil},
					},
					"assignedLabels": []map[string]string{
						{"name": "master"},
					},
				},
				{
					"displayName":        "agent-1",
					"offline":            true,
					"numExecutors":       4,
					"offlineCauseReason": "Maintenance",
					"executors":          []map[string]interface{}{},
					"assignedLabels": []map[string]string{
						{"name": "linux"},
						{"name": "docker"},
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	cfg := testConfig(server.URL)
	client, _ := NewClient(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	nodes, err := client.GetNodes(ctx)
	if err != nil {
		t.Fatalf("GetNodes failed: %v", err)
	}

	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(nodes))
	}

	if nodes[0].DisplayName != "master" {
		t.Errorf("expected first node 'master', got %s", nodes[0].DisplayName)
	}
	if nodes[0].Offline {
		t.Error("expected master to be online")
	}

	if nodes[1].DisplayName != "agent-1" {
		t.Errorf("expected second node 'agent-1', got %s", nodes[1].DisplayName)
	}
	if !nodes[1].Offline {
		t.Error("expected agent-1 to be offline")
	}
	if nodes[1].OfflineCauseReason != "Maintenance" {
		t.Errorf("expected offline reason 'Maintenance', got %s", nodes[1].OfflineCauseReason)
	}
}

func TestGetQueue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, _, ok := r.BasicAuth()
		if !ok || user != "testuser" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		response := map[string]interface{}{
			"items": []map[string]interface{}{
				{
					"id": 123,
					"task": map[string]string{
						"name": "test-job",
						"url":  "https://jenkins.example.com/job/test-job/",
					},
					"why":          "Waiting for next available executor",
					"inQueueSince": 1609459200000,
					"buildable":    true,
					"blocked":      false,
					"stuck":        false,
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	cfg := testConfig(server.URL)
	client, _ := NewClient(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	queue, err := client.GetQueue(ctx)
	if err != nil {
		t.Fatalf("GetQueue failed: %v", err)
	}

	if len(queue.Items) != 1 {
		t.Fatalf("expected 1 queue item, got %d", len(queue.Items))
	}

	item := queue.Items[0]
	if item.ID != 123 {
		t.Errorf("expected item ID 123, got %d", item.ID)
	}
	if item.Task.Name != "test-job" {
		t.Errorf("expected task name 'test-job', got %s", item.Task.Name)
	}
}

func TestAuthenticationFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	cfg := testConfig(server.URL)
	client, _ := NewClient(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.GetRootInfo(ctx)
	if err == nil {
		t.Error("expected authentication error")
	}
}

func TestForbiddenError(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		// Always return forbidden (no crumb endpoint either)
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	cfg := testConfig(server.URL)
	client, _ := NewClient(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.GetRootInfo(ctx)
	if err == nil {
		t.Error("expected forbidden error")
	}
}

func TestBuildTreeParam(t *testing.T) {
	tests := []struct {
		fields   []string
		expected string
	}{
		{
			fields:   []string{"name", "url"},
			expected: "tree=name%2Curl",
		},
		{
			fields:   []string{"jobs[name,url,color]"},
			expected: "tree=jobs%5Bname%2Curl%2Ccolor%5D",
		},
	}

	for _, tt := range tests {
		result := buildTreeParam(tt.fields...)
		if result != tt.expected {
			t.Errorf("buildTreeParam(%v) = %s, want %s", tt.fields, result, tt.expected)
		}
	}
}

func TestEncodeJobPath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "simple-job",
			expected: "simple-job",
		},
		{
			input:    "folder/job",
			expected: "folder/job/job",
		},
		{
			input:    "my folder/my job",
			expected: "my%20folder/job/my%20job",
		},
	}

	for _, tt := range tests {
		result := encodeJobPath(tt.input)
		if result != tt.expected {
			t.Errorf("encodeJobPath(%s) = %s, want %s", tt.input, result, tt.expected)
		}
	}
}

func TestItoa(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{0, "0"},
		{1, "1"},
		{42, "42"},
		{123, "123"},
		{-5, "-5"},
	}

	for _, tt := range tests {
		result := itoa(tt.input)
		if result != tt.expected {
			t.Errorf("itoa(%d) = %s, want %s", tt.input, result, tt.expected)
		}
	}
}
