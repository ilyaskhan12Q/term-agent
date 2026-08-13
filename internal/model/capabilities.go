package model

// ModelCapabilities flags features supported by a provider/model.
type ModelCapabilities struct {
	SupportsNativeToolCalling bool
	SupportsStreaming         bool
	SupportsSystemPrompt      bool
	MaxContextWindow          int
}
