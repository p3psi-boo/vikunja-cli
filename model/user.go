package model

type User struct {
	ID       int64        `json:"id,omitempty"`
	Username string       `json:"username,omitempty"`
	Name     string       `json:"name,omitempty"`
	Email    string       `json:"email,omitempty"`
	Created  NullableTime `json:"created,omitempty"`
	Updated  NullableTime `json:"updated,omitempty"`
}
