package user

type User struct {
	Id       string `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
	LastName string `json:"lastName,omitempty"`
}
