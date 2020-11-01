/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

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
	"context"
	"github.com/Octops/agones-discover-openmatch/internal/runtime"
	"github.com/Octops/agones-discover-openmatch/pkg/config"
	"github.com/Octops/agones-discover-openmatch/pkg/matchfunction"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// functionCmd represents the function command
var functionCmd = &cobra.Command{
	Use:   "function",
	Short: "Start the Match Function Server",
	Long: `The Match Function is the component that implements the core matchmaking logic. 
A Match Function receives a MatchProfile as input should return matches for this MatchProfile.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger := runtime.NewLogger(verbose)
		mmfServer, err := matchfunction.NewServer()
		if err != nil {
			logger.Fatal(errors.Wrap(err, "failed to create match function server"))
		}

		ctx, cancel := context.WithCancel(context.Background())
		runtime.SetupSignal(cancel)

		if err := mmfServer.Serve(ctx, config.OpenMatch().MatchFunctionPort); err != nil {
			logger.Fatal(errors.Wrap(err, "failed to start match function server"))
		}
	},
}

func init() {
	rootCmd.AddCommand(functionCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// functionCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// functionCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
