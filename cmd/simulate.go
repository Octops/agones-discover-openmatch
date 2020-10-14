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
	"github.com/Octops/agones-discover-openmatch/pkg/app"
	"github.com/spf13/cobra"
)

var (
	interval    string
	playersPool int
)

// simulateCmd represents the simulate command
var simulateCmd = &cobra.Command{
	Use:   "simulate",
	Short: "Simulate players requesting matches on a interval basis",
	Long:  `The Player Simulator will request matches on a interval basis for different pool of players.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger := runtime.NewLogger(true).WithField("source", "player_simulator")
		if err := app.RunPlayerSimulator(logger, context.Background(), interval, playersPool); err != nil {
			logger.Fatal(err)
		}
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
	simulateCmd.Flags().StringVar(&interval, "interval", "5s", "interval between match requests, 10s, 1m, 5m")
	simulateCmd.Flags().IntVar(&playersPool, "players-pool", 10, "number of players to create matchmaking requests")
}
