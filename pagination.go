package toolindex

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

type cursorToken struct {
	Offset   int    `json:"offset"`
	Checksum uint64 `json:"checksum"`
}

func encodeCursor(offset int, checksum uint64) (string, error) {
	payload, err := json.Marshal(cursorToken{Offset: offset, Checksum: checksum})
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(payload), nil
}

func decodeCursor(cursor string) (cursorToken, error) {
	if cursor == "" {
		return cursorToken{Offset: 0}, nil
	}
	decoded, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return cursorToken{}, fmt.Errorf("%w: %v", ErrInvalidCursor, err)
	}
	var token cursorToken
	if err := json.Unmarshal(decoded, &token); err != nil {
		return cursorToken{}, fmt.Errorf("%w: %v", ErrInvalidCursor, err)
	}
	if token.Offset < 0 {
		return cursorToken{}, ErrInvalidCursor
	}
	return token, nil
}

func paginateResults[T any](items []T, limit int, cursor string, checksum uint64) ([]T, string, error) {
	token, err := decodeCursor(cursor)
	if err != nil {
		return nil, "", err
	}
	if cursor != "" && token.Checksum != checksum {
		return nil, "", ErrInvalidCursor
	}

	if token.Offset > len(items) {
		return []T{}, "", nil
	}

	end := token.Offset + limit
	if end > len(items) {
		end = len(items)
	}
	page := items[token.Offset:end]

	nextCursor := ""
	if end < len(items) {
		nextCursor, err = encodeCursor(end, checksum)
		if err != nil {
			return nil, "", err
		}
	}

	return page, nextCursor, nil
}
