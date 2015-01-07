package mqtt

////////////////////Interface//////////////////////////////

type ServerSession interface {
	Session

	EnableRetransmissionAlerts()

	Forward(message Message)
}

////////////////////Implementation////////////////////////

type serverSession struct {
	session
}

func (this *serverSession) EnableRetransmissionAlerts() {

}

func (this *serverSession) Forward(message Message) {

}
