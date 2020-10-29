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

type ActorDaemonConnector interface {
	Daemon
	From() Actor
	To() Daemon
	AsActorDaemonConnector() ActorDaemonConnector
}

type DaemonActorConnector interface {
	Daemon
	From() Daemon
	To() Actor
	AsDaemonActorConnector() DaemonActorConnector
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

func NewActorDaemonConnector(from Actor, to Daemon) ActorDaemonConnector {
	return &actorDaemonConnectorInstance{
		from: from,
		to:   to.Clone(),
	}
}

func NewDaemonActorConnector(from Daemon, to Actor) DaemonActorConnector {
	return &daemonActorConnectorInstance{
		from: from.Clone(),
		to:   to,
	}
}
