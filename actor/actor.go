package actor

import (
	"context"
	"fmt"
)

var ErrorInputFormat = fmt.Errorf("error input format")

type Actor interface {
	AsActorFn() ActorFn
	AsActor() Actor
	Call(ctx context.Context, in interface{}) (out interface{}, err error)
	ConnectActor(Actor) Actor
	ConnectDaemon(Daemon) Daemon
}

type ActorFn func(ctx context.Context, in interface{}) (out interface{}, err error)

func (fn ActorFn) ConnectActor(actor Actor) Actor {
	return NewActorsConnector(fn, actor)
}

func (fn ActorFn) ConnectDaemon(daemon Daemon) Daemon {
	return NewActorDaemonConnector(fn, daemon)
}

func (fn ActorFn) AsDaemon() Daemon {
	return NewDaemon(fn.AsDaemonFn())
}

func (fn ActorFn) Run(ctx context.Context) (Daemon, error) {
	return fn.AsDaemon().Run(ctx)
}

func (fn ActorFn) AsDaemonFn() DaemonFn {
	return func(ctx context.Context, in chan interface{}, out chan interface{}, errChan chan error) error {
		for {
			select {
			case <-ctx.Done():
				return nil

			case inData, ok := <-in:
				if !ok {
					return nil
				}

				outData, err := fn(ctx, inData)
				if err != nil {
					select {
					case <-ctx.Done():
						return nil
					case errChan <- err:
						continue
					}
				}

				if outData == nil {
					return nil
				}

				select {
				case <-ctx.Done():
					return nil
				case out <- outData:
				}
			}
		}
	}
}

func (fn ActorFn) AsActorFn() ActorFn {
	return fn
}

func (fn ActorFn) AsActor() Actor {
	return fn
}

func (fn ActorFn) Call(ctx context.Context, in interface{}) (out interface{}, err error) {
	return fn(ctx, in)
}

func NewActor(fn ActorFn) Actor {
	return fn
}
