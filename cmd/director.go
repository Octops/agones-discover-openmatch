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
	"github.com/Octops/agones-discover-openmatch/pkg/allocator"
	"github.com/Octops/agones-discover-openmatch/pkg/director/openmatch"
	"github.com/pkg/errors"

	"github.com/spf13/cobra"
)

var (
	intervalDirector  string
	agonesDiscoverURL string
)

// directorCmd represents the director command
var directorCmd = &cobra.Command{
	Use:   "director",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger := runtime.NewLogger(verbose).WithField("component", "director")
		ctx, cancel := context.WithCancel(context.Background())
		runtime.SetupSignal(cancel)

		logger.Info("starting OpenMatch Director")
		// TODO: Refactor using Flags and Registry
		client, err := allocator.NewAgonesDiscoverClientHTTP(agonesDiscoverURL)
		if err != nil {
			logger.Fatal(errors.Wrap(err, "failed to creating Agones Discover Client"))
		}

		agonesAllocator := allocator.NewAllocatorService(&allocator.AgonesDiscoverAllocator{
			Client: client,
		})
		if err := openmatch.RunDirector(ctx, logger, openmatch.ConnFuncInsecure, intervalDirector, agonesAllocator); err != nil {
			logger.Fatal(errors.Wrap(err, "failed to start the Director"))
		}
	},
}

func init() {
	rootCmd.AddCommand(directorCmd)

	directorCmd.Flags().StringVar(&intervalDirector, "interval", "5s", "interval the Director will fetch matches")
	directorCmd.Flags().StringVar(&agonesDiscoverURL, "agones-discover-url", "http://localhost:8081", "the Agones Discover server URL")
}
