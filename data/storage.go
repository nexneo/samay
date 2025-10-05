package data

import (
	"errors"
	"os"

	"google.golang.org/protobuf/proto"
)

type listingEntry struct {
	Name  string
	IsDir bool
}

type storageBackend interface {
	Exists(path string) (bool, error)
	Read(path string) ([]byte, error)
	Write(path string, data []byte) error
	Delete(path string) error
	List(path string) ([]listingEntry, error)
}

type osBackend struct{}

func (osBackend) Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func (osBackend) Read(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (osBackend) Write(path string, data []byte) error {
	return os.WriteFile(path, data, 0666)
}

func (osBackend) Delete(path string) error {
	return os.Remove(path)
}

func (osBackend) List(path string) ([]listingEntry, error) {
	infos, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	entries := make([]listingEntry, 0, len(infos))
	for _, info := range infos {
		entries = append(entries, listingEntry{Name: info.Name(), IsDir: info.IsDir()})
	}
	return entries, nil
}

type Store struct {
	backend storageBackend
}

// NewStore builds a Store backed by the provided persistence implementation.
func NewStore(backend storageBackend) *Store {
	return &Store{backend: backend}
}

var defaultStore = NewStore(osBackend{})

type persistentEntity interface {
	Persistable
	Destroyable
}

type GenericStore[T persistentEntity] struct {
	store *Store
}

// NewGenericStore wires a Store into a typed helper for CRUD operations on protobuf entities.
func NewGenericStore[T persistentEntity](store *Store) GenericStore[T] {
	return GenericStore[T]{store: store}
}

// Save persists the entity if it has not been written yet, honoring prepare hooks first.
func (g GenericStore[T]) Save(entity T) error {
	return g.store.save(entity)
}

// Update persists the current in-memory state regardless of whether the entity already exists.
func (g GenericStore[T]) Update(entity T) error {
	return g.store.update(entity)
}

// Load hydrates the entity from storage using its Location-derived path.
func (g GenericStore[T]) Load(entity T) error {
	return g.store.load(entity)
}

// Persisted reports whether a stored representation currently exists and returns backend errors.
func (g GenericStore[T]) Persisted(entity T) (bool, error) {
	return g.store.persisted(entity)
}

// Destroy removes the entity's stored representation and runs any cleanup hook.
func (g GenericStore[T]) Destroy(entity T) error {
	return g.store.destroy(entity)
}

func (s *Store) save(entity Persistable) error {
	exists, err := s.backend.Exists(entity.Location())
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	return s.write(entity)
}

func (s *Store) update(entity Persistable) error {
	return s.write(entity)
}

func (s *Store) persisted(entity Persistable) (bool, error) {
	if err := s.load(entity); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *Store) load(entity Persistable) error {
	data, err := s.backend.Read(entity.Location())
	if err != nil {
		return err
	}
	return proto.Unmarshal(data, entity)
}

func (s *Store) loadFromPath(path string, entity Persistable) error {
	data, err := s.backend.Read(path)
	if err != nil {
		return err
	}
	return proto.Unmarshal(data, entity)
}

func (s *Store) destroy(entity Destroyable) error {
	if err := s.backend.Delete(entity.Location()); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return entity.cleanupAfterRemove()
}

func (s *Store) list(path string) ([]listingEntry, error) {
	return s.backend.List(path)
}

func (s *Store) write(entity Persistable) error {
	payload, err := proto.Marshal(entity)
	if err != nil {
		return err
	}
	if err := entity.prepareForSave(); err != nil {
		return err
	}
	return s.backend.Write(entity.Location(), payload)
}
