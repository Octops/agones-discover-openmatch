package openmatch

import (
	"context"
	"fmt"
	"github.com/Octops/agones-discover-openmatch/internal/runtime"
	"github.com/Octops/agones-discover-openmatch/pkg/config"
	"github.com/Octops/agones-discover-openmatch/pkg/director"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"io"
	"math/rand"
	"open-match.dev/open-match/pkg/pb"
	"time"
)

type MatchFunctionServer struct {
	HostName string
	Port     int32
}

type FetchResponse struct {
	Matches []*pb.Match
	Err     error
}

func RunDirector(ctx context.Context, logger *logrus.Entry) error {
	conn, err := grpc.Dial(config.OpenMatch().BackEnd, grpc.WithInsecure())
	if err != nil {
		logger.Error(errors.Wrap(err, "failed to connect to Open Match Backend"))
	}

	defer conn.Close()
	client := pb.NewBackendServiceClient(conn)

	fetch := FetchMatches(client, MatchFunctionServer{
		HostName: "0.0.0.0",
		Port:     8082,
	})

	assign := AssignTickets(client)
	profiles := GenerateProfiles()

	if err := director.Run()(ctx, profiles, fetch, assign); err != nil {
		logger.Error(errors.Wrap(err, ""))
		return err
	}

	return nil
}

func FetchMatches(client pb.BackendServiceClient, matchFunctionServer MatchFunctionServer) director.FetchMatchesFunc {
	return func(ctx context.Context, profile *pb.MatchProfile) ([]*pb.Match, error) {
		logger := runtime.Logger()
		ctxFetch, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()

		fetchResponse := FetchResponse{}

		go func(p *pb.MatchProfile) {
			defer cancel()
			if fetchResponse.Matches, fetchResponse.Err = fetch(ctxFetch, client, profile, matchFunctionServer); fetchResponse.Err != nil {
				logger.Error(errors.Wrap(fetchResponse.Err, "failed to fetch matches from Open Match Backend"))
			}
		}(profile)

		select {
		case <-ctxFetch.Done():
			return fetchResponse.Matches, fetchResponse.Err
		}
	}
}

func AssignTickets(client pb.BackendServiceClient) director.AssignFunc {
	return func(ctx context.Context, matches []*pb.Match) error {
		logger := runtime.Logger()
		for _, match := range matches {
			ticketIDs := []string{}
			for _, t := range match.GetTickets() {
				ticketIDs = append(ticketIDs, t.Id)
			}

			// TODO: This should be extracted to a proper service that will consume from Agones Discover
			conn := fmt.Sprintf("%d.%d.%d.%d:2222", rand.Intn(256), rand.Intn(256), rand.Intn(256), rand.Intn(256))
			req := &pb.AssignTicketsRequest{
				Assignments: []*pb.AssignmentGroup{
					{
						TicketIds: ticketIDs,
						Assignment: &pb.Assignment{
							Connection: conn,
						},
					},
				},
			}

			if _, err := client.AssignTickets(context.Background(), req); err != nil {
				return fmt.Errorf("AssignTickets failed for match %v, got %w", match.GetMatchId(), err)
			}

			logger.Debugf("Assigned server %v to match %v", conn, match.GetMatchId())
		}

		return nil
	}
}

func GenerateProfiles() director.GenerateProfilesFunc {
	return func() ([]*pb.MatchProfile, error) {
		var profiles []*pb.MatchProfile
		worlds := []string{"Dune", "Nova", "Pandora", "Orion"}
		for _, world := range worlds {
			profiles = append(profiles, &pb.MatchProfile{
				Name: "mode_based_profile",
				Pools: []*pb.Pool{
					{
						Name: "pool_mode_" + world,
						TagPresentFilters: []*pb.TagPresentFilter{
							{
								Tag: world,
							},
						},
					},
				},
			})
		}

		return profiles, nil
	}
}

func fetch(ctx context.Context, client pb.BackendServiceClient, profile *pb.MatchProfile, matchFunctionServer MatchFunctionServer) ([]*pb.Match, error) {
	req := &pb.FetchMatchesRequest{
		Config: &pb.FunctionConfig{
			Host: matchFunctionServer.HostName,
			Port: matchFunctionServer.Port,
			Type: pb.FunctionConfig_GRPC,
		},
		Profile: profile,
	}

	stream, err := client.FetchMatches(ctx, req)
	if err != nil {
		return nil, err
	}

	var result []*pb.Match
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		result = append(result, resp.GetMatch())
	}

	return result, nil
}
