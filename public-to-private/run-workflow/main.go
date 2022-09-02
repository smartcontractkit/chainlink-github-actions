package main

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/google/go-github/v47/github"
	"github.com/kelseyhightower/envconfig"
	"github.com/sethvargo/go-githubactions"
	"golang.org/x/oauth2"
)

const COMPLETED_STATUS = "completed"

type Inputs struct {
	GithubToken   string        `envconfig:"token" required:"true"`
	Owner         string        `envconfig:"owner" default:"smartcontractkit"`
	Repository    string        `envconfig:"repository" required:"true"`
	Branch        string        `envconfig:"ref" required:"true"`
	WorkflowFile  string        `envconfig:"workflow_file" required:"true"`
	Timeout       time.Duration `envconfig:"timeout" default:"5m"`
	RetryInterval time.Duration `envconfig:"retry_interval" default:"15s"`
	Inputs        string        `envconfig:"inputs" default:"{}"`
}

func (i *Inputs) inputsToMap(c *Common) map[string]interface{} {
	inputsMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(i.Inputs), &inputsMap)
	if err != nil {
		c.action.Fatalf("could not unmarshall inputs json: %v", err)
	}
	return inputsMap
}

// getInputs Loads inputs from the environment
func getInputs(c *Common) *Inputs {
	var inputs Inputs
	if err := envconfig.Process("", &inputs); err != nil {
		c.action.Fatalf("could not load inputs from environment: %v", err)
	}

	return &inputs
}

type Outputs struct {
	Status     string
	Conclusion string
	WorkflowID int64
}

// setOutputs Sets the outputs in the format that a docker action can parse
func (o *Outputs) setOutputs(c *Common) {
	ao := *o
	val := reflect.ValueOf(ao)
	typeOfS := val.Type()

	for i := 0; i < val.NumField(); i++ {
		k := strings.ToLower(typeOfS.Field(i).Name)
		v := fmt.Sprintf("%v", val.Field(i).Interface())
		c.action.Infof("Setting output: %s = %s", k, v)
		c.action.SetOutput(k, v)
	}
}

type Common struct {
	ctx               context.Context
	client            *github.Client
	action            *githubactions.Action
	workflowStartedAt time.Time
	workflowID        int64
}

// connectToGithub Sets up the github client and connects
func (c *Common) connectToGithub(inputs *Inputs) {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: inputs.GithubToken},
	)
	tc := oauth2.NewClient(c.ctx, ts)
	c.client = github.NewClient(tc)
}

func (c *Common) startWorkflow(inputs *Inputs) {
	c.action.Infof("Starting workflow %s", inputs.WorkflowFile)
	event := &github.CreateWorkflowDispatchEventRequest{
		Ref:    inputs.Branch,
		Inputs: inputs.inputsToMap(c),
	}
	resp, err := c.client.Actions.CreateWorkflowDispatchEventByFileName(c.ctx, inputs.Owner, inputs.Repository, inputs.WorkflowFile, *event)

	if err != nil {
		c.action.Fatalf("Failed to make call to start workflow: %v", err)
	}
	if resp.StatusCode != 204 {
		c.action.Fatalf("Failed to start the workflow: %v", err)
	}
}

func (c *Common) getListOfWorkflowRuns(inputs *Inputs) (*github.WorkflowRuns, error) {
	opts := &github.ListWorkflowRunsOptions{
		Branch:  inputs.Branch,
		Event:   "workflow_dispatch",
		Created: fmt.Sprintf(">=%s", c.workflowStartedAt.Format(time.RFC3339)),
	}
	runs, resp, err := c.client.Actions.ListWorkflowRunsByFileName(c.ctx, inputs.Owner, inputs.Repository, inputs.WorkflowFile, opts)
	if err != nil {
		c.action.Infof("failed to get the workflow: %v", err)
		return nil, err
	}
	if resp.StatusCode != 200 {
		err = fmt.Errorf("failed to get the workflow, status code %d", resp.StatusCode)
		c.action.Infof("%v", err)
		return nil, err
	}
	if runs.GetTotalCount() < 1 {
		err = fmt.Errorf("failed fo find any workflow runs")
		c.action.Infof("%v", err)
		return nil, err
	}
	return runs, nil
}

