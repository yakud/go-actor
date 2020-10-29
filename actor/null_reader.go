package actor

import "context"

func NullReaderDaemon() Daemon {
	return NewDaemon(func(ctx context.Context, in chan interface{}, out chan interface{}, err chan error) error {
		for {
			select {
			case <-ctx.Done():
				return nil
			case <-in:
			case <-out:
			case <-err:
			}
		}
	})
}
