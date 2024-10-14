// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.

package controller

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	s3csiawscomv1alpha1 "github.com/awslabs/aws-s3-csi-driver/api/v1alpha1"
)

// MountpointClaimReconciler reconciles a MountpointClaim object
type MountpointClaimReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=s3.csi.aws.com,resources=mountpointclaims,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=s3.csi.aws.com,resources=mountpointclaims/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=s3.csi.aws.com,resources=mountpointclaims/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MountpointClaim object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *MountpointClaimReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	mountpointClaim := &s3csiawscomv1alpha1.MountpointClaim{}
	err := r.Get(ctx, req.NamespacedName, mountpointClaim)
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Error(err, "MountpointClaim does not exists, ignoring")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get MountpointClaim")
		return ctrl.Result{}, err
	}

	found := &corev1.Pod{}
	err = r.Get(ctx, types.NamespacedName{Name: mountpointClaim.Name, Namespace: "mount-s3"}, found)
	if err != nil && apierrors.IsNotFound(err) {
		pod, err := r.mountpointPodForClaim(mountpointClaim)
		if err != nil {
			log.Error(err, "Failed to define a new Mountpoint Pod resource for MountpointClaim")

			mountpointClaim.Status.Status = "FailedToCreateMountpointPod"
			if err := r.Status().Update(ctx, mountpointClaim); err != nil {
				log.Error(err, "Failed to update MountpointClaim status")
				return ctrl.Result{}, err
			}

			return ctrl.Result{}, err
		}

		log.Info("Creating a new Pod",
			"Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
		if err = r.Create(ctx, pod); err != nil {
			log.Error(err, "Failed to create new Mountpoint Pod",
				"Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
			return ctrl.Result{}, err
		}

		mountpointClaim.Status.MountpointPodID = ptr.To(string(pod.UID))
		mountpointClaim.Status.Status = "MountpointPodSpawned"
		if err := r.Status().Update(ctx, mountpointClaim); err != nil {
			log.Error(err, "Failed to update MountpointClaim status")
			return ctrl.Result{}, err
		}

		return ctrl.Result{RequeueAfter: time.Minute}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Pod")
		// Let's return the error for the reconciliation be re-trigged again
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MountpointClaimReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&s3csiawscomv1alpha1.MountpointClaim{}).
		Complete(r)
}

func (r *MountpointClaimReconciler) mountpointPodForClaim(claim *s3csiawscomv1alpha1.MountpointClaim) (*corev1.Pod, error) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      claim.Name,
			Namespace: "mount-s3",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Image: "busybox",
				Name:  "mountpoint",
			}},
		},
	}

	if err := ctrl.SetControllerReference(claim, pod, r.Scheme); err != nil {
		return nil, err
	}
	return pod, nil
}
