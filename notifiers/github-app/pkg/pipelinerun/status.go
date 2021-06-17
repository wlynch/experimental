package pipelinerun

import (
	"context"
	"fmt"

	"github.com/google/go-github/v32/github"
	"github.com/tektoncd/experimental/notifiers/github-app/pkg/annotations"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"github.com/tektoncd/pipeline/pkg/pod"
	"knative.dev/pkg/apis"
)

func (r *Reconciler) HandleStatus(ctx context.Context, pr *v1beta1.PipelineRun) error {
	client, err := r.GitHub.NewClient("")
	if err != nil {
		return err
	}

	owner := pr.Annotations[annotations.Owner]
	repo := pr.Annotations[annotations.Repo]
	commit := pr.Annotations[annotations.Commit]

	var description *string
	if m := pr.GetStatusCondition().GetCondition(apis.ConditionSucceeded).GetMessage(); m != "" {
		description = github.String(m)
	}

	status := &github.RepoStatus{
		State:       state(pr.Status),
		Description: description,
		TargetURL:   github.String(dashboardURL(pr)),
		Context:     github.String(pr.GetName()),
	}
	_, _, err = client.Repositories.CreateStatus(ctx, owner, repo, commit, status)
	return err
}

func dashboardURL(tr *v1beta1.PipelineRun) string {
	// TODO: generalize host, object type.
	return fmt.Sprintf("https://dashboard.dogfooding.tekton.dev/#/namespaces/%s/pipelineruns/%s", tr.GetNamespace(), tr.GetName())
}

const (
	StatePending = "pending"
	StateSuccess = "success"
	StateError   = "error"
	StateFailure = "failure"
)

//pending, success, error, or failure.
func state(s v1beta1.PipelineRunStatus) *string {
	c := s.GetCondition(apis.ConditionSucceeded)
	if c == nil {
		return github.String(StatePending)
	}

	switch v1beta1.PipelineRunReason(c.Reason) {
	case pod.ReasonPending, v1beta1.PipelineRunReasonStarted, v1beta1.PipelineRunReasonRunning:
		return github.String(StatePending)
	case v1beta1.PipelineRunReasonSuccessful:
		return github.String(StateSuccess)
	case v1beta1.PipelineRunReasonFailed, v1beta1.PipelineRunReasonCancelled, v1beta1.PipelineRunReasonTimedOut:
		return github.String(StateFailure)
	default:
		return github.String(StatePending)
	}
}
