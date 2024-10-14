# Development of Mountpoint for Amazon S3 CSI Driver

## Updating CRDs

We're using [Kubebuilder](https://github.com/kubernetes-sigs/kubebuilder) to define our CRDs and write custom controller for them.

You can update CRD definitions in `api/{version}`, after updating them:
```bash
$ make generate
$ make manifests
```
