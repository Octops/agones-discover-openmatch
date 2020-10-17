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
	"strconv"
	"sync"
	"time"
)

type MatchRequest struct {
	Ticket     *pb.Ticket
	StringArgs map[string]string
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
				StringArgs: CreateStringArgs(),
			}})
	}

	return players, nil
}

func (p *TimeIntervalPlayerSimulator) RequestMatchForPlayers(players []*Player) error {
	for _, player := range players {
		req := &pb.CreateTicketRequest{
			Ticket: &pb.Ticket{
				SearchFields: &pb.SearchFields{
					// TODO: Split player request across search fields. Latency must be a DoubleRange
					StringArgs: player.MatchRequest.StringArgs,
				},
			},
		}

		ticket, err := p.RequestMatchFunc(context.Background(), req)
		if err != nil {
			return err
		}

		player.MatchRequest.Ticket = ticket
		p.logger.Debugf("ticketID=%s playerUID=%s tags=%s", ticket.GetId(), player.UID, ticket.SearchFields.StringArgs)
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
	skillLevels := []int{1, 2, 3, 4, 5}
	latencies := []int{20, 30, 50, 100}

	region := TagFromStringSlice(regionTags)
	world := TagFromStringSlice(worldTags)
	skill := TagFromIntSlice(skillLevels)
	latency := TagFromIntSlice(latencies)

	return map[string]string{
		"region":  region,
		"world":   world,
		"skill":   strconv.Itoa(skill),
		"latency": strconv.Itoa(latency),
	}
}

func TagFromStringSlice(tags []string) string {
	rand.Seed(time.Now().UTC().UnixNano())
	randomIndex := rand.Intn(len(tags))

	return tags[randomIndex]
}

func TagFromIntSlice(tags []int) int {
	rand.Seed(time.Now().UTC().UnixNano())
	randomIndex := rand.Intn(len(tags))

	return tags[randomIndex]
}
