package redisclient

import (
	"context"
	"errors"
	"time"

	"github.com/gomodule/redigo/redis"
)

var (
	// ErrNilConnection pool returned nil for a connection.
	ErrNilConnection = errors.New("pool returned nil for redis connection")
)

// Client wraps access to the ElasticCache/redis server.
type Client struct {
	addr     string
	password string
	pool     *redis.Pool
}

// New creates a new redis client backed by a redis pool.
func New(addr, password string) *Client {
	c := &Client{addr: addr, password: password}
	return c
}

// Init initializes the redis connection pool and runs a test command.
func (c *Client) Init() error {
	c.pool = &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			rc, err := redis.Dial("tcp", c.addr)
			if err != nil {
				return nil, err
			}

			if _, err := rc.Do("AUTH", c.password); err != nil {
				rc.Close()
				return nil, err
			}
			return rc, nil
		},
	}

	conn := c.pool.Get()
	if conn == nil {
		return ErrNilConnection
	}

	defer conn.Close()
	_, err := conn.Do("PING")
	return err
}

// Get a client from the pool and return to caller
func (c *Client) Get() redis.Conn {
	return c.pool.Get()
}

// GetContext a client from the pool with a context and return to caller
func (c *Client) GetContext(ctx context.Context) (redis.Conn, error) {
	return c.pool.GetContext(ctx)
}

// Return the connection (just close it)
func (c *Client) Return(conn redis.Conn) error {
	return conn.Close()
}
