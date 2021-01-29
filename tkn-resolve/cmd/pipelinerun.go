package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/tektoncd/experimental/tkn-resolve/pkg/resolve"
	"sigs.k8s.io/yaml"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:     "pipelinerun",
	Aliases: []string{"pr"},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(args)
		pr, err := resolve.ResolvePipelineRun(args[0])
		if err != nil {
			log.Fatal(err)
		}
		b, err := yaml.Marshal(pr)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(b))
	},
}
