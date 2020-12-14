package actor

type ActorsConnector interface {
	Actor

	From() Actor
	To() Actor
	AsActorsConnector() ActorsConnector
}

type DaemonsConnector interface {
	Daemon

	From() Daemon
	To() Daemon
	AsDaemonsConnector() DaemonsConnector
}

func NewActorsConnector(from Actor, to Actor) ActorsConnector {
	return &actorsConnectorInstance{
		from: from,
		to:   to,
	}
}

func NewDaemonsConnector(from Daemon, to Daemon) DaemonsConnector {
	return &daemonsConnectorInstance{
		from: from.Clone(),
		to:   to.Clone(),
	}
}

func NewActorDaemonConnector(from Actor, to Daemon) DaemonsConnector {
	return NewDaemonsConnector(
		from.AsActorFn().AsDaemon(),
		to,
	)
}

func NewDaemonActorConnector(from Daemon, to Actor) DaemonsConnector {
	return NewDaemonsConnector(
		from,
		to.AsActorFn().AsDaemon(),
	)
}
