package mqtt

////////////////////Interface//////////////////////////////

type ClientSession interface {
	Session

	Subscribe(topics []string)
	Unsubscribe(topics []string)

	Publish(message Message)
}

////////////////////Implementation////////////////////////
