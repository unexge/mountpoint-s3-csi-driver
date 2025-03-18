package csicontroller

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/awslabs/aws-s3-csi-driver/pkg/driver/version"
	"github.com/awslabs/aws-s3-csi-driver/pkg/podmounter/mppod"
)

// Name of this component, this needs to be unique as its used as an identifier in logs and metrics.
const Name = "aws-s3-csi-controller"

var log = logf.Log.WithName(Name)

type Controller struct {
	csi.ControllerServer

	Client    client.Client
	Creator   *mppod.Creator
	PodConfig mppod.Config
}

//--- Controller service

func (c *Controller) ControllerGetCapabilities(ctx context.Context, req *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {
	return &csi.ControllerGetCapabilitiesResponse{
		Capabilities: []*csi.ControllerServiceCapability{
			{
				Type: &csi.ControllerServiceCapability_Rpc{
					Rpc: &csi.ControllerServiceCapability_RPC{
						Type: csi.ControllerServiceCapability_RPC_PUBLISH_UNPUBLISH_VOLUME,
					},
				},
			},
		},
	}, nil
}

func (c *Controller) ControllerPublishVolume(ctx context.Context, req *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	mountCap := req.VolumeCapability.GetMount()
	if mountCap == nil {
		return nil, status.Error(codes.InvalidArgument, "Only mount volume supported")
	}

	volCtx := req.VolumeContext
	if volCtx == nil {
		return nil, status.Error(codes.InvalidArgument, "Missing volume context")
	}

	// bucketName := volCtx[volumecontext.BucketName]
	// if bucketName == "" {
	// 	return nil, status.Error(codes.InvalidArgument, "Missing bucket name")
	// }

	// authenticationSource := volCtx[volumecontext.AuthenticationSource]
	// if authenticationSource == "" {
	// 	authenticationSource = credentialprovider.AuthenticationSourceDriver // TODO: Use default auth source.
	// }

	// accessMode := req.VolumeCapability.GetAccessMode().GetMode()

	// // All the fields that makes a volume unique should be here.
	// // Ideally we should hash `mountOptions` and `volumeAttributes` as a whole.
	// volumeHash := sha256.New224()
	// volumeHash.Write([]byte(bucketName))
	// volumeHash.Write([]byte(req.VolumeId))
	// volumeHash.Write([]byte(req.NodeId))
	// volumeHash.Write([]byte(authenticationSource))
	// var readOnlyHash byte
	// if req.Readonly {
	// 	readOnlyHash = 1
	// }
	// volumeHash.Write([]byte{readOnlyHash})
	// binary.Write(volumeHash, binary.LittleEndian, accessMode)

	// var volumeHashSum [sha256.Size224]byte
	// volumeHash.Sum(volumeHashSum[:0])

	// volumeName := fmt.Sprintf("mp-%x", volumeHashSum)

	volCtxStr, err := json.Marshal(volCtx)
	if err != nil {
		return nil, err
	}

	podName := fmt.Sprintf("mp-%x", sha256.Sum224(fmt.Appendf(nil, "%s%s", req.VolumeId, req.NodeId)))

	existingPod := &corev1.Pod{}
	err = c.Client.Get(ctx, types.NamespacedName{Namespace: c.PodConfig.Namespace, Name: podName}, existingPod)
	if err != nil && !apierrors.IsNotFound(err) {
		return nil, err
	}

	if err == nil {
		// TODO: Validate if existing Pod matches with the provided specs?
		return &csi.ControllerPublishVolumeResponse{
			PublishContext: map[string]string{
				"MountpointPodName":      existingPod.Name,
				"MountpointPodNamespace": existingPod.Namespace,
			},
		}, nil
	}

	mpPod := c.Creator.Create(mppod.CreateContext{
		Name:             podName,
		NodeId:           req.NodeId,
		VolumeId:         req.VolumeId,
		VolumeAttributes: volCtx,
	})

	err = c.Client.Create(ctx, mpPod)
	if err != nil {
		log.Error(err, "Failed to create Mountpoint Pod")
		return nil, err
	}

	return &csi.ControllerPublishVolumeResponse{
		PublishContext: map[string]string{
			"MountpointPodName":      mpPod.Name,
			"MountpointPodNamespace": mpPod.Namespace,
			"VolumeContextStr":       string(volCtxStr),
		},
	}, nil
}

func (c *Controller) ControllerUnpublishVolume(ctx context.Context, req *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	log.Info("Unpublishing", "req", req)

	podName := fmt.Sprintf("mp-%x", sha256.Sum224(fmt.Appendf(nil, "%s%s", req.VolumeId, req.NodeId)))

	existingPod := &corev1.Pod{}
	err := c.Client.Get(ctx, types.NamespacedName{Namespace: c.PodConfig.Namespace, Name: podName}, existingPod)
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("Pod already deleted", "podName", podName)
			return &csi.ControllerUnpublishVolumeResponse{}, nil
		}

		return nil, err
	}

	err = c.Client.Delete(ctx, existingPod)
	return &csi.ControllerUnpublishVolumeResponse{}, err
}

//--- Identity service

func (c *Controller) GetPluginInfo(context.Context, *csi.GetPluginInfoRequest) (*csi.GetPluginInfoResponse, error) {
	return &csi.GetPluginInfoResponse{
		Name:          "s3.csi.aws.com",
		VendorVersion: version.GetVersion().DriverVersion,
	}, nil
}

func (c *Controller) GetPluginCapabilities(context.Context, *csi.GetPluginCapabilitiesRequest) (*csi.GetPluginCapabilitiesResponse, error) {
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

func (c *Controller) Probe(context.Context, *csi.ProbeRequest) (*csi.ProbeResponse, error) {
	return &csi.ProbeResponse{
		Ready: wrapperspb.Bool(true),
	}, nil
}
