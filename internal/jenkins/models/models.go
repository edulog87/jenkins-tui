// Package models contains the data structures for Jenkins API responses.
package models

import (
	"sort"
	"time"
)

// ═══════════════════════════════════════════════════════════════════════════════
// ROOT INFO
// ═══════════════════════════════════════════════════════════════════════════════

// RootInfo represents the root Jenkins API response
type RootInfo struct {
	Mode            string `json:"mode"`
	NodeDescription string `json:"nodeDescription"`
	NodeName        string `json:"nodeName"`
	NumExecutors    int    `json:"numExecutors"`
	Description     string `json:"description"`
	UseCrumbs       bool   `json:"useCrumbs"`
	UseSecurity     bool   `json:"useSecurity"`
	Version         string `json:"version,omitempty"`
	URL             string `json:"url,omitempty"`
}

// ═══════════════════════════════════════════════════════════════════════════════
// VIEWS
// ═══════════════════════════════════════════════════════════════════════════════

// View represents a Jenkins view
type View struct {
	Name        string   `json:"name"`
	URL         string   `json:"url"`
	Description string   `json:"description,omitempty"`
	Jobs        []JobRef `json:"jobs,omitempty"` // For job count
}

// JobCount returns the number of jobs in this view
func (v *View) JobCount() int {
	return len(v.Jobs)
}

// JobRef is a minimal job reference (just name)
type JobRef struct {
	Name string `json:"name"`
}

// ═══════════════════════════════════════════════════════════════════════════════
// JOBS
// ═══════════════════════════════════════════════════════════════════════════════

// Job represents a Jenkins job (basic info)
type Job struct {
	Name         string         `json:"name"`
	URL          string         `json:"url"`
	Color        string         `json:"color"`
	FullName     string         `json:"fullName,omitempty"`
	DisplayName  string         `json:"displayName,omitempty"`
	LastBuild    *BuildRef      `json:"lastBuild,omitempty"`
	HealthReport []HealthReport `json:"healthReport,omitempty"`
	Buildable    bool           `json:"buildable,omitempty"`
	InQueue      bool           `json:"inQueue,omitempty"`
}

// JobDetail represents detailed job information
type JobDetail struct {
	Name                string         `json:"name"`
	URL                 string         `json:"url"`
	FullName            string         `json:"fullName,omitempty"`
	DisplayName         string         `json:"displayName,omitempty"`
	FullDisplayName     string         `json:"fullDisplayName,omitempty"`
	Color               string         `json:"color"`
	Description         string         `json:"description"`
	Buildable           bool           `json:"buildable"`
	InQueue             bool           `json:"inQueue"`
	NextBuildNumber     int            `json:"nextBuildNumber,omitempty"`
	LastBuild           *BuildRef      `json:"lastBuild,omitempty"`
	LastSuccessfulBuild *BuildRef      `json:"lastSuccessfulBuild,omitempty"`
	LastFailedBuild     *BuildRef      `json:"lastFailedBuild,omitempty"`
	LastStableBuild     *BuildRef      `json:"lastStableBuild,omitempty"`
	LastUnstableBuild   *BuildRef      `json:"lastUnstableBuild,omitempty"`
	LastCompletedBuild  *BuildRef      `json:"lastCompletedBuild,omitempty"`
	HealthReport        []HealthReport `json:"healthReport,omitempty"`
	Builds              []BuildRef     `json:"builds,omitempty"`
	Property            []JobProperty  `json:"property,omitempty"`
	ConcurrentBuild     bool           `json:"concurrentBuild,omitempty"`
	ResumeBlocked       bool           `json:"resumeBlocked,omitempty"`
}

// JobProperty represents a job property (parameters, etc)
type JobProperty struct {
	Class                string         `json:"_class,omitempty"`
	ParameterDefinitions []ParameterDef `json:"parameterDefinitions,omitempty"`
}

// ParameterDef represents a job parameter definition
type ParameterDef struct {
	Name         string      `json:"name"`
	Description  string      `json:"description,omitempty"`
	Type         string      `json:"type,omitempty"`
	DefaultValue interface{} `json:"defaultParameterValue,omitempty"`
	Choices      []string    `json:"choices,omitempty"`
}

// ═══════════════════════════════════════════════════════════════════════════════
// BUILDS
// ═══════════════════════════════════════════════════════════════════════════════

// BuildRef represents a reference to a build
type BuildRef struct {
	Number    int     `json:"number"`
	Result    string  `json:"result,omitempty"`
	Timestamp int64   `json:"timestamp,omitempty"`
	Duration  int64   `json:"duration,omitempty"`
	URL       string  `json:"url,omitempty"`
	Building  bool    `json:"building,omitempty"`
	Stages    []Stage `json:"stages,omitempty"`
}

