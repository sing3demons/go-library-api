package database

type SliceDb[T any] struct {
	Db []T
}

func (s *SliceDb[T]) Add(record T) error {
	s.Db = append(s.Db, record)
	return nil
}

func (s *SliceDb[T]) Update(id string, data T) error {
	return nil
}

func (s *SliceDb[T]) Get(q QueryFunc[T]) ([]T, error) {
	return q()
}

func (s *SliceDb[T]) Delete(id string) error {
	return nil
}
