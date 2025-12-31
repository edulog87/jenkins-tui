package models

import (
	"encoding/json"
	"testing"
)

func TestJobIsRunning(t *testing.T) {
	tests := []struct {
		color    string
		expected bool
	}{
		{"blue", false},
		{"blue_anime", true},
		{"red", false},
		{"red_anime", true},
		{"yellow", false},
		{"yellow_anime", true},
		{"grey_anime", true},
		{"aborted", false},
		{"aborted_anime", true},
		{"notbuilt", false},
		{"notbuilt_anime", true},
		{"disabled", false},
	}

	for _, tt := range tests {
		job := &Job{Color: tt.color}
		if job.IsRunning() != tt.expected {
			t.Errorf("Job.IsRunning() for color %s = %v, want %v", tt.color, job.IsRunning(), tt.expected)
		}
	}
}

func TestBuildStatusText(t *testing.T) {
	tests := []struct {
		build    Build
		expected string
	}{
		{Build{Building: true}, "RUNNING"},
		{Build{Building: false, Result: "SUCCESS"}, "SUCCESS"},
		{Build{Building: false, Result: "FAILURE"}, "FAILURE"},
		{Build{Building: false, Result: "UNSTABLE"}, "UNSTABLE"},
		{Build{Building: false, Result: "ABORTED"}, "ABORTED"},
		{Build{Building: false, Result: ""}, "UNKNOWN"},
	}

	for _, tt := range tests {
		if tt.build.StatusText() != tt.expected {
			t.Errorf("Build.StatusText() = %s, want %s", tt.build.StatusText(), tt.expected)
		}
	}
}

func TestBuildIsRunning(t *testing.T) {
	runningBuild := Build{Building: true}
	if !runningBuild.IsRunning() {
		t.Error("expected running build to return true")
	}

	completedBuild := Build{Building: false}
	if completedBuild.IsRunning() {
		t.Error("expected completed build to return false")
	}
}

func TestJobParsing(t *testing.T) {
	jsonData := `{
		"name": "test-job",
		"url": "https://jenkins.example.com/job/test-job/",
		"color": "blue",
		"lastBuild": {
			"number": 42,
			"result": "SUCCESS",
			"timestamp": 1609459200000,
			"duration": 60000
		},
		"healthReport": [
			{"description": "Build stability: No recent builds failed.", "score": 100}
		]
	}`

	var job Job
	if err := json.Unmarshal([]byte(jsonData), &job); err != nil {
		t.Fatalf("failed to parse job JSON: %v", err)
	}

	if job.Name != "test-job" {
		t.Errorf("expected name 'test-job', got %s", job.Name)
	}
	if job.Color != "blue" {
		t.Errorf("expected color 'blue', got %s", job.Color)
	}
	if job.LastBuild == nil {
		t.Fatal("expected lastBuild to be present")
	}
	if job.LastBuild.Number != 42 {
		t.Errorf("expected build number 42, got %d", job.LastBuild.Number)
	}
	if len(job.HealthReport) != 1 {
		t.Fatalf("expected 1 health report, got %d", len(job.HealthReport))
	}
	if job.HealthReport[0].Score != 100 {
		t.Errorf("expected health score 100, got %d", job.HealthReport[0].Score)
	}
}

func TestJobParsingWithNulls(t *testing.T) {
	// Test that missing/null fields don't cause errors
	jsonData := `{
		"name": "minimal-job",
		"url": "https://jenkins.example.com/job/minimal-job/",
		"color": "notbuilt",
		"lastBuild": null,
		"healthReport": []
	}`

	var job Job
	if err := json.Unmarshal([]byte(jsonData), &job); err != nil {
		t.Fatalf("failed to parse job JSON with nulls: %v", err)
	}

	if job.Name != "minimal-job" {
		t.Errorf("expected name 'minimal-job', got %s", job.Name)
	}
	if job.LastBuild != nil {
		t.Error("expected lastBuild to be nil")
	}
	if job.IsRunning() {
		t.Error("expected job not to be running")
	}
}

func TestBuildParsing(t *testing.T) {
	jsonData := `{
		"number": 123,
		"result": "SUCCESS",
		"timestamp": 1609459200000,
		"duration": 120000,
		"estimatedDuration": 100000,
		"url": "https://jenkins.example.com/job/test-job/123/",
		"building": false,
		"displayName": "#123",
		"description": "Test build",
		"artifacts": [
			{"fileName": "app.jar", "relativePath": "target/app.jar"}
		],
		"changeSets": [
			{
				"items": [
					{
						"msg": "Fix bug in login",
						"author": {"fullName": "John Doe"},
						"commitId": "abc123",
						"timestamp": 1609459100000
					}
				]
			}
		],
		"actions": [
			{
				"_class": "hudson.model.CauseAction",
				"causes": [
					{"shortDescription": "Started by user admin", "userName": "admin", "userId": "admin"}
				]
			}
		]
	}`

	var build Build
	if err := json.Unmarshal([]byte(jsonData), &build); err != nil {
		t.Fatalf("failed to parse build JSON: %v", err)
	}

	if build.Number != 123 {
		t.Errorf("expected number 123, got %d", build.Number)
	}
	if build.Result != "SUCCESS" {
		t.Errorf("expected result 'SUCCESS', got %s", build.Result)
	}
	if build.Duration != 120000 {
		t.Errorf("expected duration 120000, got %d", build.Duration)
	}
	if build.Building {
		t.Error("expected building to be false")
	}
	if len(build.Artifacts) != 1 {
		t.Fatalf("expected 1 artifact, got %d", len(build.Artifacts))
	}
	if build.Artifacts[0].FileName != "app.jar" {
		t.Errorf("expected artifact 'app.jar', got %s", build.Artifacts[0].FileName)
	}
	if len(build.ChangeSets) != 1 || len(build.ChangeSets[0].Items) != 1 {
		t.Fatal("expected 1 changeset with 1 item")
	}
	if build.ChangeSets[0].Items[0].Author.FullName != "John Doe" {
		t.Errorf("expected author 'John Doe', got %s", build.ChangeSets[0].Items[0].Author.FullName)
	}
	causes := build.GetCauses()
	if len(causes) != 1 {
		t.Fatalf("expected 1 cause, got %d", len(causes))
	}
}

