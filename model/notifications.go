package model

const NotificationTableName = "notifications"

type Notification struct {
	Id          int64  `db:"id"`
	UserId      int64  `db:"user_id"`
	Title       string `db:"title"`
	Body        string `db:"body"`
	Type        string `db:"type"`
	ReferenceId *int64 `db:"reference_id"`
	IsRead      bool   `db:"is_read"`
	Timestamp
}

func (n *Notification) TableName() string { return NotificationTableName }

func (n *Notification) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":           n.Id,
		"user_id":      n.UserId,
		"title":        n.Title,
		"body":         n.Body,
		"type":         n.Type,
		"reference_id": n.ReferenceId,
		"is_read":      n.IsRead,
		"created_at":   n.CreatedAt,
		"updated_at":   n.UpdatedAt,
	}
}

type NotificationQueryOption struct {
	QueryOption

	UserId *int64
	IsRead *bool
}

var _ PrepareOption = &NotificationQueryOption{}

// SetDefaultSorts overrides the base default to newest first.
func (o *NotificationQueryOption) SetDefaultSorts() {
	if len(o.Sorts) == 0 {
		o.Sorts = Sorts{{Field: "created_at", Direction: "desc"}}
	}
}

// TranslateSorts prefixes every sort field with the notifications table alias "n.".
func (o *NotificationQueryOption) TranslateSorts() {
	translated := make(Sorts, len(o.Sorts))
	for i, s := range o.Sorts {
		translated[i] = struct{ Field, Direction string }{"n." + s.Field, s.Direction}
	}
	o.Sorts = translated
}
