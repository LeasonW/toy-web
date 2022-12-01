package memory

import (
	"context"
	"errors"
	cache "github.com/patrickmn/go-cache"
	"leason-toy-web/session"
	"sync"
	"time"
)

var (
	errorKeyNotFound = errors.New("session: key not found")
)

type Store struct {
	sessions   *cache.Cache
	expiration time.Duration
	mutex      sync.RWMutex
}

func NewStore(expiration time.Duration) *Store {
	return &Store{
		sessions:   cache.New(expiration, time.Second),
		expiration: expiration,
	}
}

type Session struct {
	id     string
	values sync.Map
}

func (s *Session) Get(ctx context.Context, key string) (any interface{}, err error) {
	val, ok := s.values.Load(key)
	if !ok {
		return nil, errorKeyNotFound
	}
	return val, nil
}

func (s *Session) Set(ctx context.Context, key string, value interface{}) error {
	s.values.Store(key, value)
	return nil
}

func (s *Session) ID() string {
	return s.id
}

func (s *Store) Generate(ctx context.Context, id string) (session.Session, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	sess := &Session{id: id}
	s.sessions.Set(id, sess, s.expiration)
	return sess, nil
}

func (s *Store) Refresh(ctx context.Context, id string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	val, ok := s.sessions.Get(id)
	if !ok {
		return errors.New("session: not exist")
	}
	s.sessions.Set(id, val, s.expiration)
	return nil
}

func (s *Store) Remove(ctx context.Context, id string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.sessions.Delete(id)
	return nil
}

func (s *Store) Get(ctx context.Context, id string) (session.Session, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	val, ok := s.sessions.Get(id)
	if !ok {
		return nil, errors.New("session: not exist")
	}
	return val.(*Session), nil
}
