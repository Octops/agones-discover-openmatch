package director

import (
	"context"
	"github.com/Octops/agones-discover-openmatch/internal/runtime"
	"github.com/pkg/errors"
	"open-match.dev/open-match/pkg/pb"
	"sync"
	"time"
)

type GenerateProfilesFunc func() ([]*pb.MatchProfile, error)

type FetchMatchesFunc func(ctx context.Context, profile *pb.MatchProfile) ([]*pb.Match, error)

type AssignFunc func(ctx context.Context, matches []*pb.Match) error

type DirectorFunc func(ctx context.Context, profilesFunc GenerateProfilesFunc, matchesFunc FetchMatchesFunc, assignFunc AssignFunc) error

func Run() DirectorFunc {
	return func(ctx context.Context, profilesFunc GenerateProfilesFunc, matchesFunc FetchMatchesFunc, assignFunc AssignFunc) error {
		logger := runtime.Logger().WithField("source", "director")
		profiles, err := profilesFunc()
		if err != nil {
			return errors.Wrap(err, "failed to generate profiles")
		}

		ticker := time.NewTicker(time.Second * 5)
		for {
			select {
			case <-ticker.C:
				var wg sync.WaitGroup

				for _, p := range profiles {
					wg.Add(1)
					go func(wg *sync.WaitGroup, p *pb.MatchProfile) {
						defer wg.Done()
						matches, err := matchesFunc(ctx, p)
						if err != nil {
							logger.Error(errors.Wrap(err, "failed to fetch matches"))
						}

						if err := assignFunc(ctx, matches); err != nil {
							logger.Error(errors.Wrap(err, "failed to assign matches"))
						}
					}(&wg, p)
				}
				wg.Wait()
			case <-ctx.Done():
				logger.Info("stopping director")
				ticker.Stop()
				timeout, cancel := context.WithTimeout(ctx, time.Second)
				defer cancel()

				<-timeout.Done()
				return nil
			}
		}
	}
}
