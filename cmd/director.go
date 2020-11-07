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
	octopsDiscoverURL string
)

// directorCmd represents the director command
var directorCmd = &cobra.Command{
	Use:   "director",
	Short: "The Director fetches Matches from Open Match for a set of MatchProfiles",
	Long: `The Director fetches Matches from Open Match for a set of MatchProfiles.
For these matches, it fetches Game Server from a DGS allocation system. 
It can then communicate the Assignments back to the Game Frontend.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger := runtime.NewLogger(verbose).WithField("component", "director")
		ctx, cancel := context.WithCancel(context.Background())
		runtime.SetupSignal(cancel)

		logger.Info("starting OpenMatch Director")
		agonesAllocator, err := BuildAgonesAllocator()
		if err != nil {
			logger.Fatal(err)
		}

		if err := openmatch.RunDirector(ctx, logger, openmatch.ConnFuncInsecure, intervalDirector, agonesAllocator); err != nil {
			logger.Fatal(errors.Wrap(err, "failed to start the Director"))
		}
	},
}

func BuildAgonesAllocator() (*allocator.AllocatorService, error) {

	// TODO: Refactor using Flags and Registry
	client, err := allocator.NewAgonesDiscoverClientHTTP(octopsDiscoverURL)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create AgonesDiscoverClientHTTP")
	}

	return allocator.NewAllocatorService(&allocator.AgonesDiscoverAllocator{
		Client: client,
	}), nil
}

func init() {
	rootCmd.AddCommand(directorCmd)

	directorCmd.Flags().StringVar(&intervalDirector, "interval", "5s", "interval the Director will fetch matches")
	directorCmd.Flags().StringVar(&octopsDiscoverURL, "octops-discover-url", "http://localhost:8081", "the Octops Discover server URL")
}
