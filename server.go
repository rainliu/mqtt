package mqtt

////////////////////Interface//////////////////////////////

type ServerSession interface {
	Session

	EnableRetransmissionAlerts()

	Forward(message Message)
}

////////////////////Implementation////////////////////////
