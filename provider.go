package mqtt

type Provider interface {
	GetStack() Stack

	AddTransport(transport Transport)
	GetTransport(net string) Transport
	GetTransports() []Transport
	RemoveTransport(transport Transport)

	AddListener(listener Listener)
	RemoveListener(listener Listener)

	CreateClientSession(transport Transport) ClientSession
	GetClientSessions() []ClientSession
	DeleteClientSession(clientSession ClientSession)

	CreateServerSession(transport Transport) ServerSession
	GetServerSessions() []ServerSession
	DeleteServerSession(serverSession ServerSession)
}
