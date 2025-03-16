package books

type Book struct {
	ID     string `json:"id"`
	Href   string `json:"href,omitempty"`
	Title  string `json:"title"`
	Author string `json:"author"`
}
