package state

type SubOnStart func() error
type SubOnMessage func(channel string, data []byte) error