func TestNodeParsing(t *testing.T) {
	jsonData := `{
		"displayName": "agent-1",
		"offline": false,
		"temporarilyOffline": false,
		"numExecutors": 4,
		"executors": [
			{"currentExecutable": {"url": "https://jenkins.example.com/job/test/1/", "number": 1}},
			{"currentExecutable": null},
			{"currentExecutable": null},
			{"currentExecutable": null}
		],
		"assignedLabels": [
			{"name": "linux"},
			{"name": "docker"}
		],
		"offlineCauseReason": "",
		"idle": false
	}`

	var node Node
	if err := json.Unmarshal([]byte(jsonData), &node); err != nil {
		t.Fatalf("failed to parse node JSON: %v", err)
	}

	if node.DisplayName != "agent-1" {
		t.Errorf("expected displayName 'agent-1', got %s", node.DisplayName)
	}
	if node.Offline {
		t.Error("expected node to be online")
	}
	if node.NumExecutors != 4 {
		t.Errorf("expected 4 executors, got %d", node.NumExecutors)
	}
	if len(node.Executors) != 4 {
		t.Fatalf("expected 4 executor entries, got %d", len(node.Executors))
	}
	if node.Executors[0].CurrentExecutable.URL == "" {
		t.Error("expected first executor to have a running build")
	}
	if len(node.AssignedLabels) != 2 {
		t.Fatalf("expected 2 labels, got %d", len(node.AssignedLabels))
	}
}

func TestQueueParsing(t *testing.T) {
	jsonData := `{
		"items": [
			{
				"id": 456,
				"task": {"name": "queued-job", "url": "https://jenkins.example.com/job/queued-job/"},
				"why": "Waiting for next available executor",
				"inQueueSince": 1609459200000,
				"buildable": true,
				"blocked": false,
				"stuck": false
			}
		]
	}`

	var queue Queue
	if err := json.Unmarshal([]byte(jsonData), &queue); err != nil {
		t.Fatalf("failed to parse queue JSON: %v", err)
	}

	if len(queue.Items) != 1 {
		t.Fatalf("expected 1 queue item, got %d", len(queue.Items))
	}

	item := queue.Items[0]
	if item.ID != 456 {
		t.Errorf("expected ID 456, got %d", item.ID)
	}
	if item.Task.Name != "queued-job" {
		t.Errorf("expected task name 'queued-job', got %s", item.Task.Name)
	}
	if !item.Buildable {
		t.Error("expected item to be buildable")
	}
	if item.Stuck {
		t.Error("expected item not to be stuck")
	}
}

func TestRootInfoParsing(t *testing.T) {
	jsonData := `{
		"mode": "NORMAL",
		"nodeDescription": "the master Jenkins node",
		"nodeName": "",
		"numExecutors": 2,
		"description": "Jenkins CI Server",
		"useCrumbs": true,
		"useSecurity": true
	}`

	var info RootInfo
	if err := json.Unmarshal([]byte(jsonData), &info); err != nil {
		t.Fatalf("failed to parse root info JSON: %v", err)
	}

	if info.Mode != "NORMAL" {
		t.Errorf("expected mode 'NORMAL', got %s", info.Mode)
	}
	if info.NumExecutors != 2 {
		t.Errorf("expected 2 executors, got %d", info.NumExecutors)
	}
	if !info.UseCrumbs {
		t.Error("expected useCrumbs to be true")
	}
	if !info.UseSecurity {
		t.Error("expected useSecurity to be true")
	}
}

func TestViewParsing(t *testing.T) {
	jsonData := `{
		"name": "Production",
		"url": "https://jenkins.example.com/view/Production/"
	}`

	var view View
	if err := json.Unmarshal([]byte(jsonData), &view); err != nil {
		t.Fatalf("failed to parse view JSON: %v", err)
	}

	if view.Name != "Production" {
		t.Errorf("expected name 'Production', got %s", view.Name)
	}
	if view.URL != "https://jenkins.example.com/view/Production/" {
		t.Errorf("unexpected URL: %s", view.URL)
	}
}
