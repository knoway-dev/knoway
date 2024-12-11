package properties

import "context"

const (
	statusCodeKey   = "http.response.status_code"
	errorMessageKey = "http.response.error_message"
)

func SetStatusCodeToCtx(ctx context.Context, statusCode int) error {
	return SetProperty(ctx, statusCodeKey, statusCode)
}

func GetStatusCodeFromCtx(ctx context.Context) (int, bool) {
	return GetProperty[int](ctx, statusCodeKey)
}

func SetErrorMessageToCtx(ctx context.Context, errorMessage string) error {
	return SetProperty(ctx, errorMessage, errorMessage)
}

func GetErrorMessageFromCtx(ctx context.Context) (string, bool) {
	return GetProperty[string](ctx, errorMessageKey)
}
