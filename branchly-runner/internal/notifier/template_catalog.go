package notifier

import (
	"embed"
	"fmt"
	"strconv"
	"strings"
)

const (
	TemplateSlugJobCompleted = "branchly-job-completed"
	TemplateSlugJobFailed    = "branchly-job-failed"
	TemplateSlugPROpened     = "branchly-pr-opened"
)

type TemplateDefinition struct {
	Slug         string
	Name         string
	HTMLTemplate string
}

//go:embed templates/*.html
var templateFS embed.FS

func templateDefinitions() ([]TemplateDefinition, error) {
	completed, err := readTemplate("templates/" + TemplateSlugJobCompleted + ".html")
	if err != nil {
		return nil, err
	}
	failed, err := readTemplate("templates/" + TemplateSlugJobFailed + ".html")
	if err != nil {
		return nil, err
	}
	prOpened, err := readTemplate("templates/" + TemplateSlugPROpened + ".html")
	if err != nil {
		return nil, err
	}

	return []TemplateDefinition{
		{
			Slug:         TemplateSlugJobCompleted,
			Name:         "Branchly - Job Completed",
			HTMLTemplate: completed,
		},
		{
			Slug:         TemplateSlugJobFailed,
			Name:         "Branchly - Job Failed",
			HTMLTemplate: failed,
		},
		{
			Slug:         TemplateSlugPROpened,
			Name:         "Branchly - Pull Request Opened",
			HTMLTemplate: prOpened,
		},
	}, nil
}

func readTemplate(path string) (string, error) {
	b, err := templateFS.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read template %s: %w", path, err)
	}
	return string(b), nil
}

func eventSubjectAndSlug(event string, data JobNotifData) (string, string, bool) {
	switch event {
	case "job_completed":
		return "Job completed on " + data.RepoFullName, TemplateSlugJobCompleted, true
	case "job_failed":
		return "Job failed on " + data.RepoFullName, TemplateSlugJobFailed, true
	case "pr_opened":
		return "Pull request opened on " + data.RepoFullName, TemplateSlugPROpened, true
	default:
		return "", "", false
	}
}

func templateVars(data JobNotifData) map[string]string {
	return map[string]string{
		"repo_full_name": data.RepoFullName,
		"branch_name":    data.BranchName,
		"agent_name":     data.AgentName,
		"prompt":         data.Prompt,
		"duration":       formatDuration(data.DurationSeconds),
		"estimated_cost": formatEstimatedCost(data.EstimatedCostUSD),
		"pr_url":         strings.TrimSpace(data.PRUrl),
		"job_logs_url":   strings.TrimSpace(data.JobLogsUrl),
		"finished_at":    data.FinishedAt.UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT"),
		"error_message":  data.ErrorMessage,
	}
}

func formatDuration(seconds float64) string {
	total := int(seconds + 0.5)
	if total < 60 {
		return strconv.Itoa(total) + "s"
	}
	minutes := total / 60
	remain := total % 60
	if remain == 0 {
		return strconv.Itoa(minutes) + "m"
	}
	return strconv.Itoa(minutes) + "m " + strconv.Itoa(remain) + "s"
}

func formatEstimatedCost(cost *float64) string {
	if cost == nil {
		return ""
	}
	return fmt.Sprintf("$%.4f", *cost)
}
