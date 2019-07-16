package protocauth

import "context"

type Auther interface {
	Authenticate(ctx context.Context) (context.Context, error)
}
