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

type AgonesAllocatorArgs struct {
	KeyFile              string
	CertFile             string
	CaCertFile           string
	AllocatorServiceHost string
	AllocatorServicePort int
	Namespace            string
	MultiCluster         bool
}

type OctopsDiscoverArgs struct {
	DiscoverServiceURL string
}

var (
	intervalDirector    string
	allocatorMode       string
	agonesAllocatorArgs = &AgonesAllocatorArgs{}
	octopsDiscoverArgs  = &OctopsDiscoverArgs{}
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
		agonesAllocator, err := BuildAgonesAllocatorService(allocatorMode)
		if err != nil {
			logger.Fatal(err)
		}

		if err := openmatch.RunDirector(ctx, logger, openmatch.ConnFuncInsecure, intervalDirector, agonesAllocator); err != nil {
			logger.Fatal(errors.Wrap(err, "failed to start the Director"))
		}
	},
}

func BuildAgonesAllocatorService(mode string) (*allocator.AllocatorService, error) {
	var allocatorSvc *allocator.AllocatorService
	switch mode {
	case "agones":
		config := &allocator.AgonesAllocatorClientConfig{
			KeyFile:              agonesAllocatorArgs.KeyFile,
			CertFile:             agonesAllocatorArgs.CertFile,
			CaCertFile:           agonesAllocatorArgs.CaCertFile,
			AllocatorServiceHost: agonesAllocatorArgs.AllocatorServiceHost,
			AllocatorServicePort: agonesAllocatorArgs.AllocatorServicePort,
			Namespace:            agonesAllocatorArgs.Namespace,
			MultiCluster:         agonesAllocatorArgs.MultiCluster,
		}

		client, err := allocator.NewAgonesAllocatorClient(config)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create AgonesDiscoverClientHTTP")
		}
		allocatorSvc = allocator.NewAllocatorService(&allocator.AgonesAllocator{
			Client: client,
		})
	default: //"discover"
		// TODO: Refactor using Flags and Registry
		client, err := allocator.NewAgonesDiscoverClientHTTP(octopsDiscoverArgs.DiscoverServiceURL)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create AgonesDiscoverClientHTTP")
		}
		allocatorSvc = allocator.NewAllocatorService(&allocator.AgonesDiscoverAllocator{
			Client: client,
		})
	}

	return allocatorSvc, nil
}

func init() {
	rootCmd.AddCommand(directorCmd)

	directorCmd.Flags().StringVar(&intervalDirector, "interval", "5s", "interval the Director will fetch matches")
	directorCmd.Flags().StringVar(&allocatorMode, "mode", "discover", "allocator mode for the director")
	directorCmd.Flags().StringVar(&octopsDiscoverArgs.DiscoverServiceURL, "octops-discover-url", "http://localhost:8081", "the Octops Discover server URL")
	directorCmd.Flags().StringVar(&agonesAllocatorArgs.KeyFile, "key", "", "the private key file for the client certificate in PEM format")
	directorCmd.Flags().StringVar(&agonesAllocatorArgs.CertFile, "cert", "", "the public key file for the client certificate in PEM format")
	directorCmd.Flags().StringVar(&agonesAllocatorArgs.CaCertFile, "cacert", "", "the CA cert file for server signing certificate in PEM format")
	directorCmd.Flags().StringVar(&agonesAllocatorArgs.AllocatorServiceHost, "allocator-host", "0.0.0.0", "the host address for allocator server")
	directorCmd.Flags().IntVar(&agonesAllocatorArgs.AllocatorServicePort, "allocator-port", 443, "the host address for allocator server")
	directorCmd.Flags().StringVar(&agonesAllocatorArgs.Namespace, "namespace", "default", "the game server kubernetes namespace")
	directorCmd.Flags().BoolVar(&agonesAllocatorArgs.MultiCluster, "multicluster", false, "set to true to enable the multi-cluster allocation")
}
