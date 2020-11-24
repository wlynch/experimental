package common

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"github.com/tektoncd/pipeline/test"
	"gomodules.xyz/jsonpatch/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"knative.dev/pkg/controller"
)

const (
	Path   = "/metadata/annotations/results.tekton.dev~1id"
	IDName = "results.tekton.dev/id"
)

// AnnotationPath creates a jsonpatch path used for adding results_id to Result
// annotations field.
func AnnotationPath(resultID string, path string, op string) ([]byte, error) {
	patches := []jsonpatch.JsonPatchOperation{{
		Operation: op,
		Path:      path,
		Value:     resultID,
	}}
	return json.Marshal(patches)
}

func ReconcileTaskRun(ctx context.Context, asset test.Assets, taskRun *v1beta1.TaskRun) (*v1beta1.TaskRun, error) {
	c := asset.Controller
	clients := asset.Clients
	if err := c.Reconciler.Reconcile(ctx, taskRun.GetNamespacedName().String()); err != nil {
		return nil, err
	}
	tr, err := clients.Pipeline.TektonV1beta1().TaskRuns(taskRun.Namespace).Get(taskRun.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return tr, err
}

func ReconcilePipelineRun(ctx context.Context, asset test.Assets, pipelineRun *v1beta1.PipelineRun) (*v1beta1.PipelineRun, error) {
	c := asset.Controller
	clients := asset.Clients
	if err := c.Reconciler.Reconcile(ctx, pipelineRun.GetNamespacedName().String()); err != nil {
		return nil, err
	}
	pr, err := clients.Pipeline.TektonV1beta1().PipelineRuns(pipelineRun.Namespace).Get(pipelineRun.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return pr, err
}

func Reconcile(ctx context.Context, t *testing.T, ctrl *controller.Impl, name types.NamespacedName) {
	ctrl.EnqueueKey(name)
	if err := ctrl.Reconciler.Reconcile(ctx, name.String()); err != nil {
		t.Fatalf("Failed to reconcile %s: %v", name.String, err)
	}
}
