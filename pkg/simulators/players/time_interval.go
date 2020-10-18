package players

import (
	"context"
	"fmt"
	"github.com/Octops/agones-discover-openmatch/internal/runtime"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"math/rand"
	"open-match.dev/open-match/pkg/pb"
	"sync"
	"time"
)

const (
	GAME_MODE_SESSION = "mode.session"
)

type MatchRequest struct {
	Ticket     *pb.Ticket
	Tags       []string
	StringArgs map[string]string
	DoubleArgs map[string]float64
}

type Player struct {
	UID          string
	MatchRequest *MatchRequest
}

type RequestMatchFunc func(ctx context.Context, ticket *pb.CreateTicketRequest, opts ...grpc.CallOption) (*pb.Ticket, error)

/*
- Create pool of players
- Request match on a interval basis
*/
type TimeIntervalPlayerSimulator struct {
	mux              *sync.Mutex
	logger           *logrus.Entry
	Interval         time.Duration
	PlayersPool      int
	RequestMatchFunc RequestMatchFunc
	Players          []*Player
}

func NewTimeIntervalPlayerSimulator(interval string, playersPool int, requestMatchFunc RequestMatchFunc) (*TimeIntervalPlayerSimulator, error) {
	duration, err := time.ParseDuration(interval)
	if err != nil {
		return nil, err
	}

	return &TimeIntervalPlayerSimulator{
		mux:              &sync.Mutex{},
		logger:           runtime.NewLogger(true),
		Interval:         duration,
		PlayersPool:      playersPool,
		RequestMatchFunc: requestMatchFunc,
		Players:          []*Player{},
	}, nil
}

func (p *TimeIntervalPlayerSimulator) Run(ctx context.Context) error {
	p.logger.WithFields(logrus.Fields{
		"interval":     p.Interval,
		"players_pool": p.PlayersPool,
	}).Infof("starting Players Simulator")
	ctxSimulator, cancel := context.WithCancel(ctx)
	ticker := time.NewTicker(p.Interval)

	go func() {
		defer func() {
			ticker.Stop()
			cancel()
		}()

		//Create fist batch before ticker. Useful if longer intervals are set
		p.CreateMatchmakingRequests()

		for {
			select {
			case t := <-ticker.C:
				p.logger.Infof("create matchmaking requests for %d Players at %s", p.PlayersPool, t.String())
				p.CreateMatchmakingRequests()
			case <-ctxSimulator.Done():
				return
			}
		}
	}()

	<-ctx.Done()
	return nil
}

func (p *TimeIntervalPlayerSimulator) CreatePlayers(count int) ([]*Player, error) {
	players := []*Player{}
	if count < 0 {
		return players, fmt.Errorf("number of player can't be lower than zero: %d", count)
	}

	for i := 0; i < count; i++ {

		players = append(players, &Player{
			UID: uuid.New().String(),
			MatchRequest: &MatchRequest{
				Tags:       []string{GAME_MODE_SESSION},
				StringArgs: CreateStringArgs(),
				DoubleArgs: CreateDoubleArgs(),
			}})
	}

	return players, nil
}

func (p *TimeIntervalPlayerSimulator) RequestMatchForPlayers(players []*Player) error {
	for _, player := range players {
		req := &pb.CreateTicketRequest{
			Ticket: &pb.Ticket{
				SearchFields: &pb.SearchFields{
					Tags:       player.MatchRequest.Tags,
					StringArgs: player.MatchRequest.StringArgs,
					DoubleArgs: player.MatchRequest.DoubleArgs,
				},
			},
		}

		ticket, err := p.RequestMatchFunc(context.Background(), req)
		if err != nil {
			return err
		}

		player.MatchRequest.Ticket = ticket
		p.logger.Debugf("ticketID=%s playerUID=%s stringArgs=%s doubleArgs=%v", ticket.GetId(), player.UID, ticket.SearchFields.StringArgs, ticket.SearchFields.DoubleArgs)
	}

	p.AddPlayers(players)
	return nil
}

func (p *TimeIntervalPlayerSimulator) AddPlayers(players []*Player) {
	p.mux.Lock()
	defer p.mux.Unlock()

	p.Players = append(p.Players, players...)
}

func (p *TimeIntervalPlayerSimulator) CreateMatchmakingRequests() {
	go func() {
		players, err := p.CreatePlayers(p.PlayersPool)
		if err != nil {
			p.logger.Error(err)
		}

		err = p.RequestMatchForPlayers(players)
		if err != nil {
			p.logger.Error(err)
		}

		p.logger.Infof("total Players: %d", len(p.Players))
	}()
}

func CreateDoubleArgs() map[string]float64 {
	skillLevels := []float64{10, 100, 1000}
	latencies := []float64{25, 50, 75, 100}
	skill := TagFromFloatSlice(skillLevels)
	latency := TagFromFloatSlice(latencies)

	return map[string]float64{
		"skill":   skill,
		"latency": latency,
	}
}

func CreateStringArgs() map[string]string {
	regionTags := []string{
		"us-east-1",
		"us-east-2",
		"us-west-1",
		"us-west-2",
	}
	worldTags := []string{
		"Dune",
		"Nova",
		"Pandora",
		"Orion",
	}

	region := TagFromStringSlice(regionTags)
	world := TagFromStringSlice(worldTags)

	return map[string]string{
		"region": region,
		"world":  world,
	}
}

func TagFromStringSlice(tags []string) string {
	rand.Seed(time.Now().UTC().UnixNano())
	randomIndex := rand.Intn(len(tags))

	return tags[randomIndex]
}

func TagFromFloatSlice(tags []float64) float64 {
	rand.Seed(time.Now().UTC().UnixNano())
	randomIndex := rand.Intn(len(tags))

	return tags[randomIndex]
}
