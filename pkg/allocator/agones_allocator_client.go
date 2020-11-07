package allocator

import (
	pb "agones.dev/agones/pkg/allocation/go"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io/ioutil"
)

var (
	ErrKeyFileInvalid    = errors.New("the private key file for the client certificate is invalid")
	ErrCertFileInvalid   = errors.New("the public key file for the client certificate is invalid")
	ErrCaCertFileInvalid = errors.New("the CA cert file for server signing certificate is invalid")
)

type AgonesAllocatorClientConfig struct {
	KeyFile              string
	CertFile             string
	CaCertFile           string
	AllocatorServiceHost string
	AllocatorServicePort int
	Namespace            string
	MultiCluster         bool
}

// gRPC Client for Agones Allocator Service
type AgonesAllocatorClient struct {
	Config *AgonesAllocatorClientConfig
}

func NewAgonesAllocatorClient(config *AgonesAllocatorClientConfig) (*AgonesAllocatorClient, error) {
	var validationErrors error

	ok, err := ValueIsEmpty(config.KeyFile, ErrKeyFileInvalid)
	if !ok && err != nil {
		validationErrors = errors.Wrap(validationErrors, err.Error())
	}

	ok, err = ValueIsEmpty(config.CertFile, ErrCertFileInvalid)
	if !ok && err != nil {
		validationErrors = errors.Wrap(validationErrors, err.Error())
	}

	//ok, err = ValueIsEmpty(config.CaCertFile, ErrCaCertFileInvalid)
	//if !ok && err != nil {
	//	validationErrors = errors.Wrap(validationErrors, err.Error())
	//}

	//TODO Validate Host, Port and Namespace

	if validationErrors != nil {
		return nil, validationErrors
	}

	return &AgonesAllocatorClient{Config: config}, nil
}

func (c *AgonesAllocatorClient) Allocate(ctx context.Context, request *pb.AllocationRequest) (*pb.AllocationResponse, error) {
	cert, err := ioutil.ReadFile(c.Config.CertFile)
	if err != nil {
		panic(err)
	}
	key, err := ioutil.ReadFile(c.Config.KeyFile)
	if err != nil {
		panic(err)
	}
	cacert, err := ioutil.ReadFile(c.Config.CaCertFile)
	if err != nil {
		panic(err)
	}

	dialOpts, err := createRemoteClusterDialOption(cert, key, cacert)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create dial option")
	}

	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", c.Config.AllocatorServiceHost, c.Config.AllocatorServicePort), dialOpts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create dial remote allocator service")
	}
	defer conn.Close()

	grpcClient := pb.NewAllocationServiceClient(conn)
	response, err := grpcClient.Allocate(context.Background(), request)
	if err != nil {
		return nil, errors.Wrap(err, "failed to allocate request")
	}

	return response, nil
}

// createRemoteClusterDialOption creates a grpc client dial option with TLS configuration.
func createRemoteClusterDialOption(clientCert, clientKey, caCert []byte) (grpc.DialOption, error) {
	// Load client cert
	cert, err := tls.X509KeyPair(clientCert, clientKey)
	if err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{Certificates: []tls.Certificate{cert}}
	if len(caCert) != 0 {
		// Load CA cert, if provided and trust the server certificate.
		// This is required for self-signed certs.
		tlsConfig.RootCAs = x509.NewCertPool()
		if !tlsConfig.RootCAs.AppendCertsFromPEM(caCert) {
			return nil, errors.New("only PEM format is accepted for server CA")
		}
	}

	return grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)), nil
}
