package mqtt

////////////////////Interface//////////////////////////////

type QOS int

const (
	QOS_ZERO QOS = iota
	QOS_ONE
	QOS_TWO
)

type Message interface {
	GetDup() bool
	SetDup(dup bool)

	GetQos() QOS
	SetQos(qos QOS)

	GetRetain() bool
	SetRetain(retain bool)

	GetTopic() string
	SetTopic(topic string)

	GetContent() string
	SetContent(content string)

	Packetize() PublishPacket
}

////////////////////Implementation////////////////////////
