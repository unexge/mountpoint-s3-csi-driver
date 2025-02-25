package mppod_test

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/awslabs/aws-s3-csi-driver/pkg/podmounter/mppod"
	"github.com/awslabs/aws-s3-csi-driver/pkg/util/testutil/assert"
)

func TestCreatingMountpointPods(t *testing.T) {
	// Mountpoint Pod values
	namespace := "mount-s3"
	mountpointVersion := "1.10.0"
	image := "mp-image:latest"
	imagePullPolicy := corev1.PullAlways
	command := "/bin/aws-s3-csi-mounter"
	priorityClassName := "mount-s3-critical"

	// Test Pod values
	testNode := "test-node"
	testPodUID := "test-pod-uid"
	testVolName := "test-vol"

	csiDriverVersion := "1.12.0"

	creator := mppod.NewCreator(mppod.Config{
		Namespace:         namespace,
		MountpointVersion: mountpointVersion,
		PriorityClassName: priorityClassName,
		Container: mppod.ContainerConfig{
			Image:           image,
			ImagePullPolicy: imagePullPolicy,
			Command:         command,
		},
		CSIDriverVersion: csiDriverVersion,
	})

	verifyDefaultValues := func(mpPod *corev1.Pod) {
		// This is a hash of `testPodUID` + `testVolName`
		assert.Equals(t, "mp-8ef7856a0c7f1d5706bd6af93fdc4bc90b33cf2ceb6769b4afd62586", mpPod.Name)
		assert.Equals(t, namespace, mpPod.Namespace)
		assert.Equals(t, map[string]string{
			mppod.LabelMountpointVersion: mountpointVersion,
			mppod.LabelPodUID:            testPodUID,
			mppod.LabelVolumeName:        testVolName,
			mppod.LabelCSIDriverVersion:  csiDriverVersion,
		}, mpPod.Labels)

		assert.Equals(t, priorityClassName, mpPod.Spec.PriorityClassName)
		assert.Equals(t, corev1.RestartPolicyOnFailure, mpPod.Spec.RestartPolicy)
		assert.Equals(t, []corev1.Volume{
			{
				Name: mppod.CommunicationDirName,
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{},
				},
			},
		}, mpPod.Spec.Volumes)
		assert.Equals(t, &corev1.Affinity{
			NodeAffinity: &corev1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
					NodeSelectorTerms: []corev1.NodeSelectorTerm{
						{
							MatchFields: []corev1.NodeSelectorRequirement{{
								Key:      metav1.ObjectNameField,
								Operator: corev1.NodeSelectorOpIn,
								Values:   []string{testNode},
							}},
						},
					},
				},
			},
		}, mpPod.Spec.Affinity)
		assert.Equals(t, []corev1.Toleration{
			{Operator: corev1.TolerationOpExists},
		}, mpPod.Spec.Tolerations)

		assert.Equals(t, image, mpPod.Spec.Containers[0].Image)
		assert.Equals(t, imagePullPolicy, mpPod.Spec.Containers[0].ImagePullPolicy)
		assert.Equals(t, []string{command}, mpPod.Spec.Containers[0].Command)
		assert.Equals(t, []corev1.VolumeMount{
			{
				Name:      mppod.CommunicationDirName,
				MountPath: "/" + mppod.CommunicationDirName,
			},
		}, mpPod.Spec.Containers[0].VolumeMounts)

	}

	t.Run("Empty PV", func(t *testing.T) {
		mpPod := creator.Create(&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				UID: types.UID(testPodUID),
			},
			Spec: corev1.PodSpec{
				NodeName: testNode,
			},
		}, &corev1.PersistentVolume{
			ObjectMeta: metav1.ObjectMeta{
				Name: testVolName,
			},
		})

		verifyDefaultValues(mpPod)
	})

	t.Run("With ServiceAccountName specified in PV", func(t *testing.T) {
		mpPod := creator.Create(&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				UID: types.UID(testPodUID),
			},
			Spec: corev1.PodSpec{
				NodeName: testNode,
			},
		}, &corev1.PersistentVolume{
			ObjectMeta: metav1.ObjectMeta{
				Name: testVolName,
			},
			Spec: corev1.PersistentVolumeSpec{
				PersistentVolumeSource: corev1.PersistentVolumeSource{
					CSI: &corev1.CSIPersistentVolumeSource{
						VolumeAttributes: map[string]string{
							"mountpointPodServiceAccountName": "mount-s3-sa",
						},
					},
				},
			},
		})

		verifyDefaultValues(mpPod)
		assert.Equals(t, "mount-s3-sa", mpPod.Spec.ServiceAccountName)
	})
}
