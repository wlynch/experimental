# tkn-resolve

tkn plugin for client-side resolving of Tekton resources.

This tool allows you to specify local file refs within a TaskRun or PipelineRun. Local file refs are indicated by prefixing `./` (relative) or `/` (absolute) to the ref name. Relative paths are relative with respect to the working directory when the tool is executed. Local file refs are rendered as embedded Tasks within the respective TaskRun or PipelineRun.

For example, given the following Task `task.yaml`:

```yaml
apiVersion: tekton.dev/v1beta1
kind: Task
metadata:
  name: echo-hello-world
spec:
  steps:
    - name: echo
      image: ubuntu
      command:
        - echo
      args:
        - "Hello World"
```

and TaskRun `taskrun.yaml`:

```yaml
apiVersion: tekton.dev/v1beta1
kind: TaskRun
metadata:
  name: echo-hello-world-task-run
spec:
  taskRef:
    name: ./task.yaml
```

executing `tkn resolve taskrun ./taskrun.yaml` would result in the following output:

```yaml
apiVersion: tekton.dev/v1beta1
kind: TaskRun
metadata:
  creationTimestamp: null
  name: echo-hello-world-task-run
spec:
  serviceAccountName: ""
status:
  podName: ""
  taskSpec:
    steps:
    - args:
      - Hello World
      command:
      - echo
      image: ubuntu
      name: echo
      resources: {}
```

## Installation

```sh
$ GOBIN=${TKN_PLUGINS_DIR:-"${HOME}/.config/tkn/plugins"} go install github.com/tektoncd/experimental/tkn-resolve
```

## Usage with kubectl

```sh
$ tkn resolve pipelinerun run.yaml | kubectl apply -f -
```