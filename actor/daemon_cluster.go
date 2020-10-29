package actor

import (
	"context"
	"log"
)

func NewDaemonsCluster(size int, daemon Daemon) Daemon {
	return NewDaemon(func(ctx context.Context, in chan interface{}, out chan interface{}, errChan chan error) error {
		if in == nil {
			in = make(chan interface{})
		}
		if out == nil {
			out = make(chan interface{})
		}
		if errChan == nil {
			errChan = make(chan error)
		}

		daemons := make([]Daemon, 0, size)
		for id := 0; id < size; id++ {
			d := daemon.Clone()

			d.SetIn(in)
			d.SetOut(out)
			d.SetErr(errChan)

			daemonInstance, err := d.Run(ctx)
			if err != nil {
				return err
			}
			daemons = append(daemons, daemonInstance)
		}

		wait := make(chan struct{})
		go func() {
			defer close(wait)
			for _, d := range daemons {
				d.Wait()
			}
		}()

		log.Println("run daemon count:", len(daemons))

		for {
			select {
			case <-ctx.Done():
				for _, d := range daemons {
					d.Stop()
				}
				<-wait
			case <-wait:
				return nil
			}
		}
	})
}
