package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/tektoncd/experimental/tkn-resolve/pkg/resolve"
	"sigs.k8s.io/yaml"
)

func init() {
	rootCmd.AddCommand(taskRunCmd)
}

var taskRunCmd = &cobra.Command{
	Use:     "taskrun",
	Aliases: []string{"tr"},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(args)
		pr, err := resolve.ResolveTaskRun(args[0])
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
