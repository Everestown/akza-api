package pagination

import (
	"encoding/base64"
	"encoding/json"
	"time"
)

const defaultLimit = 20

// CursorPage holds pagination query parameters.
type CursorPage struct {
	Cursor string `form:"cursor"`
	Limit  int    `form:"limit" binding:"omitempty,min=1,max=100"`
}

func (p *CursorPage) GetLimit() int {
	if p.Limit <= 0 {
		return defaultLimit
	}
	return p.Limit
}

// PageResult is a generic paginated response.
type PageResult[T any] struct {
	Data    []T    `json:"data"`
	Cursor  string `json:"cursor,omitempty"`
	HasMore bool   `json:"has_more"`
	Limit   int    `json:"limit"`
}

// cursorPayload is the internal cursor structure.
type cursorPayload struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

// EncodeCursor encodes a cursor from ID and timestamp.
func EncodeCursor(id string, createdAt time.Time) string {
	p := cursorPayload{ID: id, CreatedAt: createdAt}
	b, _ := json.Marshal(p)
	return base64.StdEncoding.EncodeToString(b)
}

// DecodeCursor decodes an opaque cursor string.
func DecodeCursor(cursor string) (id string, createdAt time.Time, err error) {
	if cursor == "" {
		return "", time.Time{}, nil
	}
	b, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return "", time.Time{}, err
	}
	var p cursorPayload
	if err = json.Unmarshal(b, &p); err != nil {
		return "", time.Time{}, err
	}
	return p.ID, p.CreatedAt, nil
}

// BuildResult builds a PageResult from a slice, trimming to limit and encoding next cursor.
// The slice should contain limit+1 items; the extra item proves HasMore.
func BuildResult[T any](items []T, limit int, cursorFn func(T) string) PageResult[T] {
	hasMore := len(items) > limit
	if hasMore {
		items = items[:limit]
	}
	var nextCursor string
	if hasMore && len(items) > 0 {
		nextCursor = cursorFn(items[len(items)-1])
	}
	return PageResult[T]{
		Data:    items,
		Cursor:  nextCursor,
		HasMore: hasMore,
		Limit:   limit,
	}
}
