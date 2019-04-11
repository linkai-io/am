package browser

type LeaserService interface {
	Acquire() (string, error) // returns port number
	Return(port string) error
	Cleanup() (string, error)
}
