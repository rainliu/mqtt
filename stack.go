package mqtt

////////////////////Interface//////////////////////////////

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

////////////////////Implementation////////////////////////

var stackSingleton Stack

func GetStack() Stack {
	if stackSingleton == nil {
		return &stack{}
	} else {
		return stackSingleton
	}
}

type stack struct {
	transports []Transport
	providers  []Provider
}

func (this *stack) CreateTransport(network string, address string, port int) Transport {
	return nil
}

func (this *stack) GetTransports() []Transport {
	return this.transports
}

func (this *stack) DeleteTransport(transport Transport) {

}

func (this *stack) CreateProvider(transport Transport) Provider {
	return nil
}

func (this *stack) GetProviders() []Provider {
	return this.providers
}

func (this *stack) DeleteProvider(provider Provider) {

}

func (this *stack) Run() {

}
func (this *stack) Stop() {

}
