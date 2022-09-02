package main

import (
	"context"
	"fmt"
	"time"

	"github.com/smartcontractkit/chainlink-github-actions/public-to-private/run-workflow/action"

	"github.com/google/go-github/v47/github"
	"github.com/sethvargo/go-githubactions"
	"golang.org/x/oauth2"
)

const COMPLETED_STATUS = "completed"

var workflowStartedAt time.Time

// ConnectToGithub Sets up the github client and connects
func ConnectToGithub(inputs *action.Inputs) *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: inputs.GithubToken},
	)
	tc := oauth2.NewClient(context.Background(), ts)
	return github.NewClient(tc)
}

func StartWorkflow(client *github.Client, githubAction *githubactions.Action, inputs *action.Inputs) {
	githubAction.Infof("Starting workflow %s", inputs.WorkflowFile)
	event := &github.CreateWorkflowDispatchEventRequest{
		Ref:    inputs.Branch,
		Inputs: inputs.inputsToMap(githubAction),
	}
	resp, err := client.Actions.CreateWorkflowDispatchEventByFileName(context.Background(), inputs.Owner, inputs.Repository, inputs.WorkflowFile, *event)

	if err != nil {
		githubAction.Fatalf("Failed to make call to start workflow: %v", err)
	}
	if resp.StatusCode != 204 {
		githubAction.Fatalf("Failed to start the workflow: %v", err)
	}
}

func GetListOfWorkflowRuns(client *github.Client, githubAction *githubactions.Action, inputs *action.Inputs) (*github.WorkflowRuns, error) {
	opts := &github.ListWorkflowRunsOptions{
		Branch:  inputs.Branch,
		Event:   "workflow_dispatch",
		Created: fmt.Sprintf(">=%s", workflowStartedAt.Format(time.RFC3339)),
	}
	runs, resp, err := client.Actions.ListWorkflowRunsByFileName(context.Background(), inputs.Owner, inputs.Repository, inputs.WorkflowFile, opts)
	if err != nil {
		githubAction.Infof("failed to get the workflow: %v", err)
		return nil, err
	}
	if resp.StatusCode != 200 {
		err = fmt.Errorf("failed to get the workflow, status code %d", resp.StatusCode)
		githubAction.Infof("%v", err)
		return nil, err
	}
	if runs.GetTotalCount() < 1 {
		err = fmt.Errorf("failed fo find any workflow runs")
		githubAction.Infof("%v", err)
		return nil, err
	}
	return runs, nil
}

func GetLatestWorkflowRunFromList(runs *github.WorkflowRuns, githubAction *githubactions.Action) (*github.WorkflowRun, error) {
	var currentWorkflow *github.WorkflowRun
	var timestamp *github.Timestamp
	for i, wf := range runs.WorkflowRuns {
		githubAction.Infof("workflow at index %d had completion status of %s at %s with id %d", i, *wf.Status, wf.CreatedAt.Time.String(), *wf.ID)
		if timestamp == nil {
			timestamp = wf.CreatedAt
		}
		// we only care about uncompleted workflows
		if *wf.Status != COMPLETED_STATUS {
			if currentWorkflow == nil {
				currentWorkflow = wf
			}
			githubAction.Infof("found an active workflow")
			// we want the latest workflow so if this one is newer, use it
			if wf.CreatedAt.After(timestamp.Time) {
				githubAction.Infof("workflow is newer, using it unless we find a newer one")
				currentWorkflow = wf
			}
		}
	}
	if currentWorkflow == nil {
		err := fmt.Errorf("failed to find an unfinished workflow")
		githubAction.Infof("%v", err)
		return nil, err
	}
	return currentWorkflow, nil
}

