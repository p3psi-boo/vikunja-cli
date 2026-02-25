package model

type Project struct {
	ID              int64        `json:"id,omitempty"`
	Title           string       `json:"title,omitempty"`
	Description     string       `json:"description,omitempty"`
	ParentProjectID *int64       `json:"parent_project_id,omitempty"`
	HexColor        string       `json:"hex_color,omitempty"`
	IsFavorite      bool         `json:"is_favorite,omitempty"`
	Created         NullableTime `json:"created,omitempty"`
	Updated         NullableTime `json:"updated,omitempty"`
}

type ProjectCreatePayload struct {
	Title           string `json:"title"`
	Description     string `json:"description,omitempty"`
	ParentProjectID *int64 `json:"parent_project_id,omitempty"`
	HexColor        string `json:"hex_color,omitempty"`
}

type ProjectUpdatePayload struct {
	Title           *string `json:"title,omitempty"`
	Description     *string `json:"description,omitempty"`
	ParentProjectID *int64  `json:"parent_project_id,omitempty"`
	HexColor        *string `json:"hex_color,omitempty"`
	IsFavorite      *bool   `json:"is_favorite,omitempty"`
}
