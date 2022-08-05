package cilium

import (
	"sync"

	"github.com/layer5io/meshery-adapter-library/adapter"
	"github.com/layer5io/meshery-adapter-library/status"
	mesherykube "github.com/layer5io/meshkit/utils/kubernetes"
)

func (h *Handler) installSampleApp(del bool, namespace string, templates []adapter.Template, kubeconfigs []string) (string, error) {
	st := status.Installing
	if del {
		st = status.Removing
	}
	for _, template := range templates {
		err := h.applyManifest([]byte(template.String()), del, namespace, kubeconfigs)
		if err != nil {
			return st, ErrSampleApp(err)
		}
	}
	return status.Installed, nil
}

func (h *Handler) applyManifest(contents []byte, isDel bool, namespace string, kubeconfigs []string) error {
	var wg sync.WaitGroup
	var errs []error
	var errMx sync.Mutex

	for _, config := range kubeconfigs {
		wg.Add(1)
		go func(config string) {
			defer wg.Done()
			kClient, err := mesherykube.New([]byte(config))
			if err != nil {
				errMx.Lock()
				errs = append(errs, ErrNilClient)
				errMx.Unlock()
				return
			}

			err = kClient.ApplyManifest(contents, mesherykube.ApplyOptions{
				Namespace: namespace,
				Update:    true,
				Delete:    isDel,
			})
			if err != nil {
				errMx.Lock()
				errs = append(errs, err)
				errMx.Unlock()
				return
			}
		}(config)
	}

	wg.Wait()
	if len(errs) == 0 {
		return nil
	}

	return mergeErrors(errs)
}
