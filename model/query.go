package model

import (
	"fmt"
	"strings"

	"auction-service/constant"
	"auction-service/util"

	"github.com/Masterminds/squirrel"
)

// Sorts is an ordered list of sort items.
// Every entity-specific *FetchSorts dto_request type shares the same underlying
// anonymous-struct layout so that direct type-conversion is possible:
//
//	model.Sorts(request.Sorts)
type Sorts []struct {
	Field     string
	Direction string
}

// PrepareOption is implemented by every entity-specific *QueryOption struct.
type PrepareOption interface {
	GetPage() *int
	GetLimit() *int
	GetSorts() Sorts
	GetIsCount() bool
	SetDefaultSorts()
	TranslateSorts()
}

// QueryOption is the base struct embedded in all entity-specific query option structs.
type QueryOption struct {
	Page    *int
	Limit   *int
	Sorts   Sorts
	IsCount bool
}

func (o QueryOption) GetPage() *int    { return o.Page }
func (o QueryOption) GetLimit() *int   { return o.Limit }
func (o QueryOption) GetSorts() Sorts  { return o.Sorts }
func (o QueryOption) GetIsCount() bool { return o.IsCount }

// SetDefaultSorts sets the default sort to created_at DESC when no sorts are provided.
func (o *QueryOption) SetDefaultSorts() {
	if len(o.Sorts) == 0 {
		o.Sorts = Sorts{{Field: "created_at", Direction: "desc"}}
	}
}

// TranslateSorts is a no-op at the base level; override per entity to add alias prefix.
func (o *QueryOption) TranslateSorts() {}

// NewQueryOptionWithPagination constructs a QueryOption, applying PaginationDefault*
// values when page or limit is nil.
func NewQueryOptionWithPagination(page, limit *int, sorts Sorts) QueryOption {
	if page == nil {
		page = util.Pointer(constant.PaginationDefaultPage)
	}
	if limit == nil {
		limit = util.Pointer(constant.PaginationDefaultLimit)
	}
	return QueryOption{Page: page, Limit: limit, Sorts: sorts}
}

// Prepare applies sorts and pagination to stmt.
// When option.GetIsCount() is true it replaces all columns with COUNT(*).
func Prepare(stmt squirrel.SelectBuilder, option PrepareOption) squirrel.SelectBuilder {
	if option.GetIsCount() {
		return stmt.Column("COUNT(*)")
	}

	option.SetDefaultSorts()
	option.TranslateSorts()

	for _, sort := range option.GetSorts() {
		dir := strings.ToUpper(sort.Direction)
		if dir == "ASC" || dir == "DESC" {
			stmt = stmt.OrderBy(fmt.Sprintf("%s %s", sort.Field, dir))
		}
	}

	page := option.GetPage()
	limit := option.GetLimit()
	if page != nil && limit != nil && *page >= 1 && *limit >= 1 {
		stmt = stmt.Offset(uint64((*page - 1) * *limit))
		stmt = stmt.Limit(uint64(*limit))
	}

	return stmt
}
