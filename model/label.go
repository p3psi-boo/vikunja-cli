package model

type Label struct {
	ID          int64        `json:"id,omitempty"`
	Title       string       `json:"title,omitempty"`
	Description string       `json:"description,omitempty"`
	HexColor    string       `json:"hex_color,omitempty"`
	Created     NullableTime `json:"created,omitempty"`
	Updated     NullableTime `json:"updated,omitempty"`
}

type LabelCreatePayload struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	HexColor    string `json:"hex_color,omitempty"`
}

type LabelUpdatePayload struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	HexColor    *string `json:"hex_color,omitempty"`
}
