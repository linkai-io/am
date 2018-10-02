package redisclient

import (
	"context"
	"errors"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/linkai-io/am/pkg/state"
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

// Subscribe with cancel ability to channels
func (c *Client) Subscribe(ctx context.Context, onStart state.SubOnStart, onMessage state.SubOnMessage, channels ...string) error {
	// A ping is set to the server with this period to test for the health of
	// the connection and server.
	const healthCheckPeriod = time.Minute

	conn, err := c.pool.GetContext(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	psc := redis.PubSubConn{Conn: conn}

	if err := psc.Subscribe(redis.Args{}.AddFlat(channels)...); err != nil {
		return err
	}

	done := make(chan error, 1)

	// Start a goroutine to receive notifications from the server.
	go func() {
		for {
			switch n := psc.Receive().(type) {
			case error:
				done <- n
				return
			case redis.Message:
				if err := onMessage(n.Channel, n.Data); err != nil {
					done <- err
					return
				}
			case redis.Subscription:
				switch n.Count {
				case len(channels):
					// Notify application when all channels are subscribed.
					if err := onStart(); err != nil {
						done <- err
						return
					}
				case 0:
					// Return from the goroutine when all channels are unsubscribed.
					done <- nil
					return
				}
			}
		}
	}()

	ticker := time.NewTicker(healthCheckPeriod)
	defer ticker.Stop()
loop:
	for err == nil {
		select {
		case <-ticker.C:
			// Send ping to test health of connection and server. If
			// corresponding pong is not received, then receive on the
			// connection will timeout and the receive goroutine will exit.
			if err = psc.Ping(""); err != nil {
				break loop
			}
		case <-ctx.Done():
			break loop
		case err := <-done:
			// Return error from the receive goroutine.
			return err
		}
	}

	// Signal the receiving goroutine to exit by unsubscribing from all channels.
	psc.Unsubscribe()

	// Wait for goroutine to complete.
	return <-done
}
