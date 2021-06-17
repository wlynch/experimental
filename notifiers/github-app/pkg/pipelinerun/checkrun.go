// Copyright 2020 The Tekton Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pipelinerun

import (
	"context"
	"fmt"
	"strconv"

	"github.com/google/go-github/v32/github"
	"github.com/tektoncd/experimental/notifiers/github-app/pkg/annotations"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
)

func (r *Reconciler) HandleCheckRun(ctx context.Context, log *zap.SugaredLogger, pr *v1beta1.PipelineRun) error {
	gh, err := r.GitHub.NewClient(pr.Annotations[annotations.Installation])
	if err != nil {
		return err
	}

	// Render out GitHub CheckRun output.
	body, err := render(pr)
	if err != nil {
		log.Errorf("error rendering TaskRun: %v", err)
		return err
	}

	// Update or create the CheckRun.
	cr, err := UpsertCheckRun(ctx, gh, pr, &github.CheckRunOutput{
		Title:   github.String(pr.Name),
		Summary: github.String(body),
		//Text:    github.String(logs),
	})
	if err != nil {
		log.Errorf("UpsertCheckRun: %v", err)
		return err
	}

	// Update TaskRun with CheckRun ID so that we can determine if there's an
	// existing CheckRun for the TaskRun in future updates.
	// TODO: Prevent a 2nd round of reconciliation for this annotation update?
	if id := strconv.FormatInt(cr.GetID(), 10); id != pr.Annotations[annotations.CheckRun] {
		pr.Annotations[annotations.CheckRun] = id
		if _, err := r.Tekton.PipelineRuns(pr.GetNamespace()).Update(pr); err != nil {
			log.Errorf("TaskRun.Update: %v", err)
			return err
		}
	}
	return nil
}

// UpsertCheckRun updates or creates a check run for the given TaskRun.
func UpsertCheckRun(ctx context.Context, client *github.Client, pr *v1beta1.PipelineRun, output *github.CheckRunOutput) (*github.CheckRun, error) {
	owner := pr.Annotations[annotations.Owner]
	repo := pr.Annotations[annotations.Repo]
	commit := pr.Annotations[annotations.Commit]

	status, conclusion := status(pr.Status)

	if id, ok := pr.Annotations[annotations.CheckRun]; ok {
		// A check run was already associated to the TaskRun - update.
		n, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("error converting check run id: %v", err)
		}
		cr, _, err := client.Checks.UpdateCheckRun(ctx, owner, repo, n, github.UpdateCheckRunOptions{
			ExternalID:  github.String(pr.GetSelfLink()),
			Name:        pr.Name,
			Status:      github.String(status),
			Conclusion:  github.String(conclusion),
			HeadSHA:     github.String(commit),
			Output:      output,
			CompletedAt: ghtime(pr.Status.CompletionTime),
			// TODO: Replace with Task-specific URL
			DetailsURL: github.String("https://tekton.dev"),
		})
		if err != nil {
			return nil, fmt.Errorf("CreateCheck: %w", err)
		}
		return cr, nil
	}

	// There's no existing CheckRun - create.
	cr, _, err := client.Checks.CreateCheckRun(ctx, owner, repo, github.CreateCheckRunOptions{
		ExternalID:  github.String(pr.GetSelfLink()),
		Name:        pr.GetNamespacedName().String(),
		Status:      github.String(status),
		Conclusion:  github.String(conclusion),
		HeadSHA:     commit,
		Output:      output,
		StartedAt:   ghtime(pr.Status.StartTime),
		CompletedAt: ghtime(pr.Status.CompletionTime),
		// TODO: Replace with Task-specific URL
		DetailsURL: github.String("https://tekton.dev"),
	})
	if err != nil {
		return nil, fmt.Errorf("CreateCheck: %w", err)
	}
	return cr, nil
}

const (
	CheckRunStatusQueued     = "queued"
	CheckRunStatusInProgress = "in_progress"
	CheckRunStatusCompleted  = "completed"

	CheckRunConclusionSuccess        = "success"
	CheckRunConclusionFailure        = "failure"
	CheckRunConclusionCancelled      = "cancelled"
	CheckRunConclusionTimeout        = "timed_out"
	CheckRunConclusionActionRequired = "action_required"
)

func status(s v1beta1.PipelineRunStatus) (status, conclusion string) {
	c := s.GetCondition(apis.ConditionSucceeded)
	if c == nil {
		return "", ""
	}

	switch c.Reason {
	case "Pending":
		return CheckRunStatusQueued, ""
	case "Started", "Running":
		return CheckRunStatusInProgress, ""
	case "Succeeded":
		return CheckRunStatusCompleted, CheckRunConclusionSuccess
	case "Failed":
		return CheckRunStatusCompleted, CheckRunConclusionFailure
	case "TaskRunCancelled":
		return CheckRunStatusCompleted, CheckRunConclusionCancelled
	case "TaskRunTimeout":
		return CheckRunStatusCompleted, CheckRunConclusionTimeout
	}

	if c.Status == v1.ConditionFalse {
		if s.CompletionTime == nil {
			return CheckRunStatusInProgress, ""
		}
		return CheckRunStatusCompleted, CheckRunConclusionActionRequired
	}

	return "", ""
}

func ghtime(t *metav1.Time) *github.Timestamp {
	if t == nil {
		return nil
	}
	return &github.Timestamp{Time: t.Time}
}
