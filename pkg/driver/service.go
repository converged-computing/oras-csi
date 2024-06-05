package driver

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/billy-playground/oras-csi/pkg/utils"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

const (
	driverName    = "csi.oras.land"
	driverVersion = "0.1.0"
)

type Service interface{}

// I think this refers to csi-sanity
// https://github.com/kubernetes-csi/csi-test/blob/v4.4.0/cmd/csi-sanity/main.go
var SanityTestRun bool // is this a test run?
var OrasLog bool       // enable logging
var log logrus.Logger  // level for logrus

// Init primarily sets logging levels for TBA services
func Init(sanityTestRun bool, logLevel int, orasLog bool) error {
	log = *logrus.New()

	// Print the cool logo!
	for _, line := range strings.Split(utils.GetLogo(), "\n") {
		log.Info(strings.ReplaceAll(line, "\t", ""))
	}
	SanityTestRun = sanityTestRun
	log.SetLevel(logrus.Level(logLevel))
	OrasLog = orasLog
	return nil
}

func StartService(service *Service, mode, csiEndpoint string) error {
	log.Infof("StartService - endpoint %s", csiEndpoint)
	gRPCServer := CreategRPCServer()
	listener, err := CreateListener(csiEndpoint)
	if err != nil {
		return err
	}
	csi.RegisterIdentityServer(gRPCServer, &IdentityService{})

	switch (*service).(type) {
	case *NodeService:
		log.Infof("StartService - Registering node service")
		csi.RegisterNodeServer(gRPCServer, (*service).(csi.NodeServer))
	default:
		return fmt.Errorf("StartService: Unrecognized service type: %T", service)
	}

	log.Info("StartService - Starting to serve!")
	err = gRPCServer.Serve(listener)
	if err != nil {
		return err
	}
	log.Info("StartService - gRPCServer stopped without an error!")
	return nil
}

// CreateListener create listener ready for communication over given csi endpoint
func CreateListener(csiEndpoint string) (net.Listener, error) {
	log.Infof("CreateListener - endpoint %s", csiEndpoint)

	u, err := url.Parse(csiEndpoint)
	if err != nil {
		return nil, fmt.Errorf("CreateListener - Unable to parse address: %q", err)
	}

	addr := path.Join(u.Host, filepath.FromSlash(u.Path))
	if u.Host == "" {
		addr = filepath.FromSlash(u.Path)
	}

	// CSI plugins talk only over UNIX sockets currently
	if u.Scheme != "unix" {
		return nil, fmt.Errorf("CreateListener - Currently only unix domain sockets are supported, have: %s", u.Scheme)
	} else {
		// remove the socket if it's already there. This can happen if we
		// deploy a new version and the socket was created from the old running
		// plugin.
		log.Infof("CreateListener - Removing socket %s", addr)
		if err := os.Remove(addr); err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("CreateListener - Failed to remove unix domain socket file %s, error: %s", addr, err)
		}
	}

	listener, err := net.Listen(u.Scheme, addr)
	if err != nil {
		return nil, fmt.Errorf("CreateListener - Failed to listen: %v", err)
	}

	return listener, nil
}

func CreategRPCServer() *grpc.Server {
	log.Info("CreategRPCServer")
	// log response errors for better observability
	errHandler := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		resp, err := handler(ctx, req)
		if err != nil {
			stat, rpcErr := status.FromError(err)
			if rpcErr {
				log.Errorf("rpc error: %s - %s", stat.Code(), stat.Message())
			} else {
				log.Errorf("unexpected error type - %s", err.Error())
			}
		}
		return resp, err
	}
	return grpc.NewServer(grpc.UnaryInterceptor(errHandler))
}
