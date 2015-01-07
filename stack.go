package mqtt

type Stack interface {
	CreateTransport(network string, address string, port int) Transport
	GetTransports() []Transport
	DeleteTransport(transport Transport)

	CreateProvider(transport Transport) Provider
	GetProviders() []Provider
	DeleteProvider(provider Provider)

	Run()
	Stop()
}
