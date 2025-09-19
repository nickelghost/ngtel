package ngtel

import (
	"context"
)

func GetGCPLogArgs(ctx context.Context) []any {
	tracePath := GetGCPTracePath(ctx)

	if tracePath == "" {
		return nil
	}

	return []any{"trace", tracePath}
}
