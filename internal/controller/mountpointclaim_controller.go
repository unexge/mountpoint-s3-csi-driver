// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.

package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
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
	_ = log.FromContext(ctx)

	// TODO(user): your logic here

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MountpointClaimReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&s3csiawscomv1alpha1.MountpointClaim{}).
		Complete(r)
}
