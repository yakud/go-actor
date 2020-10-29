package actor

import (
	"context"
	"fmt"
)

// Connection between two actors
type daemonActorConnectorInstance struct {
	daemonPrototype

	from Daemon
	to   Actor
}

func (d *daemonActorConnectorInstance) Clone() Daemon {
	d2 := *d
	d2.from = d2.from.Clone()

	return &d2
}

func (d *daemonActorConnectorInstance) Run(ctx context.Context) (Daemon, error) {
	d.fn = d.AsDaemonFn()
	return d.daemonPrototype.Run(ctx)
}

func (d *daemonActorConnectorInstance) AsDaemonFn() DaemonFn {
	return func(ctx context.Context, in chan interface{}, out chan interface{}, errChan chan error) error {
		if in == nil {
			in = make(chan interface{})
		}
		if out == nil {
			out = make(chan interface{})
		}
		if errChan == nil {
			errChan = make(chan error)
		}

		d.SetIn(in)
		d.SetOut(out)
		d.SetErr(errChan)

		var err error
		if d.from, err = d.From().Run(ctx); err != nil {
			return err
		}

		for {
			select {
			case <-ctx.Done():
				return nil
			case fromOut, ok := <-d.From().Out():
				if !ok {
					return nil
				}

				if fromOut == nil {
					return nil
				}

				toOut, err := d.To().Call(ctx, fromOut)
				if err != nil {
					return fmt.Errorf("error 'to' call: %w", err)
				}

				if toOut != nil {
					select {
					case <-ctx.Done():
						return nil
					case d.Out() <- toOut:
					}
				}

			}
		}
	}
}

func (d *daemonActorConnectorInstance) SetIn(c chan interface{}) {
	d.in = c
	d.From().SetIn(c)
}

func (d *daemonActorConnectorInstance) SetOut(c chan interface{}) {
	d.out = c
}

func (d *daemonActorConnectorInstance) SetErr(err chan error) {
	d.err = err
	d.From().SetErr(err)
}

func (d *daemonActorConnectorInstance) In() chan interface{} {
	return d.From().In()
}

func (d *daemonActorConnectorInstance) Stop() {
	d.from.Stop()
	d.daemonPrototype.Stop()
}

func (d *daemonActorConnectorInstance) Wait() {
	d.from.Wait()
	d.daemonPrototype.Wait()
}

func (d *daemonActorConnectorInstance) From() Daemon {
	return d.from
}

func (d *daemonActorConnectorInstance) To() Actor {
	return d.to
}

func (d *daemonActorConnectorInstance) AsDaemonActorConnector() DaemonActorConnector {
	return d
}

func (d *daemonActorConnectorInstance) AsDaemon() Daemon {
	return d
}
