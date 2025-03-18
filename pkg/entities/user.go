package entities

type User struct {
	ID    string `json:"id" bson:"_id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}