// Build represents detailed build information
type Build struct {
	Number            int         `json:"number"`
	Result            string      `json:"result"`
	Timestamp         int64       `json:"timestamp"`
	Duration          int64       `json:"duration"`
	EstimatedDuration int64       `json:"estimatedDuration"`
	URL               string      `json:"url"`
	Building          bool        `json:"building"`
	DisplayName       string      `json:"displayName"`
	FullDisplayName   string      `json:"fullDisplayName,omitempty"`
	Description       string      `json:"description"`
	KeepLog           bool        `json:"keepLog,omitempty"`
	QueueID           int64       `json:"queueId,omitempty"`
	Artifacts         []Artifact  `json:"artifacts,omitempty"`
	ChangeSets        []ChangeSet `json:"changeSets,omitempty"`
	Actions           []Action    `json:"actions,omitempty"`
}

// Action represents a build action (causes, parameters, etc)
type Action struct {
	Class      string       `json:"_class,omitempty"`
	Causes     []BuildCause `json:"causes,omitempty"`
	Parameters []Parameter  `json:"parameters,omitempty"`
}

// Parameter represents a build parameter
type Parameter struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

// GetCauses extracts causes from build actions
func (b *Build) GetCauses() []BuildCause {
	var causes []BuildCause
	for _, action := range b.Actions {
		causes = append(causes, action.Causes...)
	}
	return causes
}

// GetParameters extracts parameters from build actions
func (b *Build) GetParameters() []Parameter {
	var params []Parameter
	for _, action := range b.Actions {
		params = append(params, action.Parameters...)
	}
	return params
}

// Artifact represents a build artifact
type Artifact struct {
	FileName     string `json:"fileName"`
	RelativePath string `json:"relativePath"`
	DisplayPath  string `json:"displayPath,omitempty"`
}

// ChangeSet represents a set of changes
type ChangeSet struct {
	Kind  string       `json:"kind,omitempty"`
	Items []ChangeItem `json:"items,omitempty"`
}

// ChangeItem represents a single change/commit
type ChangeItem struct {
	Msg           string   `json:"msg"`
	Author        Author   `json:"author"`
	CommitID      string   `json:"commitId"`
	Timestamp     int64    `json:"timestamp"`
	Comment       string   `json:"comment,omitempty"`
	AffectedPaths []string `json:"affectedPaths,omitempty"`
}

// Author represents a commit author
type Author struct {
	FullName string `json:"fullName"`
	ID       string `json:"id,omitempty"`
}

// BuildCause represents why a build was triggered
type BuildCause struct {
	Class            string `json:"_class,omitempty"`
	ShortDescription string `json:"shortDescription"`
	UserName         string `json:"userName,omitempty"`
	UserID           string `json:"userId,omitempty"`
	UpstreamProject  string `json:"upstreamProject,omitempty"`
	UpstreamBuild    int    `json:"upstreamBuild,omitempty"`
}

// ═══════════════════════════════════════════════════════════════════════════════
// PIPELINE STAGES (Workflow API - wfapi)
// ═══════════════════════════════════════════════════════════════════════════════

// WFAPIRun represents a pipeline run from the Workflow API
type WFAPIRun struct {
	Links               map[string]WFAPILink `json:"_links"`
	ID                  string               `json:"id"`
	Name                string               `json:"name"`
	Status              string               `json:"status"`
	StartTimeMillis     int64                `json:"startTimeMillis"`
	EndTimeMillis       int64                `json:"endTimeMillis"`
	DurationMillis      int64                `json:"durationMillis"`
	QueueDurationMillis int64                `json:"queueDurationMillis"`
	PauseDurationMillis int64                `json:"pauseDurationMillis"`
	Stages              []WFAPIStage         `json:"stages"`
}

// WFAPIStage represents a pipeline stage from Workflow API
type WFAPIStage struct {
	Links               map[string]WFAPILink `json:"_links"`
	ID                  string               `json:"id"`
	Name                string               `json:"name"`
	ExecNode            string               `json:"execNode"`
	Status              string               `json:"status"`
	StartTimeMillis     int64                `json:"startTimeMillis"`
	DurationMillis      int64                `json:"durationMillis"`
	PauseDurationMillis int64                `json:"pauseDurationMillis"`
	StageFlowNodes      []WFAPIFlowNode      `json:"stageFlowNodes"`
}

// WFAPIFlowNode represents a flow node within a stage
type WFAPIFlowNode struct {
	Links                map[string]WFAPILink `json:"_links"`
	ID                   string               `json:"id"`
	Name                 string               `json:"name"`
	ExecNode             string               `json:"execNode"`
	Status               string               `json:"status"`
	ParameterDescription string               `json:"parameterDescription,omitempty"`
	StartTimeMillis      int64                `json:"startTimeMillis"`
	DurationMillis       int64                `json:"durationMillis"`
	PauseDurationMillis  int64                `json:"pauseDurationMillis"`
	ParentNodes          []string             `json:"parentNodes"`
}

