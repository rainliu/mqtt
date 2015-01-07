package mqtt

type ServerSession interface {
	Session

	EnableRetransmissionAlerts()

	Forward(message Message)
}
