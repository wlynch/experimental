/*
Copyright 2020 The Tekton Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package pipelinerun

import (
	"context"

	"github.com/tektoncd/experimental/notifiers/github-app/pkg/annotations"
	"github.com/tektoncd/experimental/notifiers/github-app/pkg/github"
	tektonclient "github.com/tektoncd/pipeline/pkg/client/clientset/versioned/typed/pipeline/v1beta1"
	listers "github.com/tektoncd/pipeline/pkg/client/listers/pipeline/v1beta1"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

// Reconciler updates GitHub CheckRun/Status results for PipelineRun outputs.
type Reconciler struct {
	Logger            *zap.SugaredLogger
	PipelineRunLister listers.PipelineRunLister
	GitHub            *github.GitHubClientFactory
	Kubernetes        kubernetes.Interface
	Tekton            tektonclient.TektonV1beta1Interface
}

// Reconcile creates or updates the check run.
func (r *Reconciler) Reconcile(ctx context.Context, reconcileKey string) error {
	log := r.Logger.With(zap.String("key", reconcileKey))
	log.Infof("reconciling resource")

	namespace, name, err := cache.SplitMetaNamespaceKey(reconcileKey)
	if err != nil {
		log.Errorf("invalid resource key: %s", reconcileKey)
		return nil
	}

	// Get the Task Run resource with this namespace/name
	pr, err := r.PipelineRunLister.PipelineRuns(namespace).Get(name)
	if err != nil {
		log.Errorf("Error retrieving PipelineRun: %v", err)
		return err
	}
	log = log.With(zap.String("uid", string(pr.UID)))

	log.Info("Sending update")

	// If no installation is associated, assume a non-GitHub App status.
	if id := pr.Annotations[annotations.Installation]; id == "" {
		return r.HandleStatus(ctx, pr)
	}
	// Create Check Run with GitHub App
	return r.HandleCheckRun(ctx, log, pr)

}
