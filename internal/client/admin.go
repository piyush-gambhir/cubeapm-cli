package client

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/piyush-gambhir/cubeapm-cli/internal/types"
)

const logsDeleteBasePath = "/api/logs/delete"

// DeleteLogsRun starts a log deletion task.
func (c *Client) DeleteLogsRun(filter string) (string, error) {
	params := url.Values{}
	params.Set("filter", filter)

	resp, err := c.post(c.adminBaseURL, logsDeleteBasePath+"/run_task", params)
	if err != nil {
		return "", fmt.Errorf("starting log deletion: %w", err)
	}
	defer resp.Body.Close()

	if err := c.checkResponse(resp); err != nil {
		return "", err
	}

	var result types.DeleteRunResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("parsing delete response: %w", err)
	}

	return result.TaskID, nil
}

// DeleteLogsStop stops a log deletion task.
func (c *Client) DeleteLogsStop(taskID string) error {
	params := url.Values{}
	params.Set("task_id", taskID)

	resp, err := c.post(c.adminBaseURL, logsDeleteBasePath+"/stop_task", params)
	if err != nil {
		return fmt.Errorf("stopping log deletion: %w", err)
	}
	defer resp.Body.Close()

	return c.checkResponse(resp)
}

// DeleteLogsList lists active log deletion tasks.
func (c *Client) DeleteLogsList() ([]types.DeleteTask, error) {
	resp, err := c.get(c.adminBaseURL, logsDeleteBasePath+"/active_tasks", nil)
	if err != nil {
		return nil, fmt.Errorf("listing delete tasks: %w", err)
	}
	defer resp.Body.Close()

	if err := c.checkResponse(resp); err != nil {
		return nil, err
	}

	var tasks []types.DeleteTask
	if err := json.NewDecoder(resp.Body).Decode(&tasks); err != nil {
		return nil, fmt.Errorf("parsing delete tasks: %w", err)
	}

	return tasks, nil
}
