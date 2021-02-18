/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

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
package cmd

import (
	"github.com/onbeep/devenv/pkg/devenv"
	"log"

	"github.com/spf13/cobra"
)

// glassCmd represents the glass command
var glassCmd = &cobra.Command{
	Use:   "glass",
	Short: "Nuke and pave an environment (destroy, then recreate).",
	Long: `
Nuke and pave an environment (destroy, then recreate).

`,
	Run: func(cmd *cobra.Command, args []string) {
		if name == "" {
			if len(args) > 0 {
				name = args[0]
			}
		}

		err := devenv.Glass(name)
		if err != nil {
			log.Fatalf("Error running glass: %s", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(glassCmd)

}
