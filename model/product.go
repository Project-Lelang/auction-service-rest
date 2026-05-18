package model

import (
	"encoding/json"
)

const ProductTableName = "products"

type Product struct {
	Id             string  `db:"id"`
	UserId         string  `db:"user_id"`
	Name           string  `db:"name"`
	Description    *string `db:"description"`
	Condition      string  `db:"condition"`
	CoverImagePath *string `db:"cover_image_path"`
	ImagePaths     *string `db:"image_paths"` // stored as JSON string
	Status         string  `db:"status"`
	Timestamp

	// relations
	User            *User                  `db:"-"`
	StatusHistories []ProductStatusHistory `db:"-"`

	// computed
	CoverImageLink *string  `db:"-"`
	ImageLinks     []string `db:"-"`
}

type ProductQueryOption struct {
	QueryOption

	UserId    *string
	Status    *string
	Condition *string
	Search    *string
}

var _ PrepareOption = &ProductQueryOption{}

// SetDefaultSorts overrides the base default to use created_at DESC.
func (o *ProductQueryOption) SetDefaultSorts() {
	if len(o.Sorts) == 0 {
		o.Sorts = Sorts{{Field: "created_at", Direction: "desc"}}
	}
}

// TranslateSorts prefixes every sort field with the products table alias "p.".
func (o *ProductQueryOption) TranslateSorts() {
	translated := make(Sorts, len(o.Sorts))
	for i, s := range o.Sorts {
		translated[i] = struct{ Field, Direction string }{"p." + s.Field, s.Direction}
	}
	o.Sorts = translated
}

// ParseImagePaths deserialises the JSON image_paths column value into a string slice.
func ParseImagePaths(raw *string) []string {
	if raw == nil || *raw == "" {
		return []string{}
	}
	var urls []string
	_ = json.Unmarshal([]byte(*raw), &urls)
	return urls
}

// MarshalImagePaths serialises a []string to a JSON string for storage.
func MarshalImagePaths(paths []string) string {
	b, _ := json.Marshal(paths)
	return string(b)
}

func (p *Product) TableName() string { return ProductTableName }

func (p *Product) ToMap() map[string]interface{} {
	imagePathsJSON := MarshalImagePaths([]string{})
	if p.ImagePaths != nil {
		imagePathsJSON = *p.ImagePaths
	}
	return map[string]interface{}{
		"id":               p.Id,
		"user_id":          p.UserId,
		"name":             p.Name,
		"description":      p.Description,
		"`condition`":      p.Condition,
		"cover_image_path": p.CoverImagePath,
		"image_paths":      imagePathsJSON,
		"status":           p.Status,
		"created_at":       p.CreatedAt,
		"updated_at":       p.UpdatedAt,
	}
}
