package infrastructure

import (
	"auction-service/global"

	storage_go "github.com/supabase-community/storage-go"
)

func NewSupabaseStorageClient(cfg *global.SupabaseConfig) *storage_go.Client {
	if cfg == nil {
		return nil
	}
	return storage_go.NewClient(cfg.URL+"/storage/v1", cfg.ServiceKey, nil)
}
