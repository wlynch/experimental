package resolve

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"sigs.k8s.io/yaml"
)

func ResolvePipelineRun(path string) (*v1beta1.PipelineRun, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	pr := new(v1beta1.PipelineRun)
	if err := yaml.Unmarshal(b, pr); err != nil {
		return nil, err
	}

	// Resolve Pipeline (either by file or by API)
	spec := pr.Spec.PipelineSpec
	if pr.Spec.PipelineRef != nil {
		if strings.HasPrefix(pr.Spec.PipelineRef.Name, "./") || strings.HasPrefix(pr.Spec.PipelineRef.Name, "/") {
			p, err := readPipeline(pr.Spec.PipelineRef.Name)
			if err != nil {
				return nil, err
			}
			spec = &p.Spec
		} else {
			// API resolve
		}
	}

	// Resolve PipelineTasks
	for i, pt := range spec.Tasks {
		if pt.TaskRef == nil {
			// Not a ref - nothing to do.
			continue
		}

		var ts v1beta1.TaskSpec
		if strings.HasPrefix(pt.TaskRef.Name, "./") || strings.HasPrefix(pt.TaskRef.Name, "/") {
			t, err := readTask(pt.TaskRef.Name)
			if err != nil {
				return nil, err
			}
			ts = t.Spec
		} else {
			// API Resolve
		}

		pt.TaskSpec = &v1beta1.EmbeddedTask{TaskSpec: ts}
		pt.TaskRef = nil
		spec.Tasks[i] = pt
	}

	pr.Spec.PipelineSpec = spec
	pr.Spec.PipelineRef = nil
	return pr, nil
}

func readPipeline(path string) (*v1beta1.Pipeline, error) {
	f1, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f1.Close()
	b1, err := ioutil.ReadAll(f1)
	if err != nil {
		return nil, err
	}
	p := new(v1beta1.Pipeline)
	if err := yaml.Unmarshal(b1, p); err != nil {
		return nil, err
	}
	return p, nil
}

func readTask(path string) (*v1beta1.Task, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	t := new(v1beta1.Task)
	if err := yaml.Unmarshal(b, t); err != nil {
		return nil, err
	}
	return t, nil
}