// GetMostRecentWorkflowRunId Check the list of current workflows and get the most recent one that is not complete
func GetMostRecentWorkflowRunId(client *github.Client, githubAction *githubactions.Action, inputs *action.Inputs) int64 {
	maxRetrys := 5
	var err error
	var workflow *github.WorkflowRun
	var runs *github.WorkflowRuns
	for i := 0; i < maxRetrys; i++ {
		githubAction.Infof("Checking for active workflows attempt: %d", i+1)

		// if err is nil then it is the first time and we do not need to sleep, otherwise wait
		if err != nil {
			time.Sleep(inputs.RetryInterval)
		}

		// get the list
		runs, err = GetListOfWorkflowRuns(client, inputs)
		if err != nil || runs == nil {
			continue
		}

		// check the list for the most recent unfinished workflow
		workflow, err = GetLatestWorkflowRunFromList(runs, githubAction)
		if err != nil || workflow == nil {
			continue
		}

		// if we make it here then we have found the latest unfinished workflow
		break
	}

	// fail if we ended the retry loop in error
	if err != nil || workflow == nil {
		githubAction.Fatalf("did not find any active workflows: %v", err)
	}

	githubAction.Infof("Successfully found a workflow in progress: %d", *workflow.ID)
	return *workflow.ID
}

// GetWorkflowRun Gets the workflow run with updated status
func GetWorkflowRun(client *github.Client, githubAction *githubactions.Action, inputs *action.Inputs, workflowID int64) *github.WorkflowRun {
	wfr, resp, err := client.Actions.GetWorkflowRunByID(context.Background(), inputs.Owner, inputs.Repository, workflowID)
	if err != nil {
		githubAction.Infof("Failed to get the workflow run: %v", err)
		return nil
	}
	if resp.StatusCode != 200 {
		githubAction.Infof("Failed to get the workflow, found status code: %d", resp.StatusCode)
		return nil
	}
	return wfr
}

// PollWorkflow Poll the workflow until the status is complete or we hit the timeout
func PollWorkflow(githubAction githubactions.Action, inputs *action.Inputs, workflowID int64) *action.Outputs {
	githubAction.Infof("Beginning to poll for the workflows status every %v until it is complete or we timeout after %v", inputs.RetryInterval, inputs.Timeout)
	pollStart := time.Now()
	stop := false
	var status string
	var latestWorkflow *github.WorkflowRun
	testContext, testCancel := context.WithTimeout(context.Background(), inputs.Timeout)
	defer testCancel()
	ticker := time.NewTicker(inputs.RetryInterval)
	for {
		select {
		case <-ticker.C:
			// check the latest status, if it is completed then we stop
			latestWorkflow = GetWorkflowRun(inputs, workflowID)
			status = *latestWorkflow.Status
			t := time.Now()
			diff := t.Sub(pollStart)
			githubAction.Infof("current workflow run status is \"%v\" after %v", status, diff)
			stop = status == COMPLETED_STATUS
		case <-testContext.Done():
			// timed out
			stop = true
			ticker.Stop()
			break
		}

		// if stop is true, kill the loop
		if stop {
			break
		}
	}

	if status != COMPLETED_STATUS {
		githubAction.Fatalf("workflow did not reach completed status, id: %d", workflowID)
	}

	return &action.Outputs{
		Status:     status,
		Conclusion: *latestWorkflow.Conclusion,
		WorkflowID: workflowID,
	}
}

func main() {

	// get a new action object to use, avoid using the global so we can more easily test
	githubAction := githubactions.New()

	// get the inputs from the action
	inputs := action.GetInputs(githubAction)

	// setup the connection to github
	client := ConnectToGithub(inputs)

	// start the workflow
	workflowStartedAt = time.Now()
	StartWorkflow(client, githubAction, inputs)

	// get the workflow we want to poll
	workflowID := GetMostRecentWorkflowRunId(client, githubAction, inputs)

	// poll the workflow until it is finished
	outputs := PollWorkflow(githubAction, inputs, workflowID)

	// send the outputs back out to the action
	outputs.SetOutputs(githubAction)
}