// WFAPILink represents a link in the Workflow API response
type WFAPILink struct {
	Href string `json:"href"`
}

// PipelineRun represents a pipeline run with stages
type PipelineRun struct {
	ID                string  `json:"id"`
	Name              string  `json:"name"`
	Status            string  `json:"status"`
	Result            string  `json:"result"`
	State             string  `json:"state"`
	StartTimeMillis   int64   `json:"startTimeMillis"`
	EndTimeMillis     int64   `json:"endTimeMillis"`
	DurationMillis    int64   `json:"durationMillis"`
	EstimatedDuration int64   `json:"estimatedDurationMillis,omitempty"`
	Stages            []Stage `json:"stages,omitempty"`
}

// Stage represents a pipeline stage
type Stage struct {
	ID                  string `json:"id"`
	Name                string `json:"name"`
	Status              string `json:"status"`
	Result              string `json:"result,omitempty"`
	State               string `json:"state,omitempty"`
	StartTimeMillis     int64  `json:"startTimeMillis"`
	DurationMillis      int64  `json:"durationMillis"`
	PauseDurationMillis int64  `json:"pauseDurationMillis,omitempty"`
	ExecNode            string `json:"execNode,omitempty"`
	Type                string `json:"type,omitempty"`
}

// StageLog represents the log for a stage
type StageLog struct {
	NodeID string `json:"nodeId"`
	Text   string `json:"text"`
	Length int64  `json:"length"`
}

// ═══════════════════════════════════════════════════════════════════════════════
// HEALTH
// ═══════════════════════════════════════════════════════════════════════════════

// HealthReport represents job health information
type HealthReport struct {
	Description   string `json:"description"`
	IconClassName string `json:"iconClassName,omitempty"`
	IconURL       string `json:"iconUrl,omitempty"`
	Score         int    `json:"score"`
}

// ═══════════════════════════════════════════════════════════════════════════════
// QUEUE
// ═══════════════════════════════════════════════════════════════════════════════

// Queue represents the Jenkins build queue
type Queue struct {
	Items []QueueItem `json:"items"`
}

// QueueItem represents an item in the build queue
type QueueItem struct {
	ID                   int64   `json:"id"`
	Task                 TaskRef `json:"task"`
	Why                  string  `json:"why"`
	InQueueSince         int64   `json:"inQueueSince"`
	Buildable            bool    `json:"buildable"`
	Blocked              bool    `json:"blocked"`
	Stuck                bool    `json:"stuck"`
	BuildableStartMillis int64   `json:"buildableStartMilliseconds,omitempty"`
	WaitingFor           string  `json:"waitingFor,omitempty"`
}

// QueueWaitTime returns how long the item has been waiting
func (q *QueueItem) QueueWaitTime() time.Duration {
	if q.InQueueSince == 0 {
		return 0
	}
	return time.Since(time.UnixMilli(q.InQueueSince))
}

// TaskRef represents a reference to a task/job
type TaskRef struct {
	Name  string `json:"name"`
	URL   string `json:"url"`
	Color string `json:"color,omitempty"`
}

// ═══════════════════════════════════════════════════════════════════════════════
// NODES
// ═══════════════════════════════════════════════════════════════════════════════

// Node represents a Jenkins node/agent
type Node struct {
	DisplayName         string                 `json:"displayName"`
	Description         string                 `json:"description,omitempty"`
	Offline             bool                   `json:"offline"`
	TemporarilyOffline  bool                   `json:"temporarilyOffline"`
	NumExecutors        int                    `json:"numExecutors"`
	Executors           []Executor             `json:"executors"`
	AssignedLabels      []Label                `json:"assignedLabels"`
	OfflineCauseReason  string                 `json:"offlineCauseReason"`
	Idle                bool                   `json:"idle"`
	JnlpAgent           bool                   `json:"jnlpAgent,omitempty"`
	LaunchSupported     bool                   `json:"launchSupported,omitempty"`
	ManualLaunchAllowed bool                   `json:"manualLaunchAllowed,omitempty"`
	MonitorData         map[string]interface{} `json:"monitorData,omitempty"`
}

// GetDiskSpace returns available disk space (if monitored)
func (n *Node) GetDiskSpace() (available, total int64) {
	if n.MonitorData == nil {
		return 0, 0
	}
	if disk, ok := n.MonitorData["hudson.node_monitors.DiskSpaceMonitor"]; ok {
		if m, ok := disk.(map[string]interface{}); ok {
			if size, ok := m["size"].(float64); ok {
				return int64(size), 0
			}
		}
	}
	return 0, 0
}

// BusyExecutors returns the count of busy executors
func (n *Node) BusyExecutors() int {
	count := 0
	for _, exec := range n.Executors {
		if exec.CurrentExecutable.URL != "" {
			count++
		}
	}
	return count
}

