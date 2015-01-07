package mqtt

type ClientSession interface {
	Session

	Subscribe(topics []string)
	Unsubscribe(topics []string)

	Publish(message Message)
}
