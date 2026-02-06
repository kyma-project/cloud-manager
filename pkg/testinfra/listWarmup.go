package testinfra

import (
	"context"
	"sync"

	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func WarmupList(ctx context.Context, group string, skipKinds []string, c client.Reader, scheme *runtime.Scheme) {
	arrList := util.AllListsInGroup(group, skipKinds, scheme)

	worker := func(ctx context.Context, jobs <-chan client.ObjectList, wg *sync.WaitGroup) {
		defer func() {
			wg.Done()
		}()

		for list := range jobs {
			_ = c.List(ctx, list)
		}
	}

	jobs := make(chan client.ObjectList)
	var wg sync.WaitGroup
	workerCount := 20
	if len(arrList) < workerCount {
		workerCount = len(arrList)
	}
	if workerCount == 0 {
		return
	}

	wg.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go worker(ctx, jobs, &wg)
	}

	for _, list := range arrList {
		jobs <- list
	}

	close(jobs)
	wg.Wait()
}
