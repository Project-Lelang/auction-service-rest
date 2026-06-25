package loader

import (
	"strconv"

	"github.com/graph-gophers/dataloader"
)

func NewDataloader(batchFn dataloader.BatchFunc) dataloader.Loader {
	return *dataloader.NewBatchedLoader(batchFn)
}

func NewDataLoaderP(batchFn dataloader.BatchFunc) *dataloader.Loader {
	loader := NewDataloader(batchFn)
	return &loader
}

func int64Key(id int64) dataloader.Key {
	return dataloader.StringKey(strconv.FormatInt(id, 10))
}

func parseInt64Key(key dataloader.Key) int64 {
	id, _ := strconv.ParseInt(key.String(), 10, 64)
	return id
}
