package taskrun

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/tektoncd/experimental/results/pkg/watcher/reconciler/common"
	"github.com/tektoncd/experimental/results/pkg/watcher/reconciler/internal"
	pb "github.com/tektoncd/experimental/results/proto/proto"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	faketekton "github.com/tektoncd/pipeline/pkg/client/clientset/versioned/fake"
	informers "github.com/tektoncd/pipeline/pkg/client/informers/externalversions"
	tektoninject "github.com/tektoncd/pipeline/pkg/client/injection/client"
	taskruninformer "github.com/tektoncd/pipeline/pkg/client/injection/informers/pipeline/v1beta1/taskrun"
	"github.com/tektoncd/pipeline/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"knative.dev/pkg/controller"
)

/*func TestReconcile(t *testing.T) {
	taskRunTest := NewTaskRunTest(t)

	testFuncs := map[string]func(t *testing.T){
		//"Create": taskRunTest.testCreateTaskRun,
		//"Unchange": taskRunTest.testUnchangeTaskRun,
		//"Update": taskRunTest.testUpdateTaskRun,
	}

	for name, testFunc := range testFuncs {
		t.Run(name, testFunc)
	}
}
*/

var (
	taskRun = &v1beta1.TaskRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "Tekton-TaskRun",
			Namespace:   "default",
			Annotations: map[string]string{"demo": "demo"},
			UID:         "12345",
		},
	}
)

type TaskRunTest struct {
	taskRun *v1beta1.TaskRun
	asset   test.Assets

	ctx     context.Context
	ctrl    *controller.Impl
	results pb.ResultsClient
	tekton  *faketekton.Clientset
}

func NewTaskRunTest(t *testing.T) TaskRunTest {
	ctx := context.Background()
	tekton := faketekton.NewSimpleClientset(taskRun)
	informer := informers.NewSharedInformerFactory(tekton, 0)
	informer.Tekton().V1beta1().TaskRuns().Informer().GetIndexer().Add(taskRun)
	ctx = context.WithValue(ctx, taskruninformer.Key{}, informer.Tekton().V1beta1().TaskRuns())
	ctx = context.WithValue(ctx, tektoninject.Key{}, tekton)
	results := internal.NewResultsClient(t)

	return TaskRunTest{
		taskRun: taskRun,
		ctx:     ctx,
		results: results,
		tekton:  tekton,
		ctrl:    NewController(ctx, nil, results),
	}
}

func newController(t *testing.T, objs ...runtime.Object) (context.Context, *controller.Impl, *faketekton.Clientset, pb.ResultsClient) {
	ctx := context.Background()
	tekton := faketekton.NewSimpleClientset(taskRun)
	informer := informers.NewSharedInformerFactory(tekton, 0)
	for _, o := range objs {
		informer.Tekton().V1beta1().TaskRuns().Informer().GetIndexer().Add(o)
	}
	ctx = context.WithValue(ctx, taskruninformer.Key{}, informer.Tekton().V1beta1().TaskRuns())
	ctx = context.WithValue(ctx, tektoninject.Key{}, tekton)
	results := internal.NewResultsClient(t)
	ctrl := NewController(ctx, nil, results)

	return ctx, ctrl, tekton, results
}

func taskrun() *v1beta1.TaskRun {
	return &v1beta1.TaskRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "Tekton-TaskRun",
			Namespace:   "default",
			Annotations: map[string]string{"demo": "demo"},
			UID:         "12345",
		},
	}
}

func reconcile(ctx context.Context, t *testing.T, ctrl *controller.Impl, tr *v1beta1.TaskRun) *v1beta1.TaskRun {
	common.Reconcile(ctx, t, ctrl, taskRun.GetNamespacedName())
	
	tekton := tektoninject.Get(ctx)
	tr, err := tekton.TektonV1beta1().TaskRuns(taskRun.Namespace).Get(taskRun.Name, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get completed TaskRun %s: %v", taskRun.Name, err)
	}
}

func TestCreateTaskRun(t *testing.T) {
	ctx, ctrl, tekton, results := newController(t, taskRun)
	common.Reconcile(ctx, t, ctrl, taskRun.GetNamespacedName())

	tr, err := tekton.TektonV1beta1().TaskRuns(taskRun.Namespace).Get(taskRun.Name, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get completed TaskRun %s: %v", taskRun.Name, err)
	}

	if _, ok := tr.Annotations[common.IDName]; !ok {
		t.Fatalf("Expected completed TaskRun %s should be updated with a results_id field in annotations", tr.Name)
	}
	if _, err := results.GetResult(ctx, &pb.GetResultRequest{Name: tr.Annotations[common.IDName]}); err != nil {
		t.Fatalf("Expected completed TaskRun %s not created in api server", tr.Name)
	}
}

func TestUnchangeTaskRun(t *testing.T) {
	ctx, ctrl, tekton, _ := newController(t, taskRun)

	common.Reconcile(ctx, t, ctrl, taskRun.GetNamespacedName())
	tr1, err := tekton.TektonV1beta1().TaskRuns(taskRun.Namespace).Get(taskRun.Name, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get completed TaskRun %s: %v", taskRun.Name, err)
	}

	common.Reconcile(ctx, t, ctrl, taskRun.GetNamespacedName())
	tr2, err := tekton.TektonV1beta1().TaskRuns(taskRun.Namespace).Get(taskRun.Name, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get completed TaskRun %s: %v", taskRun.Name, err)
	}

	if diff := cmp.Diff(tr1, tr2); diff != "" {
		t.Fatalf("Expected completed TaskRun should remain unchanged when it has a results_id in annotations: %v", diff)
	}
}

func TestUpdateTaskRun(t *testing.T) {
	ctx, ctrl, tekton, _ := newController(t, taskRun)

	tr, err := common.ReconcileTaskRun(tt.ctx, tt.asset, tt.taskRun)
	if err != nil {
		t.Fatalf("Failed to get completed TaskRun %s: %v", tt.taskRun.Name, err)
	}
	tr.UID = "234435"
	if _, err := tt.asset.Clients.Pipeline.TektonV1beta1().TaskRuns(tt.taskRun.Namespace).Update(tr); err != nil {
		t.Fatalf("Failed to update TaskRun %s to Tekton Pipeline Client: %v", tt.taskRun.Name, err)
	}
	updatetr, err := common.ReconcileTaskRun(tt.ctx, tt.asset, tr)
	if err != nil {
		t.Fatalf("Failed to reconcile TaskRun %s: %v", tt.taskRun.Name, err)
	}
	updatetr.ResourceVersion = tr.ResourceVersion
	if diff := cmp.Diff(tr, updatetr); diff != "" {
		t.Fatalf("Expected completed TaskRun should be updated in cluster: %v", diff)
	}
	res, err := tt.client.GetResult(tt.ctx, &pb.GetResultRequest{Name: tr.Annotations[common.IDName]})
	if err != nil {
		t.Fatalf("Expected completed TaskRun %s not created in api server", tt.taskRun.Name)
	}
	p, err := convert.ToTaskRunProto(updatetr)
	if err != nil {
		t.Fatalf("failed to convert to proto: %v", err)
	}
	want := &pb.Result{
		Name: tr.Annotations[common.IDName],
		Executions: []*pb.Execution{{
			Execution: &pb.Execution_TaskRun{p},
		}},
	}
	if diff := cmp.Diff(want, res, protocmp.Transform()); diff != "" {
		t.Fatalf("Expected completed TaskRun should be upated in api server: %v", diff)

	}
}
*/
