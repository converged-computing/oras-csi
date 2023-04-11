package driver

import (
	"context"
	"math/rand"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type NodeService struct {
	csi.UnimplementedNodeServer
	Service

	mountPointsCount int
	mountPoints      []*orasHandler
	nodeId           string
}

var _ csi.NodeServer = &NodeService{}

var nodeCapabilities = []csi.NodeServiceCapability_RPC_Type{
	//csi.NodeServiceCapability_RPC_GET_VOLUME_STATS,
	//csi.NodeServiceCapability_RPC_VOLUME_CONDITION,
	//csi.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME,
}

// NewNodeService creates the node service that runs on every node.
func NewNodeService(container string, rootPath, pluginDataPath, nodeId string, mountPointsCount int) (*NodeService, error) {
	log.Infof("NewNodeService creation (container %s, rootDir %s, pluginDataDir %s, nodeId %s, mountPointsCount %d)", container, rootPath, pluginDataPath, nodeId, mountPointsCount)

	// One plugin per mount point. In layman's terms, there is 1:1 plugin:kubernetes node
	mountPoints := make([]*orasHandler, mountPointsCount)
	for i := 0; i < mountPointsCount; i++ {
		mountPoints[i] = NewOrasHandler(container, rootPath, pluginDataPath, nodeId, i, mountPointsCount)
		if err := mountPoints[i].MountOras(); err != nil {
			return nil, err
		}
	}
	if OrasLog {
		mountPoints[0].SetOrasLogging()
	}

	ns := &NodeService{
		mountPointsCount: mountPointsCount,
		mountPoints:      mountPoints,
		nodeId:           nodeId,
	}
	return ns, nil
}

func (ns *NodeService) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	log.Infof("NodePublishVolume - VolumeId: %s, Readonly: %v, VolumeContext %v, PublishContext %v, VolumeCapability %v TargetPath %s", req.GetVolumeId(), req.GetReadonly(), req.GetVolumeContext(), req.GetPublishContext(), req.GetVolumeCapability(), req.GetTargetPath())
	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "NodePublishVolume: VolumeId must be provided")
	}
	if req.TargetPath == "" {
		return nil, status.Error(codes.InvalidArgument, "NodePublishVolume: TargetPath must be provided")
	}
	if req.VolumeCapability == nil {
		return nil, status.Error(codes.InvalidArgument, "NodePublishVolume: VolumeCapability must be provided")
	}

	var source string
	if subDir, found := req.GetVolumeContext()["mfsSubDir"]; found {
		source = subDir
	} else {
		source = ns.mountPoints[0].MfsPathToVolume(req.VolumeId)
	}
	target := req.TargetPath
	options := req.VolumeCapability.GetMount().MountFlags
	if req.GetReadonly() {
		options = append(options, "ro")
	}
	if handler, err := ns.pickHandler(req.GetVolumeContext(), req.GetPublishContext()); err != nil {
		return nil, err
	} else {
		if err := handler.BindMount(source, target, options...); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	return &csi.NodePublishVolumeResponse{}, nil
}

func (ns *NodeService) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	log.Infof("NodeUnpublishVolume - VolumeId: %s, TargetPath: %s)", req.GetVolumeId(), req.GetTargetPath())
	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "NodeUnpublishVolume: Volume Id must be provided")
	}
	if req.TargetPath == "" {
		return nil, status.Error(codes.InvalidArgument, "NodeUnpublishVolume: Target Path must be provided")
	}

	found, err := ns.mountPoints[0].VolumeExist(req.VolumeId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	} else if !found {
		found, err = ns.mountPoints[0].MountVolumeExist(req.VolumeId)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		if !found {
			return nil, status.Errorf(codes.NotFound, "NodeUnpublishVolume: volume %s not found", req.VolumeId)
		}
	}
	if err = ns.mountPoints[0].BindUMount(req.TargetPath); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &csi.NodeUnpublishVolumeResponse{}, nil
}

func (ns *NodeService) NodeGetInfo(ctx context.Context, req *csi.NodeGetInfoRequest) (*csi.NodeGetInfoResponse, error) {
	log.Infof("NodeGetInfo")
	return &csi.NodeGetInfoResponse{
		NodeId: ns.nodeId,
	}, nil
}

func (ns *NodeService) NodeGetCapabilities(ctx context.Context, req *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	log.Infof("NodeGetCapabilities")
	var caps []*csi.NodeServiceCapability
	for _, capa := range nodeCapabilities {
		caps = append(caps, &csi.NodeServiceCapability{
			Type: &csi.NodeServiceCapability_Rpc{
				Rpc: &csi.NodeServiceCapability_RPC{
					Type: capa,
				},
			},
		})
	}
	return &csi.NodeGetCapabilitiesResponse{
		Capabilities: caps,
	}, nil
}

/*
func (ns *NodeService) NodeGetVolumeStats(ctx context.Context, req *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error) {
	log.Infof("NodeService::NodeGetVolumeStats (volume_id %s, volume_path %s, staging_path %s)",
		req.VolumeId, req.VolumePath, req.StagingTargetPath)

	if req.VolumeId == "" {
		return nil, status.Error(codes.InvalidArgument, "NodeGetVolumeStats: VolumeId must be provided")
	}
	if req.VolumePath == "" {
		return nil, status.Error(codes.InvalidArgument, "NodeGetVolumeStats: VolumePath must be provided")
	}

	cond := false
	_, err := ioutil.ReadDir(req.VolumePath)
	if err != nil {
		log.Infof("%s %s corrupted", req.VolumeId, req.VolumePath)
		cond = true
	} else {
		log.Infof("%s %s NOT corrupted", req.VolumeId, req.VolumePath)
	}
	return &csi.NodeGetVolumeStatsResponse{VolumeCondition: &csi.VolumeCondition{
		Abnormal: cond,
		Message:  "",
	}}, nil
}
*/
//////////////

// pickHandler - Returns proper handler. Currently picks random mfs handler.
func (ns *NodeService) pickHandler(volumeContext map[string]string, publishContext map[string]string) (*orasHandler, error) {
	if ns.mountPointsCount <= 0 {
		return nil, status.Error(codes.Internal, "pickHandler: there is no mfs handlers")
	}
	return ns.mountPoints[rand.Uint32()%uint32(ns.mountPointsCount)], nil
}

// pickHandlerFromVolumeId - Unimplemented, always picks first handler.
func (ns *NodeService) pickHandlerFromVolumeId(volumeId string) (*orasHandler, error) {
	if ns.mountPointsCount <= 0 {
		return nil, status.Error(codes.Internal, "pickHandlerFromVolumeId: there is no mfs handlers")
	}
	return ns.mountPoints[0], nil
}
