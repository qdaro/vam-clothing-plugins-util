package lib

type Variant string

const (
	Success Variant = "success"
	Warning Variant = "warning"
	Error   Variant = "danger"
	Info    Variant = "info"
)

type Note struct {
	Variant Variant `json:"variant"`
	Text    string  `json:"text"`
	Details *string `json:"details,omitempty"`
}

type Message struct {
	Icon  *string `json:"icon,omitempty"`
	Title string  `json:"title"`
	Notes []Note  `json:"notes"`
}

type AppConfig struct {
	OnTop bool `json:"onTop"`
}
