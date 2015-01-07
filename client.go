package mqtt

////////////////////Interface//////////////////////////////

type ClientSession interface {
	Session

	Subscribe(topics []string)
	Unsubscribe(topics []string)

	Publish(message Message)
}

////////////////////Implementation////////////////////////

type clientSession struct {
	session
}

func (this *clientSession) Subscribe(topics []string) {

}

func (this *clientSession) Unsubscribe(topics []string) {

}

func (this *clientSession) Publish(message Message) {

}
