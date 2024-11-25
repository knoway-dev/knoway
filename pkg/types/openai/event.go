package openai

type Event string

const (
	EventError Event = "error"
)

type ErrorEvent struct {
	Event Event `json:"event"`
	Error Error `json:"error"`
}
