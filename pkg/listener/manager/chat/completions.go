package chat

import "net/http"

func (l *OpenAIChatListener) onCompletionsRequestWithError(writer http.ResponseWriter, request *http.Request) (any, error) {
	return nil, nil
}
