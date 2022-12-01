package redis

import (
	"context"
	"errors"
	"github.com/go-redis/redis/v9"
	"leason-toy-web/session"
	"time"
)

type Store struct {
	client     redis.Cmdable
	expiration time.Duration
}

type Session struct {
	client redis.Cmdable
	id     string
}

func NewStore(client redis.Cmdable, expiration time.Duration) *Store {
	return &Store{
		expiration: expiration,
		client:     client,
	}
}

func (s *Store) Generate(ctx context.Context, id string) (session.Session, error) {
	_, err := s.client.HSet(ctx, id, id, id).Result() // 后面的field value仅仅是用于填充命令
	if err != nil {
		return nil, err
	}

	_, err = s.client.Expire(ctx, id, s.expiration).Result()
	if err != nil {
		return nil, err
	}

	return &Session{
		client: s.client,
		id:     id,
	}, nil
}

func (s *Store) Refresh(ctx context.Context, id string) error {
	ok, err := s.client.Expire(ctx, id, s.expiration).Result()
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("session not exist")
	}
	return nil
}

func (s *Store) Remove(ctx context.Context, id string) error {
	_, err := s.client.Del(ctx, id).Result()
	return err
}

func (s *Store) Get(ctx context.Context, id string) (session.Session, error) {
	cnt, err := s.client.Exists(ctx, id).Result()
	if err != nil {
		return nil, err
	}
	if cnt != 1 {
		return nil, errors.New("session not exist")
	}
	return &Session{
		id:     id,
		client: s.client,
	}, nil
}

func (s *Session) Get(ctx context.Context, key string) (any interface{}, err error) {
	return s.client.HGet(ctx, s.id, key).Result()
}

func (s *Session) Set(ctx context.Context, key string, value interface{}) error {
	const lua = `
if redis.call("exists", KEYS[1])
then
	return redis.call("hset", KEYS[1], ARGV[1], ARGV[2])
else
	return -1
end
`
	res, err := s.client.Eval(ctx, lua, []string{s.id}, key, value).Int()
	if err != nil {
		return err
	}
	if res < 0 {
		return errors.New("session not found")
	}
	return nil
}

func (s *Session) ID() string {
	return s.id
}
