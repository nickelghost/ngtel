package ngtel

import "context"

// GCPProjectContextKey represents the key used to get gcp project value inside of the context.
const GCPProjectContextKey = "gcp_project"

// InjectGCPProject injects GCP project value into the context and returns it.
func InjectGCPProject(ctx context.Context, gcpProject string) context.Context {
	return context.WithValue(ctx, GCPProjectContextKey, gcpProject) //nolint:revive,staticcheck
}

// GetGCPProject reads GCP project id from the context.
func GetGCPProject(ctx context.Context) string {
	gcpProject, _ := ctx.Value(GCPProjectContextKey).(string)

	return gcpProject
}
