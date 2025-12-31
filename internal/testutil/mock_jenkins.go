// Package testutil provides test utilities for the jenkins-tui project.
package testutil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
)

// MockJenkins is a mock Jenkins server for testing
type MockJenkins struct {
	Server *httptest.Server

	mu           sync.RWMutex
	username     string
	password     string
	useCrumbs    bool
	crumb        string
	crumbField   string
	rootInfo     map[string]interface{}
	views        []map[string]interface{}
	jobs         []map[string]interface{}
	nodes        []map[string]interface{}
	queue        []map[string]interface{}
	builds       map[string][]map[string]interface{} // jobName -> builds
	buildDetails map[string]map[string]interface{}   // "jobName/buildNum" -> build
	logs         map[string]string                   // "jobName/buildNum" -> log content
}

// NewMockJenkins creates a new mock Jenkins server
func NewMockJenkins(username, password string) *MockJenkins {
	m := &MockJenkins{
		username:     username,
		password:     password,
		builds:       make(map[string][]map[string]interface{}),
		buildDetails: make(map[string]map[string]interface{}),
		logs:         make(map[string]string),
	}

	m.Server = httptest.NewServer(http.HandlerFunc(m.handler))
	return m
}

// Close closes the mock server
func (m *MockJenkins) Close() {
	m.Server.Close()
}

// URL returns the mock server URL
func (m *MockJenkins) URL() string {
	return m.Server.URL
}

// SetRootInfo sets the root info response
func (m *MockJenkins) SetRootInfo(info map[string]interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rootInfo = info
}

// SetViews sets the views response
func (m *MockJenkins) SetViews(views []map[string]interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.views = views
}

// SetJobs sets the jobs response
func (m *MockJenkins) SetJobs(jobs []map[string]interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.jobs = jobs
}

// SetNodes sets the nodes response
func (m *MockJenkins) SetNodes(nodes []map[string]interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.nodes = nodes
}

// SetQueue sets the queue response
func (m *MockJenkins) SetQueue(queue []map[string]interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.queue = queue
}

// EnableCrumbs enables CSRF protection
func (m *MockJenkins) EnableCrumbs(crumb, field string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.useCrumbs = true
	m.crumb = crumb
	m.crumbField = field
}

// AddBuildLog adds a build log
func (m *MockJenkins) AddBuildLog(jobName string, buildNum int, log string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := jobName + "/" + itoa(buildNum)
	m.logs[key] = log
}

func (m *MockJenkins) handler(w http.ResponseWriter, r *http.Request) {
	// Check authentication
	user, pass, ok := r.BasicAuth()
	if !ok || user != m.username || pass != m.password {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	path := r.URL.Path

	switch {
	case path == "/api/json":
		m.handleRootInfo(w, r)
	case path == "/crumbIssuer/api/json":
		m.handleCrumb(w, r)
	case path == "/queue/api/json":
		m.handleQueue(w, r)
	case path == "/computer/api/json":
		m.handleNodes(w, r)
	default:
		// Handle other paths
		w.WriteHeader(http.StatusNotFound)
	}
}

func (m *MockJenkins) handleRootInfo(w http.ResponseWriter, r *http.Request) {
	response := m.rootInfo
	if response == nil {
		response = map[string]interface{}{
			"mode":            "NORMAL",
			"nodeDescription": "Mock Jenkins",
			"nodeName":        "",
			"numExecutors":    2,
			"useCrumbs":       m.useCrumbs,
			"useSecurity":     true,
		}
	}

	// Add views if requested
	if m.views != nil {
		response["views"] = m.views
	}

	// Add jobs if requested
	if m.jobs != nil {
		response["jobs"] = m.jobs
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (m *MockJenkins) handleCrumb(w http.ResponseWriter, r *http.Request) {
	if !m.useCrumbs {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	response := map[string]string{
		"crumb":             m.crumb,
		"crumbRequestField": m.crumbField,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (m *MockJenkins) handleQueue(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"items": m.queue,
	}
	if m.queue == nil {
		response["items"] = []interface{}{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (m *MockJenkins) handleNodes(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"computer": m.nodes,
	}
	if m.nodes == nil {
		response["computer"] = []interface{}{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	negative := n < 0
	if negative {
		n = -n
	}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	if negative {
		digits = append([]byte{'-'}, digits...)
	}
	return string(digits)
}
