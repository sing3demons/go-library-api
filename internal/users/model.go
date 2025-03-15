package users

type User struct {
	ID    string `json:"id" bson:"_id"`
	Href  string `json:"href,omitempty" bson:"-"`
	Name  string `json:"name"`
	Email string `json:"email"`
}
