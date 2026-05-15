package loader

import "github.com/graph-gophers/dataloader"

func NewDataloader(batchFn dataloader.BatchFunc) dataloader.Loader {
	return *dataloader.NewBatchedLoader(batchFn)
}

func NewDataLoaderP(batchFn dataloader.BatchFunc) *dataloader.Loader {
	loader := NewDataloader(batchFn)
	return &loader
}
