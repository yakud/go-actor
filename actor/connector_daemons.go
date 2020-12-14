package actor

import (
	"context"
	"fmt"
)

// Connection between two actors
type daemonsConnectorInstance struct {
	errChan chan error
	from    Daemon
	to      Daemon
}

func (d *daemonsConnectorInstance) Close() {
	d.from.Close()
	d.to.Close()
}

func (d *daemonsConnectorInstance) ConnectActor(actor Actor) Daemon {
	return NewDaemonActorConnector(d, actor)
}

func (d *daemonsConnectorInstance) ConnectDaemon(daemon Daemon) Daemon {
	return NewDaemonsConnector(d, daemon)
}

func (d *daemonsConnectorInstance) DisableCloseChannelsOnStop(disabled bool) {
	d.from.DisableCloseChannelsOnStop(disabled)
	d.to.DisableCloseChannelsOnStop(disabled)
}

func (d *daemonsConnectorInstance) Clone() Daemon {
	d2 := *d
	d2.from = d2.from.Clone()
	d2.to = d2.to.Clone()

	return &d2
}

func (d *daemonsConnectorInstance) Run(ctx context.Context) (Daemon, error) {
	dl := d.Clone()
	err := dl.AsDaemonFn()(ctx, dl.In(), dl.Out(), dl.Err())

	return dl, err
}

func (d *daemonsConnectorInstance) AsDaemonFn() DaemonFn {
	return func(ctx context.Context, in chan interface{}, out chan interface{}, errors chan error) error {
		if d.IsLaunched() {
			return fmt.Errorf("already launched")
		}

		if d.from.In() == nil {
			if in == nil {
				in = make(chan interface{})
			}
			d.SetIn(in)
		}

		if d.to.Out() == nil {
			if out == nil {
				out = make(chan interface{})
			}
			d.SetOut(out)
		}

		if errors == nil {
			errors = make(chan error)
		}
		d.SetErr(errors)

		if d.from.Out() == nil {
			d.from.SetOut(make(chan interface{}))
		}

		//d.from.SetOut(d.to.In())
		d.to.SetIn(d.from.Out())

		var err error

		if !d.from.IsLaunched() {
			d.from, err = d.from.Run(ctx)
			if err != nil {
				return err
			}
		}

		if !d.to.IsLaunched() {
			d.to, err = d.to.Run(ctx)
			if err != nil {
				return err
			}
		}

		return nil
	}
}

func (d *daemonsConnectorInstance) IsLaunched() bool {
	return d.from.IsLaunched() && d.to.IsLaunched()
}

func (d *daemonsConnectorInstance) SetIn(c chan interface{}) Daemon {
	d.from.SetIn(c)
	return d
}

func (d *daemonsConnectorInstance) SetOut(c chan interface{}) Daemon {
	d.to.SetOut(c)
	return d
}

func (d *daemonsConnectorInstance) SetErr(errors chan error) Daemon {
	d.errChan = errors
	d.from.SetErr(d.errChan)
	d.to.SetErr(d.errChan)
	return d
}

func (d *daemonsConnectorInstance) In() chan interface{} {
	return d.from.In()
}

func (d *daemonsConnectorInstance) Out() chan interface{} {
	return d.to.Out()
}

func (d *daemonsConnectorInstance) Err() chan error {
	return d.errChan
}

func (d *daemonsConnectorInstance) Stop() {
	d.from.Stop()
	d.to.Stop()
}

func (d *daemonsConnectorInstance) Wait() {
	d.from.Wait()
	d.to.Wait()
}

func (d *daemonsConnectorInstance) AsDaemon() Daemon {
	return d
}

func (d *daemonsConnectorInstance) From() Daemon {
	return d.from
}

func (d *daemonsConnectorInstance) To() Daemon {
	return d.to
}

func (d *daemonsConnectorInstance) AsDaemonsConnector() DaemonsConnector {
	return d
}
