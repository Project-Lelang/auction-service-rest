package model

import (
	"encoding/json"
)

const ProductTableName = "products"

type Product struct {
	Id            string    `db:"id"`
	UserId        string    `db:"user_id"`
	Name          string    `db:"name"`
	Description   *string   `db:"description"`
	Condition     string    `db:"condition"`
	CoverImageUrl *string   `db:"cover_image_url"`
	ImageUrls     *string   `db:"image_urls"` // stored as JSON string
	Status        string  `db:"status"`
	Timestamp

	// relations
	User            *User                  `db:"-"`
	StatusHistories []ProductStatusHistory `db:"-"`
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

// ParseImageUrls deserialises the JSON image_urls column value into a string slice.
func ParseImageUrls(raw *string) []string {
	if raw == nil || *raw == "" {
		return []string{}
	}
	var urls []string
	_ = json.Unmarshal([]byte(*raw), &urls)
	return urls
}

// MarshalImageUrls serialises a []string to a JSON string for storage.
func MarshalImageUrls(urls []string) string {
	b, _ := json.Marshal(urls)
	return string(b)
}

func (p *Product) TableName() string { return ProductTableName }

func (p *Product) ToMap() map[string]interface{} {
	imageUrlsJSON := MarshalImageUrls([]string{})
	if p.ImageUrls != nil {
		imageUrlsJSON = *p.ImageUrls
	}
	return map[string]interface{}{
		"id":              p.Id,
		"user_id":         p.UserId,
		"name":            p.Name,
		"description":     p.Description,
		"condition":       p.Condition,
		"cover_image_url": p.CoverImageUrl,
		"image_urls":      imageUrlsJSON,
		"status":          p.Status,
		"created_at":      p.CreatedAt,
		"updated_at":      p.UpdatedAt,
	}
}
