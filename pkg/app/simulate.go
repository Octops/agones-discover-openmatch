package app

import (
	"context"
	"github.com/Octops/agones-discover-openmatch/internal/runtime"
	"github.com/Octops/agones-discover-openmatch/pkg/frontend"
	"github.com/Octops/agones-discover-openmatch/pkg/simulators/players"
	"github.com/sirupsen/logrus"
)

func RunPlayerSimulator(logger *logrus.Entry, ctx context.Context, interval string, playersPool int) error {
	ctx, cancel := context.WithCancel(context.Background())
	runtime.SetupSignal(cancel)

	conn, err := frontend.FrontEndConn()
	if err != nil {
		logger.Fatal(err)
	}

	feService, err := frontend.NewFrontEndService(conn)
	if err != nil {
		logger.Fatal(err)
	}

	simulator, err := players.NewTimeIntervalPlayerSimulator(interval, playersPool, feService.CreateTicket)
	if err != nil {
		logger.Fatal(err)
	}

	if err := simulator.Run(ctx); err != nil {
		logger.Fatal(err)
	}

	defer conn.Close()

	return nil
}
