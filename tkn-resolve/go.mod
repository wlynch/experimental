module github.com/tektoncd/experimental/tkn-resolve

go 1.15

require (
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/spf13/cobra v1.1.1
	github.com/spf13/viper v1.7.1 // indirect
	github.com/tektoncd/pipeline v0.20.1
	sigs.k8s.io/yaml v1.2.0
)

// Pin k8s deps to v0.18.8
replace (
	k8s.io/api => k8s.io/api v0.18.12
	k8s.io/apimachinery => k8s.io/apimachinery v0.18.12
	k8s.io/apiserver => k8s.io/apiserver v0.18.12
	k8s.io/client-go => k8s.io/client-go v0.18.12
)
