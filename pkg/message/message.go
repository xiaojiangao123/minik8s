package message

type Message struct {
	// The message content
	Type    string `json:"type"`
	URL     string `json:"url"`
	Name    string `json:"name"`
	Content string `json:"content"`
}

const (
	// The message type
	Add    = "Add"
	Delete = "Delete"
	Update = "Update"
	Get    = "Get"
)
