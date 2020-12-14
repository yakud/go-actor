package actor

import "context"

func NullReaderDaemon() Daemon {
	return NewDaemon(func(ctx context.Context, in chan interface{}, out chan interface{}, err chan error) error {
		for {
			select {
			case <-ctx.Done():
				return nil
			case _, ok := <-in:
				if !ok {
					return nil
				}
			case _, ok := <-out:
				if !ok {
					return nil
				}
			case _, ok := <-err:
				if !ok {
					return nil
				}
			}
		}
	})
}
