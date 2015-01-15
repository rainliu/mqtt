package mqtt

////////////////////Interface//////////////////////////////

type QOS byte

const (
	QOS_ZERO QOS = iota
	QOS_ONE
	QOS_TWO
	QOS_RESERVED
	QOS_FAILURE QOS = 0x80
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

	GetClientId() string
	SetClientId(clientId string)

	Packetize(packetId uint16) PacketPublish
}

////////////////////Implementation////////////////////////

type message struct {
	dup      bool
	qos      QOS
	retain   bool
	topic    string
	content  string
	clientId string
}

func NewMessage(dup bool, qos QOS, retain bool, topic string, content string) Message {
	return &message{dup: dup,
		qos:     qos,
		retain:  retain,
		topic:   topic,
		content: content}
}

func (this *message) GetDup() bool {
	return this.dup
}
func (this *message) SetDup(dup bool) {
	this.dup = dup
}

func (this *message) GetQos() QOS {
	return this.qos
}
func (this *message) SetQos(qos QOS) {
	this.qos = qos
}

func (this *message) GetRetain() bool {
	return this.retain
}
func (this *message) SetRetain(retain bool) {
	this.retain = retain
}

func (this *message) GetTopic() string {
	return this.topic
}
func (this *message) SetTopic(topic string) {
	this.topic = topic
}

func (this *message) GetContent() string {
	return this.content
}
func (this *message) SetContent(content string) {
	this.content = content
}
func (this *message) GetClientId() string {
	return this.clientId
}
func (this *message) SetClientId(clientId string) {
	this.clientId = clientId
}

func (this *message) Packetize(packetId uint16) PacketPublish {
	pkt := NewPacketPublish()
	pkt.SetPacketId(packetId)
	pkt.SetMessage(this)
	return pkt
}
