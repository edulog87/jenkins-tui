package jenkins

import (
	"context"
	"fmt"
	"net/http"

	"github.com/elogrono/jenkins-tui/internal/jenkins/models"
)

// GetRootInfo fetches basic Jenkins server information
func (c *Client) GetRootInfo(ctx context.Context) (*models.RootInfo, error) {
	var info models.RootInfo
	err := c.getJSON(ctx, "/api/json?"+buildTreeParam(
		"mode",
		"nodeDescription",
		"nodeName",
		"numExecutors",
		"description",
		"useCrumbs",
		"useSecurity",
	), &info)
	return &info, err
}

// GetViews fetches all views with job count
func (c *Client) GetViews(ctx context.Context) ([]models.View, error) {
	var resp struct {
		Views []models.View `json:"views"`
	}
	// Include jobs array to get job count in same request
	err := c.getJSON(ctx, "/api/json?"+buildTreeParam("views[name,url,jobs[name]]"), &resp)
	return resp.Views, err
}

// GetViewJobs fetches jobs for a specific view
func (c *Client) GetViewJobs(ctx context.Context, viewName string) ([]models.Job, error) {
	path := "/view/" + encodeJobPath(viewName) + "/api/json?" + buildTreeParam(
		"jobs[name,url,color,lastBuild[number,result,timestamp,duration],healthReport[description,score]]",
	)

	var resp struct {
		Jobs []models.Job `json:"jobs"`
	}
	err := c.getJSON(ctx, path, &resp)
	return resp.Jobs, err
}

// GetAllJobs fetches all jobs (from all view)
func (c *Client) GetAllJobs(ctx context.Context) ([]models.Job, error) {
	var resp struct {
		Jobs []models.Job `json:"jobs"`
	}
	err := c.getJSON(ctx, "/api/json?"+buildTreeParam(
		"jobs[name,url,color,lastBuild[number,result,timestamp,duration],healthReport[description,score]]",
	), &resp)
	return resp.Jobs, err
}

// GetJob fetches details for a specific job
func (c *Client) GetJob(ctx context.Context, jobName string) (*models.JobDetail, error) {
	path := "/job/" + encodeJobPath(jobName) + "/api/json?" + buildTreeParam(
		"name",
		"url",
		"color",
		"description",
		"buildable",
		"inQueue",
		"lastBuild[number,result,timestamp,duration,url]",
		"lastSuccessfulBuild[number,timestamp]",
		"lastFailedBuild[number,timestamp]",
		"healthReport[description,score]",
		"builds[number,result,timestamp,duration,url]",
	)

	var job models.JobDetail
	err := c.getJSON(ctx, path, &job)
	if err != nil {
		return nil, err
	}

	// Try to get stages for each build from Blue Ocean API
	// Note: Doing this sequentially can be slow if there are many builds
	// For now, let's limit it or just do it for the first few
	for i := 0; i < len(job.Builds) && i < 10; i++ {
		run, _ := c.GetPipelineRun(ctx, jobName, job.Builds[i].Number)
		if run != nil {
			job.Builds[i].Stages = run.Stages
		}
	}

	return &job, nil
}

// GetBuild fetches details for a specific build
func (c *Client) GetBuild(ctx context.Context, jobName string, buildNumber int) (*models.Build, error) {
	path := "/job/" + encodeJobPath(jobName) + "/" + itoa(buildNumber) + "/api/json?" + buildTreeParam(
		"number",
		"result",
		"timestamp",
		"duration",
		"estimatedDuration",
		"url",
		"building",
		"displayName",
		"description",
		"executor[currentExecutable[url]]",
		"artifacts[fileName,relativePath]",
		"changeSets[items[msg,author[fullName],commitId,timestamp]]",
		"causes[shortDescription,userName,userId]",
	)

	var build models.Build
	err := c.getJSON(ctx, path, &build)
	return &build, err
}

// TriggerBuild triggers a build for a specific job
func (c *Client) TriggerBuild(ctx context.Context, jobName string) error {
	path := "/job/" + encodeJobPath(jobName) + "/build"
	resp, err := c.doRequest(ctx, http.MethodPost, path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d triggering build", resp.StatusCode)
	}
	return nil
}

// GetBuildLog fetches the console output for a build
func (c *Client) GetBuildLog(ctx context.Context, jobName string, buildNumber int, maxBytes int) (string, error) {
	path := "/job/" + encodeJobPath(jobName) + "/" + itoa(buildNumber) + "/consoleText"
	return c.getText(ctx, path, maxBytes)
}

// GetStageLog fetches the log for a specific pipeline stage using wfapi
func (c *Client) GetStageLog(ctx context.Context, jobName string, buildNumber int, stageID string) (string, error) {
	// Try wfapi first
	path := "/job/" + encodeJobPath(jobName) + "/" + itoa(buildNumber) + "/execution/node/" + stageID + "/wfapi/log"
	log, err := c.getText(ctx, path, 500000)
	if err == nil {
		return log, nil
	}

	// Fallback to Blue Ocean API
	path = "/blue/rest/organizations/jenkins/pipelines/" + encodeJobPath(jobName) + "/runs/" + itoa(buildNumber) + "/nodes/" + stageID + "/log/"
	return c.getText(ctx, path, 500000)
}

// GetQueue fetches the build queue
func (c *Client) GetQueue(ctx context.Context) (*models.Queue, error) {
	var queue models.Queue
	err := c.getJSON(ctx, "/queue/api/json?"+buildTreeParam(
		"items[id,task[name,url],why,inQueueSince,buildable,blocked,stuck]",
	), &queue)
	return &queue, err
}

