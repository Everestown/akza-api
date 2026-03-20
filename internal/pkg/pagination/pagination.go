package pagination

import (
	"encoding/base64"
	"encoding/json"
	"time"
)

const defaultLimit = 20

type CursorPage struct {
	Cursor string `form:"cursor"`
	Limit  int    `form:"limit" binding:"omitempty,min=1,max=100"`
}

func (p *CursorPage) GetLimit() int {
	if p.Limit <= 0 { return defaultLimit }
	return p.Limit
}

type PageResult[T any] struct {
	Data    []T    `json:"data"`
	Cursor  string `json:"cursor,omitempty"`
	HasMore bool   `json:"has_more"`
	Limit   int    `json:"limit"`
}

type cursorPayload struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

func EncodeCursor(id int64, createdAt time.Time) string {
	p := cursorPayload{ID: id, CreatedAt: createdAt}
	b, _ := json.Marshal(p)
	return base64.StdEncoding.EncodeToString(b)
}

func DecodeCursor(cursor string) (id int64, createdAt time.Time, err error) {
	if cursor == "" { return 0, time.Time{}, nil }
	b, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil { return 0, time.Time{}, err }
	var p cursorPayload
	if err = json.Unmarshal(b, &p); err != nil { return 0, time.Time{}, err }
	return p.ID, p.CreatedAt, nil
}

func BuildResult[T any](items []T, limit int, cursorFn func(T) string) PageResult[T] {
	hasMore := len(items) > limit
	if hasMore { items = items[:limit] }
	var nextCursor string
	if hasMore && len(items) > 0 {
		nextCursor = cursorFn(items[len(items)-1])
	}
	return PageResult[T]{Data: items, Cursor: nextCursor, HasMore: hasMore, Limit: limit}
}
