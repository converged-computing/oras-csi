package driver

import (
	"context"
	"fmt"
	"math/rand"

	handler "github.com/billy-playground/oras-csi/pkg/oras"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type NodeService struct {
	csi.UnimplementedNodeServer
	Service

	// Number of oras handlers per node
	handlersCount int
	handlers      []*handler.OrasHandler
	nodeId        string
}

var _ csi.NodeServer = &NodeService{}

// NodeStageVolume only exists to validate arguments
func (ns *NodeService) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}
	if len(req.GetStagingTargetPath()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Target path missing in request")
	}
	if req.GetVolumeCapability() == nil {
		return nil, status.Error(codes.InvalidArgument, "Volume capability missing in request")
	}

	return &csi.NodeStageVolumeResponse{}, nil
}

func (ns *NodeService) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}
	if len(req.GetStagingTargetPath()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Target path missing in request")
	}

	return &csi.NodeUnstageVolumeResponse{}, nil
}

// NewNodeService creates the node service that runs on every node.
func NewNodeService(rootPath, pluginDataPath, nodeId string, handlersCount int, enforceNamespaces bool) (*NodeService, error) {
	log.Infof("NewNodeService creation (rootDir %s, pluginDataDir %s, nodeId %s, handlersCount %d)", rootPath, pluginDataPath, nodeId, handlersCount)

	// Create N handlers per node
	handlers := make([]*handler.OrasHandler, handlersCount)
	for i := 0; i < handlersCount; i++ {
		handlers[i] = handler.NewOrasHandler(rootPath, pluginDataPath, enforceNamespaces, nodeId, i, handlersCount)
	}
	if OrasLog {
		handlers[0].SetOrasLogging()
	}

	ns := &NodeService{
		handlersCount: handlersCount,
		handlers:      handlers,
		nodeId:        nodeId,
	}
	return ns, nil
}

func (ns *NodeService) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	log.Infof("NodePublishVolume - VolumeId: %s, Readonly: %v, VolumeCapability %v TargetPath %s", req.GetVolumeId(), req.GetReadonly(), req.GetVolumeCapability(), req.GetTargetPath())
	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "NodePublishVolume: VolumeId must be provided")
	}
	if req.TargetPath == "" {
		return nil, status.Error(codes.InvalidArgument, "NodePublishVolume: TargetPath must be provided")
	}

	log.Info("Looking for volume context....")
	volumeContext := req.GetVolumeContext()

	// We are required to be provided with the container URI
	// log.Info(volumeContext)
	settings, err := handler.NewSettings(volumeContext)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("NodePublishVolume: Issue preparing settings %s", err))
	}

	// Prepare a directory just for the artifact
	// Since this is a bind mount, it can be bound more than once
	source, err := ns.handlers[0].OrasPathToVolume(settings)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("NodePublishVolume: Unable to prepare container %s", err))
	}
	log.Info("volume source directory:", source)

	target := req.TargetPath
	log.Info("volume target directory:", target)

	options := req.VolumeCapability.GetMount().MountFlags
	if req.GetReadonly() {
		options = append(options, "ro")
	}
	log.Info("volume options:", options)
	if handler, err := ns.getHandler(req.GetVolumeContext(), req.GetPublishContext()); err != nil {
		log.Info("issue getting handler:", err)
		return nil, err
	} else {
		if err := handler.BindMount(source, target, options...); err != nil {
			log.Info("issue with bind mount:", err)
			return nil, status.Error(codes.Internal, err.Error())
		}
		log.Info("Mount had no errors")
	}
	log.Info("returning NodePublishVolumeResponse")
	return &csi.NodePublishVolumeResponse{}, nil
}

// NodeUnpublishVolume removes the bind
func (ns *NodeService) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	log.Infof("NodeUnpublishVolume - VolumeId: %s, TargetPath: %s)", req.GetVolumeId(), req.GetTargetPath())
	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "NodeUnpublishVolume: Volume Id must be provided")
	}
	if req.TargetPath == "" {
		return nil, status.Error(codes.InvalidArgument, "NodeUnpublishVolume: Target Path must be provided")
	}

	// We don't have the original container URI, so we don't check that it exists.
	// We are required to be provided with the container URI
	if err := ns.handlers[0].BindUMount(req.TargetPath); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &csi.NodeUnpublishVolumeResponse{}, nil
}

// NodeGetInfo returns the nodeid
func (ns *NodeService) NodeGetInfo(ctx context.Context, req *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	log.Infof("NodeGetInfo")
	return &csi.NodeGetInfoResponse{
		NodeId: ns.nodeId,
	}, nil
}

func (ns *NodeService) NodeGetCapabilities(ctx context.Context, req *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	caps := []*csi.NodeServiceCapability{
		{
			Type: &csi.NodeServiceCapability_Rpc{
				Rpc: &csi.NodeServiceCapability_RPC{
					Type: csi.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME,
				},
			},
		},
	}
	return &csi.NodeGetCapabilitiesResponse{
		Capabilities: caps,
	}, nil
}

// getHandler - adds a check that we have handlers and returns a random handler from our set
func (ns *NodeService) getHandler(volumeContext map[string]string, publishContext map[string]string) (*handler.OrasHandler, error) {

	// No handlers created - this shouldn't happen
	if ns.handlersCount <= 0 {
		return nil, status.Error(codes.Internal, "orasHandler: there are no handlers")
	}

	// Randomly choose one
	return ns.handlers[rand.Uint32()%uint32(ns.handlersCount)], nil
}

// NodeExpandVolume not implemented
func (ns *NodeService) NodeExpandVolume(ctx context.Context, req *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "NodeExpandVolume is not implemented")
}

// NodeGetVolumeStats not implemented
func (ns *NodeService) NodeGetVolumeStats(ctx context.Context, in *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "")
}
