package metadata

import "context"

type requestIDKey struct{}

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, id)
}
func RequestID(ctx context.Context) string { v, _ := ctx.Value(requestIDKey{}).(string); return v }

type ProviderMetadata struct{ values map[string]string }

func New(values map[string]string) ProviderMetadata {
	copy := make(map[string]string, len(values))
	for k, v := range values {
		copy[k] = v
	}
	return ProviderMetadata{values: copy}
}
func (m ProviderMetadata) Get(key string) string { return m.values[key] }
func (m ProviderMetadata) Map() map[string]string {
	copy := make(map[string]string, len(m.values))
	for k, v := range m.values {
		copy[k] = v
	}
	return copy
}