// Executor represents an executor on a node
type Executor struct {
	CurrentExecutable ExecutableRef `json:"currentExecutable"`
	Idle              bool          `json:"idle,omitempty"`
	LikelyStuck       bool          `json:"likelyStuck,omitempty"`
	Number            int           `json:"number,omitempty"`
	Progress          int           `json:"progress,omitempty"`
}

// ExecutableRef represents a reference to a running executable
type ExecutableRef struct {
	URL               string `json:"url,omitempty"`
	Number            int    `json:"number,omitempty"`
	DisplayName       string `json:"displayName,omitempty"`
	FullDisplayName   string `json:"fullDisplayName,omitempty"`
	Timestamp         int64  `json:"timestamp,omitempty"`
	EstimatedDuration int64  `json:"estimatedDuration,omitempty"`
}

// GetProgress calculates the progress percentage
func (e *ExecutableRef) GetProgress() int {
	if e.EstimatedDuration == 0 || e.Timestamp == 0 {
		return 0
	}
	elapsed := time.Since(time.UnixMilli(e.Timestamp)).Milliseconds()
	progress := int((elapsed * 100) / e.EstimatedDuration)
	if progress > 100 {
		progress = 100
	}
	if progress < 0 {
		progress = 0
	}
	return progress
}

// Label represents a node label
type Label struct {
	Name string `json:"name"`
}

// ═══════════════════════════════════════════════════════════════════════════════
// RUNNING BUILDS
// ═══════════════════════════════════════════════════════════════════════════════

// RunningBuild represents a currently running build
type RunningBuild struct {
	URL             string
	Number          int
	NodeName        string
	JobName         string
	FullDisplayName string
	Progress        int
	EstimatedEnd    time.Time
}

// ═══════════════════════════════════════════════════════════════════════════════
// HELPER METHODS
// ═══════════════════════════════════════════════════════════════════════════════

// IsRunning returns true if the build status indicates running
func (j *Job) IsRunning() bool {
	return isAnimatedColor(j.Color)
}

// IsDisabled returns true if the job is disabled
func (j *Job) IsDisabled() bool {
	return j.Color == "disabled" || j.Color == "disabled_anime"
}

// GetHealthScore returns the primary health score
func (j *Job) GetHealthScore() int {
	if len(j.HealthReport) > 0 {
		return j.HealthReport[0].Score
	}
	return -1
}

// IsRunning returns true if the build status indicates running
func (b *Build) IsRunning() bool {
	return b.Building
}

// StatusText returns a human-readable status
func (b *Build) StatusText() string {
	if b.Building {
		return "RUNNING"
	}
	if b.Result == "" {
		return "UNKNOWN"
	}
	return b.Result
}

// GetTimestamp returns the build start time
func (b *Build) GetTimestamp() time.Time {
	return time.UnixMilli(b.Timestamp)
}

// GetDuration returns the build duration
func (b *Build) GetDuration() time.Duration {
	return time.Duration(b.Duration) * time.Millisecond
}

// GetProgress returns build progress percentage (estimated)
func (b *Build) GetProgress() int {
	if !b.Building || b.EstimatedDuration == 0 {
		return 100
	}
	elapsed := time.Since(b.GetTimestamp())
	progress := int((elapsed.Milliseconds() * 100) / b.EstimatedDuration)
	if progress > 100 {
		progress = 100
	}
	return progress
}

// isAnimatedColor checks if the color indicates a running build
func isAnimatedColor(color string) bool {
	switch color {
	case "blue_anime", "red_anime", "yellow_anime", "grey_anime", "aborted_anime", "notbuilt_anime":
		return true
	}
	return false
}

// ═══════════════════════════════════════════════════════════════════════════════
// SORTING
// ═══════════════════════════════════════════════════════════════════════════════

// SortJobsByLastBuild sorts jobs by last build timestamp (most recent first)
func SortJobsByLastBuild(jobs []Job) {
	sort.Slice(jobs, func(i, j int) bool {
		ti := int64(0)
		tj := int64(0)
		if jobs[i].LastBuild != nil {
			ti = jobs[i].LastBuild.Timestamp
		}
		if jobs[j].LastBuild != nil {
			tj = jobs[j].LastBuild.Timestamp
		}
		return ti > tj
	})
}

// SortBuildsByNumber sorts builds by number (most recent first)
func SortBuildsByNumber(builds []BuildRef) {
	sort.Slice(builds, func(i, j int) bool {
		return builds[i].Number > builds[j].Number
	})
}

// SortStagesByStartTime sorts stages by start time
func SortStagesByStartTime(stages []Stage) {
	sort.Slice(stages, func(i, j int) bool {
		return stages[i].StartTimeMillis < stages[j].StartTimeMillis
	})
}
