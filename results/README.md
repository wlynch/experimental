# Tekton Results API

This package contains experimental code to support a richly queryable API for
Tekton execution history and results.

The full proposal is here:
https://docs.google.com/document/d/1-XBYQ4kBlCHIHSVoYAAf_iC01_by_KoK2aRVO0t8ZQ0/edit

The main components of this design are a **queryable indexed API server** backed
by persistent storage, and an **in-cluster watcher** to report updates to the
API server.

The API server interface is defined in `./proto/api.proto`, and a reference
implementation backed by Sqlite will live in `./cmd/api`. A reference
implementation of the in-cluster watcher will live in `./cmd/watcher`.

## Roadmap

### Q4 2020

- [x] API defined
- [ ] Project repo created

### Q1 2020

Below is a tentative roadmap for upcoming Results work. Feel free to reach out
for any feature requests via an issue!

- [ ] Result API v0.1.0
  - [ ] [v1alpha2](https://github.com/tektoncd/community/blob/725d33a2bd4e0126c55c8c2fe6cabe90647add05/teps/0021-results-api.md) Result/Record CRUD
  - [ ] Basic CEL filtering
  - [ ] Pagination
- [ ] Result Watcher v0.1.0
  - [ ] TaskRun/PipelineRun result uploading

### Q2 2020

- [ ] Result API
  - [ ] Auth
  - [ ] Trigger Events
  - [ ] Notifications
- [ ] Result Watcher
  - [ ] Task/PipelineRun Cleanup
  - [ ] Trigger Events
  - [ ] Notifications

## Development

### Configure your database.

The reference implementation of the API Server requires a SQL database for
result storage. The database schema can be found under
[schema/results.sql](schema/results.sql).

Initial one-time setup is required to configure the password and initial config:

```sh
kubectl create secret generic tekton-results-mysql --namespace="tekton-pipelines" --from-literal=user=root --from-literal=password=$(openssl rand -base64 20)
kubectl create configmap mysql-initdb-config --from-file="schema/results.sql" --namespace="tekton-pipelines"
```

### Deploying

To build and deploy both components, use
[`ko`](https://github.com/GoogleCloudPlatform/ko). Make sure you have a valid
kubeconfig, and have set the `KO_DOCKER_REPO` env var.

```
ko apply -f config/
```

To only build and deploy one component:

```
ko apply -f config/watcher.yaml
```

### Regenerating protobuf-generated code

1. Install protobuf compiler

e.g., for macOS:

```
brew install protobuf
```

2. Install the protoc Go plugin

```
$ go get -u github.com/golang/protobuf/protoc-gen-go
```

3. Rebuild the generated Go code

```
$ go generate ./proto/
```
