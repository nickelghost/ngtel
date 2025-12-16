package ngtel

import (
	"context"
)

// GetGCPLogArgs returns the arguments needed to log traces in GCP format.
func GetGCPLogArgs(ctx context.Context) []any {
	tracePath := GetGCPTracePath(ctx)

	if tracePath == "" {
		return nil
	}

	return []any{"trace", tracePath}
}
