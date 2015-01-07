package mqtt

////////////////////Interface//////////////////////////////

type Provider interface {
	AddTransport(transport Transport)
	GetTransport(network string) Transport
	GetTransports() []Transport
	RemoveTransport(transport Transport)

	AddListener(listener Listener)
	RemoveListener(listener Listener)

	CreateClientSession(clientTransport ClientTransport) ClientSession
	GetClientSessions() []ClientSession
	DeleteClientSession(clientSession ClientSession)

	CreateServerSession(serverTransport ServerTransport) ServerSession
	GetServerSessions() []ServerSession
	DeleteServerSession(serverSession ServerSession)
}

////////////////////Implementation////////////////////////

type provider struct {
	transports map[Transport]Transport
	listeners  map[Listener]Listener

	clientSessions map[ClientSession]*clientSession
	serverSessions map[ServerSession]*serverSession
}

func newProvider(t Transport) *provider {
	this := &provider{}

	this.clientSessions = make(map[ClientSession]*clientSession)
	this.serverSessions = make(map[ServerSession]*serverSession)

	this.listeners = make(map[Listener]Listener)
	this.transports = make(map[Transport]Transport)

	this.transports[t] = t

	return this
}

func (this *provider) AddTransport(t Transport) {
	this.transports[t] = t
}

func (this *provider) GetTransport(network string) Transport {
	for _, t := range this.transports {
		if t.GetNetwork() == network {
			return t
		}
	}

	return nil
}

func (this *provider) GetTransports() []Transport {
	transports := make([]Transport, len(this.transports))

	l := 0
	for _, value := range this.transports {
		transports[l] = value
		l++
	}

	return transports
}

func (this *provider) RemoveTransport(t Transport) {
	delete(this.transports, t)
}

func (this *provider) AddListener(l Listener) {
	this.listeners[l] = l
}

func (this *provider) RemoveListener(l Listener) {
	delete(this.listeners, l)
}

func (this *provider) CreateClientSession(ct ClientTransport) ClientSession {
	return nil
}

func (this *provider) GetClientSessions() []ClientSession {
	clientSessions := make([]ClientSession, len(this.clientSessions))

	l := 0
	for _, value := range this.clientSessions {
		clientSessions[l] = value
		l++
	}

	return clientSessions
}

func (this *provider) DeleteClientSession(c ClientSession) {
	delete(this.clientSessions, c)
}

func (this *provider) CreateServerSession(st ServerTransport) ServerSession {
	return nil
}

func (this *provider) GetServerSessions() []ServerSession {
	serverSessions := make([]ServerSession, len(this.serverSessions))

	l := 0
	for _, value := range this.serverSessions {
		serverSessions[l] = value
		l++
	}

	return serverSessions
}

func (this *provider) DeleteServerSession(s ServerSession) {
	delete(this.serverSessions, s)
}

func (this *provider) Run() {

}
