package actor

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
)

type DaemonFn func(ctx context.Context, in chan interface{}, out chan interface{}, err chan error) error

type Daemon interface {
	AsDaemon() Daemon
	Run(ctx context.Context) (Daemon, error)
	AsDaemonFn() DaemonFn

	SetIn(chan interface{}) Daemon
	SetOut(chan interface{}) Daemon
	SetErr(chan error) Daemon

	In() chan interface{}
	Out() chan interface{}
	Err() chan error

	DisableCloseChannelsOnStop(disabled bool)
	Close()
	Stop()
	Wait()
	IsLaunched() bool
	Clone() Daemon

	ConnectActor(Actor) Daemon
	ConnectDaemon(Daemon) Daemon
}

type daemonPrototype struct {
	fn       DaemonFn
	in       chan interface{}
	out      chan interface{}
	err      chan error
	cancel   context.CancelFunc
	wg       *sync.WaitGroup
	launched uint32

	disabledCloseChannelsOnStop bool
}

func (d *daemonPrototype) Close() {
	d.DisableCloseChannelsOnStop(true)
	close(d.out)
	close(d.in)
	d.out = nil
	d.in = nil
}

func (d *daemonPrototype) ConnectActor(actor Actor) Daemon {
	return NewDaemonActorConnector(d, actor)
}

func (d *daemonPrototype) ConnectDaemon(daemon Daemon) Daemon {
	return NewDaemonsConnector(d, daemon)
}

func (d *daemonPrototype) DisableCloseChannelsOnStop(disabled bool) {
	d.disabledCloseChannelsOnStop = disabled
}

func (d *daemonPrototype) Clone() Daemon {
	d2 := *d
	return &d2
}

func (d *daemonPrototype) IsLaunched() bool {
	return atomic.LoadUint32(&d.launched) == 1
}

func (d *daemonPrototype) Run(ctx context.Context) (Daemon, error) {
	dCopy := *d
	dl := &dCopy

	if dl.in == nil {
		dl.in = make(chan interface{})
	}
	if dl.out == nil {
		dl.out = make(chan interface{})
	}
	if dl.err == nil {
		dl.err = make(chan error)
	}
	if dl.wg == nil {
		dl.wg = &sync.WaitGroup{}
	}

	ctx, dl.cancel = context.WithCancel(ctx)

	var runErr error
	var launched = make(chan struct{})

	dl.wg.Add(1)
	go func() {
		defer func() {
			if !dl.disabledCloseChannelsOnStop && dl.out != nil {
				close(dl.out)
			}
		}()
		defer dl.wg.Done()

		if !atomic.CompareAndSwapUint32(&dl.launched, 0, 1) {
			runErr = fmt.Errorf("daemon already launched")
			close(launched)
			return
		}

		defer func() {
			atomic.CompareAndSwapUint32(&dl.launched, 1, 0)
		}()

		close(launched)

		if runErr = dl.fn(ctx, dl.in, dl.out, dl.err); runErr != nil {
			select {
			case <-ctx.Done():
				return
			case dl.err <- runErr:
			}
		}
	}()

	<-launched

	return dl, runErr
}

func (d *daemonPrototype) SetIn(in chan interface{}) Daemon {
	d.in = in
	return d
}

func (d *daemonPrototype) SetOut(out chan interface{}) Daemon {
	d.out = out
	return d
}

func (d *daemonPrototype) SetErr(err chan error) Daemon {
	d.err = err
	return d
}

func (d *daemonPrototype) In() chan interface{} {
	return d.in
}

func (d *daemonPrototype) Out() chan interface{} {
	return d.out
}

func (d *daemonPrototype) Err() chan error {
	return d.err
}

func (d *daemonPrototype) Stop() {
	if atomic.LoadUint32(&d.launched) == 0 {
		return
	}

	if d.cancel != nil {
		d.cancel()
		d.cancel = nil
	}
}

func (d *daemonPrototype) Wait() {
	if d.wg != nil {
		d.wg.Wait()
	}
}

func (d *daemonPrototype) AsDaemon() Daemon {
	return d
}

func (d *daemonPrototype) AsDaemonFn() DaemonFn {
	return d.fn
}

func NewDaemon(fn DaemonFn) Daemon {
	return &daemonPrototype{
		fn: fn,
	}
}
