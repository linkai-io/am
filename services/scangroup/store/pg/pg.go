package pg

type Store struct {
}

func New() *Store {
	return &Store{}
}

func (s *Store) Init(config []byte) error {
	return nil
}
