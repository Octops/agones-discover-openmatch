package openmatch

import (
	"github.com/Octops/agones-discover-openmatch/pkg/config"
	"google.golang.org/grpc"
)

func ConnFuncInsecure() (*grpc.ClientConn, error) {
	conn, err := grpc.Dial(config.OpenMatch().BackEnd, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// TODO: Implement secure dialer
func ConnFuncSecure() (*grpc.ClientConn, error) {
	panic("implement me")
}
