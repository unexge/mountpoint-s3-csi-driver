package custom_testsuites

import (
	"context"
	"fmt"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/kubernetes/pkg/kubelet/events"
	"k8s.io/kubernetes/test/e2e/framework"
	e2eevents "k8s.io/kubernetes/test/e2e/framework/events"
	e2epod "k8s.io/kubernetes/test/e2e/framework/pod"
	storageframework "k8s.io/kubernetes/test/e2e/storage/framework"
	admissionapi "k8s.io/pod-security-admission/api"
)

type s3CSIAccessModeTestSuite struct {
	tsInfo storageframework.TestSuiteInfo
}

func InitS3AccessModeTestSuite() storageframework.TestSuite {
	return &s3CSIAccessModeTestSuite{
		tsInfo: storageframework.TestSuiteInfo{
			Name: "accessmode",
			TestPatterns: []storageframework.TestPattern{
				storageframework.DefaultFsPreprovisionedPV,
			},
		},
	}
}

func (t *s3CSIAccessModeTestSuite) GetTestSuiteInfo() storageframework.TestSuiteInfo {
	return t.tsInfo
}

func (t *s3CSIAccessModeTestSuite) SkipUnsupportedTests(_ storageframework.TestDriver, _ storageframework.TestPattern) {
}

func (t *s3CSIAccessModeTestSuite) DefineTests(driver storageframework.TestDriver, pattern storageframework.TestPattern) {
	type local struct {
		resources []*storageframework.VolumeResource
		config    *storageframework.PerTestConfig
	}
	var (
		l local
	)

	f := framework.NewFrameworkWithCustomTimeouts("accessmode", storageframework.GetDriverTimeouts(driver))
	f.NamespacePodSecurityLevel = admissionapi.LevelRestricted

	cleanup := func(ctx context.Context) {
		var errs []error
		for _, resource := range l.resources {
			errs = append(errs, resource.CleanupResource(ctx))
		}
		framework.ExpectNoError(errors.NewAggregate(errs), "while cleanup resource")
	}
	ginkgo.BeforeEach(func(ctx context.Context) {
		l = local{}
		l.config = driver.PrepareTest(ctx, f)
		ginkgo.DeferCleanup(cleanup)
	})

	validateWriteToVolumeFails := func(ctx context.Context) {
		resource := createVolumeResourceWithAccessMode(ctx, l.config, pattern, v1.ReadOnlyMany)
		l.resources = append(l.resources, resource)
		ginkgo.By("Creating pod with a volume")
		pod := e2epod.MakePod(f.Namespace.Name, nil, []*v1.PersistentVolumeClaim{resource.Pvc}, admissionapi.LevelRestricted, "")
		var err error
		pod, err = createPod(ctx, f.ClientSet, f.Namespace.Name, pod)
		framework.ExpectNoError(err)
		defer func() {
			framework.ExpectNoError(e2epod.DeletePodWithWait(ctx, f.ClientSet, pod))
		}()
		volPath := "/mnt/volume1"
		fileInVol := fmt.Sprintf("%s/file.txt", volPath)
		seed := time.Now().UTC().UnixNano()
		toWrite := 1024 // 1KB
		ginkgo.By("Checking that write to a volume fails")
		checkWriteToPathFails(ctx, f, pod, fileInVol, toWrite, seed)
	}
	ginkgo.It("should fail to write to ReadOnlyMany volume", func(ctx context.Context) {
		validateWriteToVolumeFails(ctx)
	})
	ginkgo.It("S3 express -- should fail to write to ReadOnlyMany volume", func(ctx context.Context) {
		l.config.Prefix = S3ExpressTestIdentifier
		validateWriteToVolumeFails(ctx)
	})

	ginkgo.FContext("passing environment variables to Mountpoint Pod", func() {
		ginkgo.It("injecting incorrect region should cause a failure", func(ctx context.Context) {
			ginkgo.By("Creating a volume with configured env variables")
			resource := createVolumeResourceWithAccessMode(contextWithVolumeAttributes(ctx, map[string]string{
				"mountpointContainerEnv": `
- name: AWS_REGION
  value: us-east-1
`,
			}), l.config, pattern, v1.ReadOnlyMany)
			l.resources = append(l.resources, resource)

			client := f.ClientSet.CoreV1().Pods(f.Namespace.Name)
			pod, err := client.Create(ctx, e2epod.MakePod(f.Namespace.Name, nil, []*v1.PersistentVolumeClaim{resource.Pvc}, admissionapi.LevelRestricted, ""), metav1.CreateOptions{})
			framework.ExpectNoError(err)
			defer func() {
				framework.ExpectNoError(e2epod.DeletePodWithWait(ctx, f.ClientSet, pod))
			}()

			eventSelector := fields.Set{
				"involvedObject.kind":      "Pod",
				"involvedObject.name":      pod.Name,
				"involvedObject.namespace": f.Namespace.Name,
				"reason":                   events.FailedMountVolume,
			}.AsSelector().String()
			framework.Logf("Waiting for FailedMount event: %s", eventSelector)
			err = e2eevents.WaitTimeoutForEvent(ctx, f.ClientSet, f.Namespace.Name, eventSelector, "MountVolume.SetUp failed", 5*time.Minute)
			if err == nil {
				framework.Logf("Got FailedMount event: %s", eventSelector)
			} else {
				framework.Logf("Didn't get FailedMount event: %s", eventSelector)
			}

			pod, err = client.Get(ctx, pod.Name, metav1.GetOptions{})
			framework.ExpectNoError(err)
			gomega.Expect(pod.Status.Phase).To(gomega.Equal(v1.PodPending))

		})

		ginkgo.It("injecting correct region shouldn't break anything", func(ctx context.Context) {
			ginkgo.By("Creating a volume with configured env variables")
			resource := createVolumeResourceWithAccessMode(contextWithVolumeAttributes(ctx, map[string]string{
				"mountpointContainerEnv": `
- name: AWS_REGION
  value: eu-north-1
`,
			}), l.config, pattern, v1.ReadOnlyMany)
			l.resources = append(l.resources, resource)

			pod, err := createPod(ctx, f.ClientSet, f.Namespace.Name, e2epod.MakePod(f.Namespace.Name, nil, []*v1.PersistentVolumeClaim{resource.Pvc}, admissionapi.LevelRestricted, ""))
			framework.ExpectNoError(err)
			defer func() {
				framework.ExpectNoError(e2epod.DeletePodWithWait(ctx, f.ClientSet, pod))
			}()

			ginkgo.By("Checking that writing to a volume works")
			volPath := "/mnt/volume1/file.txt"
			seed := time.Now().UTC().UnixNano()
			toWrite := 1024 // 1KB
			checkWriteToPath(ctx, f, pod, volPath, toWrite, seed)
			checkReadFromPath(ctx, f, pod, volPath, toWrite, seed)
		})
	})

}
