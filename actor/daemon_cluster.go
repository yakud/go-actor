package actor

import (
	"context"
	"sync"
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

			d.DisableCloseChannelsOnStop(true)
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

func NewDaemonsClusterWithBroadcast(size int, daemon Daemon) (broadcast Daemon, cluster Daemon) {
	in := make(chan interface{})
	out := make(chan interface{})
	errChan := make(chan error)

	daemons := make([]Daemon, 0, size)
	for id := 0; id < size; id++ {
		d := daemon.Clone()

		d.DisableCloseChannelsOnStop(true)
		d.SetIn(in)
		d.SetOut(out)
		d.SetErr(errChan)

		daemons = append(daemons, d)
	}

	cluster = NewDaemon(func(ctx context.Context, in chan interface{}, out chan interface{}, err chan error) error {
		clusterCtx, clusterCancel := context.WithCancel(ctx)

		wg := &sync.WaitGroup{}
		for _, d := range daemons {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := d.AsDaemonFn()(clusterCtx, in, out, err); err != nil {
					select {
					case <-clusterCtx.Done():
						return
					case errChan <- err:
					}
				}
			}()
		}

		wait := make(chan struct{})
		go func() {
			defer close(wait)
			wg.Wait()
		}()

		for {
			select {
			case <-ctx.Done():
				clusterCancel()
				wg.Wait()
			case <-wait:
				return nil
			}
		}
	}).SetIn(in).SetOut(out).SetErr(errChan)

	broadcast = NewBroadcastDaemon(daemons...)

	return broadcast, cluster
}
