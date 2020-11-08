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
	ErrKeyFileInvalid              = errors.New("the private key file for the client certificate is invalid")
	ErrCertFileInvalid             = errors.New("the public key file for the client certificate is invalid")
	ErrCaCertFileInvalid           = errors.New("the CA cert file for server signing certificate is invalid")
	ErrAllocatorServiceHostInvalid = errors.New("the Allocator Service host is invalid")
	ErrAllocatorServicePortInvalid = errors.New("the Allocator Service port is invalid")
	ErrAllocatorServiceNamespace   = errors.New("the Allocator Service namespace is invalid")
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

// AgonesAllocatorClient is the gRPC Client for Agones Allocator Service
type AgonesAllocatorClient struct {
	Config   *AgonesAllocatorClientConfig
	DialOpts grpc.DialOption
}

func NewAgonesAllocatorClient(config *AgonesAllocatorClientConfig) (*AgonesAllocatorClient, error) {
	if err := validateClientConfig(config); err != nil {
		return nil, err
	}

	cert, key, ca, err := loadCertificates(config)
	if err != nil {
		return nil, err
	}

	dialOpts, err := createRemoteClusterDialOption(cert, key, ca)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create dial option")
	}

	return &AgonesAllocatorClient{
		Config:   config,
		DialOpts: dialOpts,
	}, nil
}

func (c *AgonesAllocatorClient) Allocate(ctx context.Context, request *pb.AllocationRequest) (*pb.AllocationResponse, error) {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", c.Config.AllocatorServiceHost, c.Config.AllocatorServicePort), c.DialOpts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create dial remote allocator service")
	}
	defer conn.Close()

	grpcClient := pb.NewAllocationServiceClient(conn)
	response, err := grpcClient.Allocate(context.Background(), request)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func validateClientConfig(config *AgonesAllocatorClientConfig) error {
	var validationErrors error

	ok, err := ValueIsEmpty(config.KeyFile, ErrKeyFileInvalid)
	if !ok && err != nil {
		validationErrors = errors.Wrap(validationErrors, err.Error())
	}

	ok, err = ValueIsEmpty(config.CertFile, ErrCertFileInvalid)
	if !ok && err != nil {
		validationErrors = errors.Wrap(validationErrors, err.Error())
	}

	ok, err = ValueIsEmpty(config.AllocatorServiceHost, ErrAllocatorServiceHostInvalid)
	if !ok && err != nil {
		validationErrors = errors.Wrap(validationErrors, err.Error())
	}

	if config.AllocatorServicePort <= 0 {
		validationErrors = errors.Wrap(validationErrors, ErrAllocatorServicePortInvalid.Error())
	}

	ok, err = ValueIsEmpty(config.Namespace, ErrAllocatorServiceNamespace)
	if !ok && err != nil {
		validationErrors = errors.Wrap(validationErrors, err.Error())
	}

	return validationErrors
}

func loadCertificates(config *AgonesAllocatorClientConfig) (cert, key, ca []byte, err error) {
	cert, err = ioutil.ReadFile(config.CertFile)
	if err != nil {
		return nil, nil, nil, errors.Wrapf(err, "failed to read %s", config.CertFile)
	}

	key, err = ioutil.ReadFile(config.KeyFile)
	if err != nil {
		return nil, nil, nil, errors.Wrapf(err, "failed to read %s", config.KeyFile)
	}

	if len(config.CaCertFile) > 0 {
		ca, err = ioutil.ReadFile(config.CaCertFile)
		if err != nil {
			return nil, nil, nil, errors.Wrapf(err, "failed to read %s", config.CaCertFile)
		}
	}

	return cert, key, ca, nil
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