func (c *Common) getLatestWorkflowRunFromList(runs *github.WorkflowRuns) (*github.WorkflowRun, error) {
	var currentWorkflow *github.WorkflowRun
	var timestamp *github.Timestamp
	for i, wf := range runs.WorkflowRuns {
		c.action.Infof("workflow at index %d had completion status of %s at %s with id %d", i, *wf.Status, wf.CreatedAt.Time.String(), *wf.ID)
		if timestamp == nil {
			timestamp = wf.CreatedAt
		}
		// we only care about uncompleted workflows
		if *wf.Status != COMPLETED_STATUS {
			if currentWorkflow == nil {
				currentWorkflow = wf
			}
			c.action.Infof("found an active workflow")
			// we want the latest workflow so if this one is newer, use it
			if wf.CreatedAt.After(timestamp.Time) {
				c.action.Infof("workflow is newer, using it unless we find a newer one")
				currentWorkflow = wf
			}
		}
	}
	if currentWorkflow == nil {
		err := fmt.Errorf("failed to find an unfinished workflow")
		c.action.Infof("%v", err)
		return nil, err
	}
	return currentWorkflow, nil
}

// getMostRecentWorkflowRunId Check the list of current workflows and get the most recent one that is not complete
func (c *Common) getMostRecentWorkflowRunId(inputs *Inputs) {
	maxRetrys := 5
	var err error
	var workflow *github.WorkflowRun
	var runs *github.WorkflowRuns
	for i := 0; i < maxRetrys; i++ {
		c.action.Infof("Checking for active workflows attempt: %d", i+1)

		// if err is nil then it is the first time and we do not need to sleep, otherwise wait
		if err != nil {
			time.Sleep(inputs.RetryInterval)
		}

		// get the list
		runs, err = c.getListOfWorkflowRuns(inputs)
		if err != nil || runs == nil {
			continue
		}

		// check the list for the most recent unfinished workflow
		workflow, err = c.getLatestWorkflowRunFromList(runs)
		if err != nil || workflow == nil {
			continue
		}

		// if we make it here then we have found the latest unfinished workflow
		break
	}

	// fail if we ended the retry loop in error
	if err != nil || workflow == nil {
		c.action.Fatalf("did not find any active workflows: %v", err)
	}

	c.action.Infof("Successfully found a workflow in progress: %d", *workflow.ID)
	c.workflowID = *workflow.ID
}

// getWorkflowRun Gets the workflow run with updated status
func (c *Common) getWorkflowRun(inputs *Inputs) *github.WorkflowRun {
	wfr, resp, err := c.client.Actions.GetWorkflowRunByID(c.ctx, inputs.Owner, inputs.Repository, c.workflowID)
	if err != nil {
		c.action.Infof("Failed to get the workflow run: %v", err)
		return nil
	}
	if resp.StatusCode != 200 {
		c.action.Infof("Failed to get the workflow, found status code: %d", resp.StatusCode)
		return nil
	}
	return wfr
}

// pollWorkflow Poll the workflow until the status is complete or we hit the timeout
func (c *Common) pollWorkflow(inputs *Inputs) *Outputs {
	c.action.Infof("Beginning to poll for the workflows status every %v until it is complete or we timeout after %v", inputs.RetryInterval, inputs.Timeout)
	pollStart := time.Now()
	stop := false
	var status string
	var latestWorkflow *github.WorkflowRun
	testContext, testCancel := context.WithTimeout(c.ctx, inputs.Timeout)
	defer testCancel()
	ticker := time.NewTicker(inputs.RetryInterval)
	for {
		select {
		case <-ticker.C:
			// check the latest status, if it is completed then we stop
			latestWorkflow = c.getWorkflowRun(inputs)
			status = *latestWorkflow.Status
			t := time.Now()
			diff := t.Sub(pollStart)
			c.action.Infof("current workflow run status is \"%v\" after %v", status, diff)
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
		c.action.Fatalf("workflow did not reach completed status, id: %d", c.workflowID)
	}

	return &Outputs{
		Status:     status,
		Conclusion: *latestWorkflow.Conclusion,
		WorkflowID: c.workflowID,
	}
}

func main() {

	// setup our common objects
	common := &Common{
		ctx:    context.Background(),
		action: githubactions.New(),
	}

	// get the inputs from the action
	inputs := getInputs(common)

	// setup the connection to github
	common.connectToGithub(inputs)

	// start the workflow
	common.workflowStartedAt = time.Now()
	common.startWorkflow(inputs)

	// get the workflow we want to poll
	common.getMostRecentWorkflowRunId(inputs)

	// poll the workflow until it is finished
	outputs := common.pollWorkflow(inputs)

	// send the outputs back out to the action
	outputs.setOutputs(common)
}
