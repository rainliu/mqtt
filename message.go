package mqtt

////////////////////Interface//////////////////////////////

type QOS_TYPE int

const (
	QOS_0 QOS_TYPE = iota
	QOS_1
	QOS_2
)

type Message interface {
	GetQoS() QOS_TYPE
	SetQos(qos QOS_TYPE)

	GetTopic() string
	SetTopic(topic string)

	GetContent() string
	SetContent(content string)
}

////////////////////Implementation////////////////////////
