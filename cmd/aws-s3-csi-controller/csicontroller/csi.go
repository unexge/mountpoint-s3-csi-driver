package csicontroller

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/go-logr/logr"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"k8s.io/utils/ptr"

	"github.com/awslabs/mountpoint-s3-csi-driver/pkg/driver/version"
)

const (
	CreationPolicyKey         = "creationPolicy"
	CreationPolicyUseExisting = "useExisting"
	CreationPolicyCreateNew   = "createNew"
)

type CSIController struct {
	addr   string
	client *s3.Client
	log    logr.Logger
}

func NewCSIController(addr string, client *s3.Client, log logr.Logger) *CSIController {
	return &CSIController{addr: addr, client: client, log: log}
}

func (c *CSIController) Start(ctx context.Context) error {
	l, err := net.Listen("unix", c.addr)
	if err != nil {
		return fmt.Errorf("failed to listen unix socket %q: %w", c.addr, err)
	}

	srv := grpc.NewServer()

	csi.RegisterIdentityServer(srv, c)
	csi.RegisterControllerServer(srv, c)

	c.log.Info("Starting gRPC server for CSI Controller", "addr", l.Addr())
	return srv.Serve(l)
}

//-- csi.IdentityServer

func (c *CSIController) GetPluginInfo(ctx context.Context, req *csi.GetPluginInfoRequest) (*csi.GetPluginInfoResponse, error) {
	return &csi.GetPluginInfoResponse{
		Name:          "s3.csi.aws.com",
		VendorVersion: version.GetVersion().DriverVersion,
	}, nil
}

func (c *CSIController) GetPluginCapabilities(ctx context.Context, req *csi.GetPluginCapabilitiesRequest) (*csi.GetPluginCapabilitiesResponse, error) {
	return &csi.GetPluginCapabilitiesResponse{
		Capabilities: []*csi.PluginCapability{
			{
				Type: &csi.PluginCapability_Service_{
					Service: &csi.PluginCapability_Service{
						Type: csi.PluginCapability_Service_CONTROLLER_SERVICE,
					},
				},
			},
		},
	}, nil
}

func (c *CSIController) Probe(ctx context.Context, req *csi.ProbeRequest) (*csi.ProbeResponse, error) {
	return &csi.ProbeResponse{
		Ready: wrapperspb.Bool(true),
	}, nil
}

//-- csi.ControllerServer

func (c *CSIController) ControllerGetCapabilities(ctx context.Context, req *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {
	return &csi.ControllerGetCapabilitiesResponse{
		Capabilities: []*csi.ControllerServiceCapability{
			{
				Type: &csi.ControllerServiceCapability_Rpc{
					Rpc: &csi.ControllerServiceCapability_RPC{
						Type: csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
					},
				},
			},
		},
	}, nil
}

func (c *CSIController) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	c.log.Info("New CreateVolume request", "req", req)

	switch req.Parameters[CreationPolicyKey] {
	case CreationPolicyCreateNew:
		vol, err := c.createBucket(ctx, req)
		if err != nil {
			return nil, err
		}

		return &csi.CreateVolumeResponse{Volume: vol}, nil
	case CreationPolicyUseExisting:
		return &csi.CreateVolumeResponse{
			Volume: &csi.Volume{
				VolumeId: req.Name,
				VolumeContext: map[string]string{
					"bucketName": req.Parameters["bucketName"],
				},
			},
		}, nil
	default:
		return nil, status.Errorf(codes.InvalidArgument, "%q is required and must be %q or %q", CreationPolicyKey, CreationPolicyCreateNew, CreationPolicyUseExisting)
	}

}

func (c *CSIController) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	return &csi.DeleteVolumeResponse{}, nil
}

func (c *CSIController) ControllerPublishVolume(ctx context.Context, req *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ControllerPublishVolume not implemented")
}

func (c *CSIController) ControllerUnpublishVolume(ctx context.Context, req *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ControllerUnpublishVolume not implemented")
}

func (c *CSIController) ValidateVolumeCapabilities(ctx context.Context, req *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ValidateVolumeCapabilities not implemented")
}

func (c *CSIController) ListVolumes(ctx context.Context, req *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListVolumes not implemented")
}

func (c *CSIController) GetCapacity(ctx context.Context, req *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetCapacity not implemented")
}

func (c *CSIController) CreateSnapshot(ctx context.Context, req *csi.CreateSnapshotRequest) (*csi.CreateSnapshotResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateSnapshot not implemented")
}

func (c *CSIController) DeleteSnapshot(ctx context.Context, req *csi.DeleteSnapshotRequest) (*csi.DeleteSnapshotResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteSnapshot not implemented")
}

func (c *CSIController) ListSnapshots(ctx context.Context, req *csi.ListSnapshotsRequest) (*csi.ListSnapshotsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListSnapshots not implemented")
}

func (c *CSIController) ControllerExpandVolume(ctx context.Context, req *csi.ControllerExpandVolumeRequest) (*csi.ControllerExpandVolumeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ControllerExpandVolume not implemented")
}

func (c *CSIController) ControllerGetVolume(ctx context.Context, req *csi.ControllerGetVolumeRequest) (*csi.ControllerGetVolumeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ControllerGetVolume not implemented")
}

func (c *CSIController) ControllerModifyVolume(ctx context.Context, req *csi.ControllerModifyVolumeRequest) (*csi.ControllerModifyVolumeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ControllerModifyVolume not implemented")
}

//--

func (c *CSIController) createBucket(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.Volume, error) {
	bucketNamePrefix := req.Parameters["bucketNamePrefix"]
	region := req.Parameters["region"]
	name := bucketNamePrefix + req.Name

	out, err := c.client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: ptr.To(name),
		ACL:    types.BucketCannedACLPrivate,
		CreateBucketConfiguration: &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraintEuNorth1,
			Tags: []types.Tag{
				{Key: ptr.To("s3.csi.aws.com/created-by"), Value: ptr.To("mountpoint-s3-csi-driver")},
			},
		},
	}, func(o *s3.Options) {
		o.Region = region
	})
	if err != nil {
		var oe *smithy.OperationError
		if errors.As(err, &oe) {
			c.log.Error(oe.Unwrap(), "Failed to create bucket")
		}

		return nil, fmt.Errorf("failed to create bucket %q: %w", name, err)
	}

	c.log.Info("Successfully created bucket",
		"name", name,
		"location", out.Location,
		"arn", out.BucketArn)

	return &csi.Volume{
		VolumeId: req.Name,
		VolumeContext: map[string]string{
			"bucketName": name,
		},
	}, nil
}
