package model

import "auction-service/data_type"

// Timestamp is embedded in every model to satisfy the BaseModel interface
// and provide consistent created_at / updated_at handling.
type Timestamp struct {
	CreatedAt data_type.DateTime `db:"created_at"`
	UpdatedAt data_type.DateTime `db:"updated_at"`
}

func (t *Timestamp) GetCreatedAt() data_type.DateTime  { return t.CreatedAt }
func (t *Timestamp) SetCreatedAt(v data_type.DateTime) { t.CreatedAt = v }
func (t *Timestamp) GetUpdatedAt() data_type.DateTime  { return t.UpdatedAt }
func (t *Timestamp) SetUpdatedAt(v data_type.DateTime) { t.UpdatedAt = v }
