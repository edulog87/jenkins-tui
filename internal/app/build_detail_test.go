package app

import (
	"strings"
	"testing"

	"github.com/elogrono/jenkins-tui/internal/jenkins/models"
)

func TestBuildDetailView(t *testing.T) {
	m := &BuildsModel{
		width:  100,
		height: 40,
		buildDetail: &models.Build{
			Number:    123,
			Timestamp: 1609459200000, // 2021-01-01
			Duration:  60000,
			Result:    "SUCCESS",
			Building:  false,
			URL:       "http://jenkins/job/test/123/",
			ChangeSets: []models.ChangeSet{
				{
					Items: []models.ChangeItem{
						{
							Msg:      "Fix bug",
							CommitID: "abcdef123456",
							Author:   models.Author{FullName: "John Doe"},
						},
					},
				},
			},
			Artifacts: []models.Artifact{
				{FileName: "app.jar"},
			},
		},
		jobDetail: &models.JobDetail{
			Name: "test-job",
		},
		pipelineRun: &models.PipelineRun{
			Stages: []models.Stage{
				{
					Name:           "Checkout",
					Status:         "SUCCESS",
					DurationMillis: 5000,
				},
				{
					Name:           "Build",
					Status:         "SUCCESS",
					DurationMillis: 25000,
				},
			},
		},
	}

	view := m.viewBuildDetail()

	// Check for key elements in the view
	expectedStrings := []string{
		"test-job",
		"#123",
		"SUCCESS",
		"Started",
		"2021-01-01",
		"Duration",
		"1m 0s",
		"Changes",
		"Fix bug",
		"abcdef1",
		"John Doe",
		"Artifacts",
		"app.jar",
		"Pipeline Stages",
		"Checkout",
		"Build",
	}

	for _, s := range expectedStrings {
		if !strings.Contains(view, s) {
			t.Errorf("expected view to contain %q", s)
		}
	}
}
