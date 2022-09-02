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

// ConnectToGithub Sets up the github client and connects
func ConnectToGithub(githubToken string) *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubToken},
	)
	tc := oauth2.NewClient(context.Background(), ts)
	return github.NewClient(tc)
}

func StartWorkflow(client *github.Client, inputs *action.Inputs) error {
	fmt.Printf("Starting workflow %s\n", inputs.WorkflowFile)
	inputsMap, err := inputs.InputsToMap()
	if err != nil {
		return err
	}
	event := &github.CreateWorkflowDispatchEventRequest{
		Ref:    inputs.Branch,
		Inputs: inputsMap,
	}
	resp, err := client.Actions.CreateWorkflowDispatchEventByFileName(context.Background(), inputs.Owner, inputs.Repository, inputs.WorkflowFile, *event)

	if err != nil {
		return err
	}
	if resp.StatusCode != 204 {
		return fmt.Errorf("Failed to start the workflow, status code: %d", resp.StatusCode)
	}

	return nil
}

func GetListOfWorkflowRuns(client *github.Client, inputs *action.Inputs, workflowStartedAt time.Time) (*github.WorkflowRuns, error) {
	opts := &github.ListWorkflowRunsOptions{
		Branch:  inputs.Branch,
		Event:   "workflow_dispatch",
		Created: fmt.Sprintf(">=%s", workflowStartedAt.Format(time.RFC3339)),
	}
	runs, resp, err := client.Actions.ListWorkflowRunsByFileName(context.Background(), inputs.Owner, inputs.Repository, inputs.WorkflowFile, opts)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		err = fmt.Errorf("failed to get the workflow, status code %d", resp.StatusCode)
		return nil, err
	}
	if runs.GetTotalCount() < 1 {
		err = fmt.Errorf("failed fo find any workflow runs")
		return nil, err
	}
	return runs, nil
}

func GetLatestWorkflowRunFromList(runs *github.WorkflowRuns) (*github.WorkflowRun, error) {
	var currentWorkflow *github.WorkflowRun
	var timestamp *github.Timestamp
	for i, wf := range runs.WorkflowRuns {
		fmt.Printf("workflow at index %d had completion status of %s at %s with id %d\n", i, *wf.Status, wf.CreatedAt.Time.String(), *wf.ID)
		if timestamp == nil {
			timestamp = wf.CreatedAt
		}
		// we only care about uncompleted workflows
		if *wf.Status != COMPLETED_STATUS {
			if currentWorkflow == nil {
				currentWorkflow = wf
			}
			fmt.Printf("found an active workflow\n")
			// we want the latest workflow so if this one is newer, use it
			if wf.CreatedAt.After(timestamp.Time) {
				fmt.Printf("workflow is newer, using it unless we find a newer one\n")
				currentWorkflow = wf
			}
		}
	}
	if currentWorkflow == nil {
		err := fmt.Errorf("failed to find an unfinished workflow")
		fmt.Printf("%v\n", err)
		return nil, err
	}
	return currentWorkflow, nil
}

// GetMostRecentWorkflowRunId Check the list of current workflows and get the most recent one that is not complete
func GetMostRecentWorkflowRunId(client *github.Client, inputs *action.Inputs, workflowStartedAt time.Time) (int64, error) {
	maxRetrys := 5
	var err error
	var workflow *github.WorkflowRun
	var runs *github.WorkflowRuns
	for i := 0; i < maxRetrys; i++ {
		fmt.Printf("Checking for active workflows attempt: %d\n", i+1)

		// if err is nil then it is the first time and we do not need to sleep, otherwise wait
		if err != nil {
			time.Sleep(inputs.RetryInterval)
		}

		// get the list
		runs, err = GetListOfWorkflowRuns(client, inputs, workflowStartedAt)
		if err != nil || runs == nil {
			continue
		}

		// check the list for the most recent unfinished workflow
		workflow, err = GetLatestWorkflowRunFromList(runs)
		if err != nil || workflow == nil {
			continue
		}

		// if we make it here then we have found the latest unfinished workflow
		break
	}

	// fail if we ended the retry loop in error
	if err != nil || workflow == nil {
		err = fmt.Errorf("did not find any active workflows: %v", err)
		return 0, err
	}

	fmt.Printf("Successfully found a workflow in progress: %d\n", *workflow.ID)
	return *workflow.ID, nil
}

// GetWorkflowRun Gets the workflow run with updated status
func GetWorkflowRun(client *github.Client, inputs *action.Inputs, workflowID int64) (*github.WorkflowRun, error) {
	wfr, resp, err := client.Actions.GetWorkflowRunByID(context.Background(), inputs.Owner, inputs.Repository, workflowID)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Failed to get the workflow, found status code: %d", resp.StatusCode)
	}
	return wfr, nil
}

// PollWorkflow Poll the workflow until the status is complete or we hit the timeout
func PollWorkflow(client *github.Client, inputs *action.Inputs, workflowID int64) *action.Outputs {
	fmt.Printf("Beginning to poll for the workflows status every %v until it is complete or we timeout after %v\n", inputs.RetryInterval, inputs.Timeout)
	pollStart := time.Now()
	stop := false
	var status string
	var latestWorkflow *github.WorkflowRun
	var err error
	testContext, testCancel := context.WithTimeout(context.Background(), inputs.Timeout)
	defer testCancel()
	ticker := time.NewTicker(inputs.RetryInterval)
	for {
		select {
		case <-ticker.C:
			// check the latest status, if it is completed then we stop
			latestWorkflow, err = GetWorkflowRun(client, inputs, workflowID)
			if err != nil {
				fmt.Printf("failed to get the workflow run %v\n", err)
			} else {
				status = *latestWorkflow.Status
				t := time.Now()
				diff := t.Sub(pollStart)
				fmt.Printf("current workflow run status is \"%v\" after %v\n", status, diff)
				stop = status == COMPLETED_STATUS
			}
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
	inputs, err := action.GetInputs()
	if err != nil {
		githubAction.Fatalf("Failed to parse the inputs from the environment %v", err)
	}

	// setup the connection to github
	client := ConnectToGithub(inputs.GithubToken)

	// start the workflow
	workflowStartedAt := time.Now()
	err = StartWorkflow(client, inputs)
	if err != nil {
		githubAction.Fatalf("Failed to start the workflow %v", err)
	}

	// get the workflow we want to poll
	workflowID, err := GetMostRecentWorkflowRunId(client, inputs, workflowStartedAt)
	if err != nil {
		githubAction.Fatalf("did not find any active workflows: %v", err)
	}

	// poll the workflow until it is finished
	outputs := PollWorkflow(client, inputs, workflowID)
	if outputs.Status != COMPLETED_STATUS {
		githubAction.Fatalf("workflow did not reach completed status, id: %d", workflowID)
	}

	// send the outputs back out to the action
	outputs.SetOutputs(githubAction)
}
