package properties

import "context"

const (
	modelKey = "llm.request.model"
)

func SetModelToCtx(ctx context.Context, model string) error {
	return SetProperty(ctx, modelKey, model)
}

func GetModelFromCtx(ctx context.Context) (string, bool) {
	return GetProperty[string](ctx, modelKey)
}
