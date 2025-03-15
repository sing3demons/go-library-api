package database

type QueryFunc[T any] func() ([]T, error)

type DAO[T any] interface {
	Add(T) error
	Get(QueryFunc[T]) ([]T, error)
	Update(string, T) error
	Delete(string) error
	Find(filter map[string]interface{}) ([]T, error)
}

func AddToDAO[T any](d DAO[T], record T) error {
	return d.Add(record)
}

func QueryDAOWith[T any](d DAO[T], q QueryFunc[T]) ([]T, error) {
	return d.Get(q)
}
