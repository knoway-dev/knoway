package object

type ImageGenerationsUsageImage interface {
	GetWidth() uint64
	GetHeight() uint64
	GetStyle() string
	GetQuality() string
}

type LLMImagesUsage interface {
	LLMUsage

	GetOutputImages() []ImageGenerationsUsageImage
}

func AsLLMImagesUsage(u LLMUsage) (LLMImagesUsage, bool) {
	t, ok := u.(LLMImagesUsage)
	return t, ok
}
