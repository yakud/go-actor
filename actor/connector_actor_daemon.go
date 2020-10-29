package actor

import (
	"context"
	"fmt"
)

type actorDaemonConnectorInstance struct {
	daemonPrototype

	from Actor
	to   Daemon
}

// Sync connector call
func (d *actorDaemonConnectorInstance) AsActorFn() ActorFn {
	return func(ctx context.Context, in interface{}) (out interface{}, err error) {
		var daemon Daemon = d

		if !daemon.IsLaunched() {
			daemon, err = d.Run(ctx)
			if err != nil {
				return nil, err
			}
		}

		select {
		case <-ctx.Done():
			return nil, nil
		case daemon.In() <- in:
		}

		select {
		case <-ctx.Done():
			return nil, nil

		case err = <-daemon.Err():
			return nil, err

		case out, ok := <-daemon.Out():
			if !ok {
				return nil, nil
			}
			return out, nil
		}
	}
}

func (d *actorDaemonConnectorInstance) AsActor() Actor {
	return d
}

// Sync connector call. Send input to daemon and waiting for answer.
// But it may have wrong answer in parallel execution
func (d *actorDaemonConnectorInstance) Call(ctx context.Context, in interface{}) (out interface{}, err error) {
	return d.AsActor().Call(ctx, in)
}

func (d *actorDaemonConnectorInstance) Clone() Daemon {
	d2 := *d
	d2.to = d2.to.Clone()

	return &d2
}

func (d *actorDaemonConnectorInstance) Run(ctx context.Context) (Daemon, error) {
	d.fn = d.AsDaemonFn()
	return d.daemonPrototype.Run(ctx)
}

func (d *actorDaemonConnectorInstance) AsDaemonFn() DaemonFn {
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
		if d.to, err = d.To().Run(ctx); err != nil {
			return err
		}

		for {
			select {
			case <-ctx.Done():
				return nil
			case inData, ok := <-d.In():
				if !ok {
					return nil
				}

				if inData == nil {
					return nil
				}

				fromOut, err := d.From().Call(ctx, inData)
				if err != nil {
					return fmt.Errorf("error 'from' call: %w", err)
				}

				if fromOut == nil {
					return nil
				}

				select {
				case <-ctx.Done():
					return nil
				case d.To().In() <- fromOut:
				}
			}
		}
	}
}

func (d *actorDaemonConnectorInstance) SetIn(in chan interface{}) {
	d.in = in
}

func (d *actorDaemonConnectorInstance) SetOut(out chan interface{}) {
	d.out = out
	d.To().SetOut(out)
}

func (d *actorDaemonConnectorInstance) SetErr(err chan error) {
	d.err = err
	d.to.SetErr(err)
}

func (d *actorDaemonConnectorInstance) From() Actor {
	return d.from
}

func (d *actorDaemonConnectorInstance) To() Daemon {
	return d.to
}

func (d *actorDaemonConnectorInstance) AsActorDaemonConnector() ActorDaemonConnector {
	return d
}

func (d *actorDaemonConnectorInstance) AsDaemon() Daemon {
	return d
}

//func (c *actorDaemonConnectorInstance) AsActorDaemonConnector() ActorDaemonConnector {
//	return c
//}
//
//func (c *actorDaemonConnectorInstance) Call(ctx context.Context, in interface{}) (out interface{}, err error) {
//	_, err = c.AsActor().Call(ctx, in)
//	if err != nil {
//		return nil, err
//	}
//
//	for {
//		select {
//		case <-ctx.Done():
//			return nil, nil
//		case err := <-c.toLaunched.Err():
//			if err != nil {
//				return nil, err
//			}
//		case toOut, ok := <-c.toLaunched.Out():
//			if !ok {
//				return nil, nil
//			}
//
//			if toOut == nil {
//				continue
//			}
//
//			return toOut, nil
//		}
//	}
//}
//
//func (c *actorDaemonConnectorInstance) Run(ctx context.Context) (LaunchedDaemon, error) {
//	var err error
//
//	cDaemon := &(*c)
//
//	if err = cDaemon.launchDaemons(ctx); err != nil {
//		return nil, fmt.Errorf("error launch daemons")
//	}
//
//	daemon := cDaemon.AsDaemon()
//	daemonInstance, ok := daemon.(*daemonPrototype)
//	if !ok {
//		return nil, fmt.Errorf("error cast daemon to *daemonPrototype")
//	}
//
//	daemonInstance.out = c.toLaunched.out
//	daemonInstance.err = c.toLaunched.err
//
//	daemonLaunched, err := daemonInstance.Run(ctx)
//	if err != nil {
//		return nil, err
//	}
//
//	return daemonLaunched, nil
//}
//
//func (c *actorDaemonConnectorInstance) Stop() {
//	if c.toLaunched != nil {
//		if c.toLaunched.cancel != nil {
//			c.toLaunched.cancel()
//		}
//		c.toLaunched.Wait()
//
//		close(c.toLaunched.in)
//		c.toLaunched.in = nil
//
//		close(c.toLaunched.out)
//		c.toLaunched.out = nil
//
//		close(c.toLaunched.err)
//		c.toLaunched.err = nil
//
//		c.toLaunched = nil
//	}
//}
//
//func (c *actorDaemonConnectorInstance) AsActor() Actor {
//	return c.AsActorFn().AsActor()
//}
//
//func (c *actorDaemonConnectorInstance) AsDaemon() Daemon {
//	return c.AsActorFn().AsDaemon()
//}
//
//func (c *actorDaemonConnectorInstance) AsActorFn() ActorFn {
//	return func(ctx context.Context, in interface{}) (out interface{}, err error) {
//		if err := c.launchDaemons(ctx); err != nil {
//			return nil, fmt.Errorf("error launch daemons: %w", err)
//		}
//
//		if c.toLaunched == nil {
//			return nil, fmt.Errorf("'to' daemon isn't launched")
//		}
//
//		fromOut, err := c.From().Call(ctx, in)
//		if err != nil {
//			return nil, fmt.Errorf("from call error: %w", err)
//		}
//
//		select {
//		case <-ctx.Done():
//			return nil, nil
//		case err := <-c.toLaunched.Err():
//			if err != nil {
//				return nil, fmt.Errorf("actor 'to' error: %w", err)
//			}
//		case c.toLaunched.In() <- fromOut:
//		}
//
//		return nil, nil
//	}
//}
//
//func (c *actorDaemonConnectorInstance) launchDaemons(ctx context.Context) error {
//	var ok bool
//
//	if c.toLaunched == nil {
//		toDaemon, err := c.To().AsActorFn().AsDaemon().Run(ctx)
//		if err != nil {
//			return err
//		}
//
//		c.toLaunched, ok = toDaemon.(*daemonPrototype)
//		if !ok {
//			return fmt.Errorf("error cast toDaemon to *daemonPrototype")
//		}
//	}
//
//	return nil
//}
