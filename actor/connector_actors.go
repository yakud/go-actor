package actor

import "context"

// Connection between two actors
type actorsConnectorInstance struct {
	from Actor
	to   Actor
}

func (c *actorsConnectorInstance) ConnectActor(actor Actor) Actor {
	return NewActorsConnector(c, actor)
}

func (c *actorsConnectorInstance) ConnectDaemon(daemon Daemon) Daemon {
	return NewActorDaemonConnector(c, daemon)
}

func (c *actorsConnectorInstance) AsActorFn() ActorFn {
	return func(ctx context.Context, in interface{}) (out interface{}, err error) {
		fromOut, err := c.From().Call(ctx, in)
		if err != nil {
			return nil, err
		}

		if fromOut == nil {
			return nil, nil
		}

		toOut, err := c.To().Call(ctx, fromOut)
		if err != nil {
			return nil, err
		}

		if toOut == nil {
			return nil, nil
		}
		return toOut, nil
	}
}

func (c *actorsConnectorInstance) AsActor() Actor {
	return c
}

func (c *actorsConnectorInstance) Call(ctx context.Context, in interface{}) (out interface{}, err error) {
	return c.AsActorFn().Call(ctx, in)
}

func (c *actorsConnectorInstance) Connect(from Actor, to Actor) {
	c.from = from
	c.to = to
}

func (c *actorsConnectorInstance) From() Actor {
	return c.from
}

func (c *actorsConnectorInstance) To() Actor {
	return c.to
}

func (c *actorsConnectorInstance) AsActorsConnector() ActorsConnector {
	return c
}

//func (c *actorsConnectorInstance) AsActorsConnector() ActorsConnector {
//	return c
//}
//
//func (c *actorsConnectorInstance) Call(ctx context.Context, in interface{}) (out interface{}, err error) {
//	return c.AsActor().Call(ctx, in)
//}
//
//func (c *actorsConnectorInstance) Run(ctx context.Context) (LaunchedDaemon, error) {
//	return c.AsDaemon().Run(ctx)
//}
//
//func (c *actorsConnectorInstance) AsActor() Actor {
//	return c.AsActorFn().AsActor()
//}
//
//func (c *actorsConnectorInstance) AsDaemon() Daemon {
//	return c.AsActorFn().AsDaemon()
//}
//
//func (c *actorsConnectorInstance) AsActorFn() ActorFn {
//	return func(ctx context.Context, in interface{}) (out interface{}, err error) {
//		fromOut, err := c.From().Call(ctx, in)
//		if err != nil {
//			return nil, err
//		}
//
//		toOut, err := c.To().Call(ctx, fromOut)
//		if err != nil {
//			return nil, err
//		}
//
//		return toOut, nil
//	}
//}
//
//func NewActorsConnection() {
//
//}
