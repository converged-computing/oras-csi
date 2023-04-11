package driver

import (
	"context"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/protobuf/ptypes/wrappers"
)

type IdentityService struct {
	csi.UnimplementedIdentityServer
	Service
}

var _ csi.IdentityServer = &IdentityService{}

func (is *IdentityService) GetPluginInfo(ctx context.Context, req *csi.GetPluginInfoRequest) (*csi.GetPluginInfoResponse, error) {
	log.Infof("GetPluginInfo")

	return &csi.GetPluginInfoResponse{
		Name:          driverName,
		VendorVersion: driverVersion,
	}, nil
}

// GetPluginCapabilities returns the message that this provides a controller service, and does not provide volume expansion
func (is *IdentityService) GetPluginCapabilities(ctx context.Context, req *csi.GetPluginCapabilitiesRequest) (*csi.GetPluginCapabilitiesResponse, error) {
	log.Infof("GetPluginCapabilities")

	return &csi.GetPluginCapabilitiesResponse{
		Capabilities: []*csi.PluginCapability{
			{
				Type: &csi.PluginCapability_Service_{
					Service: &csi.PluginCapability_Service{
						Type: csi.PluginCapability_Service_CONTROLLER_SERVICE,
					},
				},
			},

			// Volumes cannot be expanded
			{
				Type: &csi.PluginCapability_VolumeExpansion_{
					VolumeExpansion: &csi.PluginCapability_VolumeExpansion{
						Type: csi.PluginCapability_VolumeExpansion_OFFLINE,
					},
				},
			},
		},
	}, nil
}

// Probe indicates that the plugin is in a healthy state
func (is *IdentityService) Probe(ctx context.Context, req *csi.ProbeRequest) (*csi.ProbeResponse, error) {
	log.Infof("Probe")

	return &csi.ProbeResponse{
		Ready: &wrappers.BoolValue{
			Value: true,
		},
	}, nil
}
