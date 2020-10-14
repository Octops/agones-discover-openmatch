package frontend

import (
	"github.com/Octops/agones-discover-openmatch/internal/runtime"
	"github.com/Octops/agones-discover-openmatch/pkg/config"
	"google.golang.org/grpc"
)

func FrontEndConn() (*grpc.ClientConn, error) {
	logger := runtime.Logger()
	omConfig := config.OpenMatch()

	logger.Infof("connecting to OpenMatch FrontEnd service: %s", omConfig.FrontEnd)
	return grpc.Dial(omConfig.FrontEnd, grpc.WithInsecure())
}
