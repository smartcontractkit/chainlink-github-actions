package main

import (
	"testing"
	"time"

	"github.com/google/go-github/v47/github"
	"github.com/stretchr/testify/assert"
)

func createWorkflowRun(id int64, status string, createdAt time.Time) *github.WorkflowRun {
	return &github.WorkflowRun{
		ID:        &id,
		Status:    &status,
		CreatedAt: &github.Timestamp{Time: createdAt},
	}
}

func TestGetLatestWorkflowRunFromList(t *testing.T) {
	var twoInt64 int64 = 2
	count := 2
	now := time.Now()
	d, _ := time.ParseDuration("2s")
	runsArray := []*github.WorkflowRun{
		createWorkflowRun(1, "a", now),
		createWorkflowRun(2, "b", now.Add(d)),
	}
	runs := &github.WorkflowRuns{
		TotalCount:   &count,
		WorkflowRuns: runsArray,
	}
	foundRun, _ := GetLatestWorkflowRunFromList(runs)

	assert.Equal(t, twoInt64, *foundRun.ID)
	assert.Equal(t, "b", *foundRun.Status)
}
