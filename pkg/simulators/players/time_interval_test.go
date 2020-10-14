package players

import (
	"context"
	"github.com/Octops/agones-discover-openmatch/internal/runtime"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"open-match.dev/open-match/pkg/pb"
	"sync"
	"testing"
)

func TestTimeIntervalPlayerSimulator_CreatePlayers(t *testing.T) {
	testCases := []struct {
		name         string
		playersCount int
		wantErr      bool
	}{
		{
			name:         "it should return error for negative number of players",
			playersCount: -1,
			wantErr:      true,
		},
		{
			name:         "it should not create players",
			playersCount: 0,
			wantErr:      false,
		},
		{
			name:         "it should create 1 player",
			playersCount: 1,
			wantErr:      false,
		},
		{
			name:         "it should create 10 players",
			playersCount: 10,
			wantErr:      false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := &TimeIntervalPlayerSimulator{
				mux:     &sync.Mutex{},
				logger:  runtime.NewLogger(true),
				Players: []*Player{},
			}

			players, err := p.CreatePlayers(tc.playersCount)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.playersCount, len(players))
				for _, player := range players {
					p.logger.Infof("StringArgs: %s", player.MatchRequest.StringArgs)
				}
			}
		})
	}
}

func TestTimeIntervalPlayerSimulator_RequestMatchForPlayers(t *testing.T) {
	testCases := []struct {
		name    string
		players []*Player
		wantErr bool
	}{
		{
			name: "it should create a ticker for 1 player",
			players: []*Player{
				{
					MatchRequest: &MatchRequest{
						Ticket: nil,
						StringArgs: map[string]string{
							"region": "us-east-2",
							"world":  "Dune",
							"skill":  "2",
						},
					},
				},
			},
			wantErr: false,
		},
	}

	requestMatchFunc := func(ctx context.Context, request *pb.CreateTicketRequest, opts ...grpc.CallOption) (*pb.Ticket, error) {
		return &pb.Ticket{
			Id:           uuid.New().String(),
			SearchFields: request.Ticket.SearchFields,
		}, nil
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := &TimeIntervalPlayerSimulator{
				mux:              &sync.Mutex{},
				logger:           runtime.NewLogger(true),
				Players:          []*Player{},
				RequestMatchFunc: requestMatchFunc,
			}

			err := p.RequestMatchForPlayers(tc.players)

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				for _, player := range tc.players {
					require.NotNil(t, player.MatchRequest.Ticket)
				}
				require.Equal(t, len(tc.players), len(p.Players))
			}
		})
	}
}