// GetNodes fetches all nodes/computers with detailed executor info
func (c *Client) GetNodes(ctx context.Context) ([]models.Node, error) {
	var resp struct {
		Computer []models.Node `json:"computer"`
	}
	err := c.getJSON(ctx, "/computer/api/json?"+buildTreeParam(
		"computer[displayName,offline,temporarilyOffline,numExecutors,executors[currentExecutable[url,number,displayName,fullDisplayName,timestamp,estimatedDuration],idle,likelyStuck,number,progress],assignedLabels[name],offlineCauseReason,idle,monitorData[*]]",
	), &resp)
	return resp.Computer, err
}

// GetRunningBuilds fetches all currently running builds
func (c *Client) GetRunningBuilds(ctx context.Context) ([]models.RunningBuild, error) {
	nodes, err := c.GetNodes(ctx)
	if err != nil {
		return nil, err
	}

	var running []models.RunningBuild
	for _, node := range nodes {
		for _, executor := range node.Executors {
			if executor.CurrentExecutable.URL != "" {
				running = append(running, models.RunningBuild{
					URL:      executor.CurrentExecutable.URL,
					Number:   executor.CurrentExecutable.Number,
					NodeName: node.DisplayName,
				})
			}
		}
	}

	return running, nil
}

// GetPipelineRun fetches pipeline run details with stages using the Workflow API (wfapi)
// This is the official Jenkins Workflow API and is more reliable than Blue Ocean
// Endpoint: /job/{jobName}/wfapi/runs
func (c *Client) GetPipelineRun(ctx context.Context, jobName string, buildNumber int) (*models.PipelineRun, error) {
	// First try wfapi (official Workflow API)
	run, err := c.getPipelineRunFromWFAPI(ctx, jobName, buildNumber)
	if err == nil && run != nil {
		return run, nil
	}

	// Fallback to Blue Ocean API if wfapi fails
	return c.getPipelineRunFromBlueOcean(ctx, jobName, buildNumber)
}

// getPipelineRunFromWFAPI fetches stages from Jenkins Workflow API
func (c *Client) getPipelineRunFromWFAPI(ctx context.Context, jobName string, buildNumber int) (*models.PipelineRun, error) {
	path := "/job/" + encodeJobPath(jobName) + "/" + itoa(buildNumber) + "/wfapi/describe"

	var wfRun models.WFAPIRun
	err := c.getJSON(ctx, path, &wfRun)
	if err != nil {
		return nil, err
	}

	// Convert WFAPIStage to Stage
	stages := make([]models.Stage, 0, len(wfRun.Stages))
	for _, wfStage := range wfRun.Stages {
		stage := models.Stage{
			ID:                  wfStage.ID,
			Name:                wfStage.Name,
			Status:              wfStage.Status,
			Result:              wfStage.Status, // wfapi uses status field
			StartTimeMillis:     wfStage.StartTimeMillis,
			DurationMillis:      wfStage.DurationMillis,
			PauseDurationMillis: wfStage.PauseDurationMillis,
			ExecNode:            wfStage.ExecNode,
		}
		stages = append(stages, stage)
	}

	run := &models.PipelineRun{
		ID:              wfRun.ID,
		Name:            wfRun.Name,
		Status:          wfRun.Status,
		Result:          wfRun.Status,
		StartTimeMillis: wfRun.StartTimeMillis,
		EndTimeMillis:   wfRun.EndTimeMillis,
		DurationMillis:  wfRun.DurationMillis,
		Stages:          stages,
	}

	return run, nil
}

// getPipelineRunFromBlueOcean fetches pipeline run details from Blue Ocean API (fallback)
func (c *Client) getPipelineRunFromBlueOcean(ctx context.Context, jobName string, buildNumber int) (*models.PipelineRun, error) {
	// Blue Ocean API path for pipeline stages
	path := "/blue/rest/organizations/jenkins/pipelines/" + encodeJobPath(jobName) + "/runs/" + itoa(buildNumber) + "/nodes/"

	var stages []models.Stage
	err := c.getJSON(ctx, path, &stages)
	if err != nil {
		// Blue Ocean API might not be available, return nil without error
		return nil, nil
	}

	// Sort stages by start time
	models.SortStagesByStartTime(stages)

	// Calculate total duration and status from stages
	run := &models.PipelineRun{
		ID:     itoa(buildNumber),
		Stages: stages,
	}

	// Determine overall status from stages
	for _, stage := range stages {
		if stage.Status == "FAILED" || stage.Status == "FAILURE" {
			run.Status = "FAILED"
			run.Result = "FAILURE"
			break
		} else if stage.Status == "RUNNING" || stage.Status == "IN_PROGRESS" {
			run.Status = "RUNNING"
			run.State = "RUNNING"
		} else if stage.Status == "UNSTABLE" && run.Status != "RUNNING" {
			run.Status = "UNSTABLE"
			run.Result = "UNSTABLE"
		} else if run.Status == "" {
			run.Status = stage.Status
			run.Result = stage.Result
		}
	}

	// Calculate duration
	if len(stages) > 0 {
		run.StartTimeMillis = stages[0].StartTimeMillis
		lastStage := stages[len(stages)-1]
		run.EndTimeMillis = lastStage.StartTimeMillis + lastStage.DurationMillis
		run.DurationMillis = run.EndTimeMillis - run.StartTimeMillis
	}

	return run, nil
}

// itoa converts an int to string (simple helper)
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
