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
	"github.com/Octops/agones-discover-openmatch/pkg/simulators/players"
	"google.golang.org/grpc"
	"open-match.dev/open-match/pkg/pb"

	"github.com/spf13/cobra"
)

// simulateCmd represents the simulate command
var simulateCmd = &cobra.Command{
	Use:   "simulate",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger := runtime.NewLogger(true).WithField("source", "player_simulator")
		conn, err := grpc.Dial("0.0.0.0:50504", grpc.WithInsecure())
		if err != nil {
			logger.Fatalf("Failed to connect to Open Match, got %v", err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		runtime.SetupSignal(cancel)

		feService := pb.NewFrontendServiceClient(conn)
		simulator, err := players.NewTimeIntervalPlayerSimulator("5s", 20, feService.CreateTicket)
		if err != nil {
			logger.Fatal(err)
		}

		if err := simulator.Run(ctx); err != nil {
			logger.Fatal(err)
		}

		defer conn.Close()
	},
}

func init() {
	playerCmd.AddCommand(simulateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// simulateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// simulateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
