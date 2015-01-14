package mqtt

import "net"

////////////////////Interface//////////////////////////////

type EventType int

const (
	EVENT_CONNECT EventType = iota
	EVENT_PUBLISH
	EVENT_SUBSCRIBE
	EVENT_UNSUBSCRIBE

	EVENT_TIMEOUT
	EVENT_IOEXCEPTION
	EVENT_SESSION_TERMINATED
)

type TimeoutType int

const (
	TIMEOUT_RETRANSMIT TimeoutType = iota
	TIMEOUT_SESSION
)

type Event interface {
	GetEventType() EventType
	GetSession() Session
}

type EventSessionTerminated interface {
	Event

	GetReason() string
}

type EventConnect interface {
	Event

	GetConnectFlags() byte
	GetKeepAlive() uint16
	GetClientId() string
	GetWillTopic() string
	GetWillMessage() string
	GetUserName() string
	GetPassword() []byte
}

type EventPublish interface {
	Event

	GetMessage() Message
}

type EventSubscribe interface {
	Event

	GetPacketId() uint16
	GetSubscribeTopics() []string
	GetQoSs() []QOS
}

type EventUnsubscribe interface {
	Event

	GetUnsubscribeTopics() []string
}

type EventTimeout interface {
	Event

	GetTimeoutType() TimeoutType
}

type EventIOException interface {
	Event

	GetRemoteAddr() net.Addr
}

////////////////////Implementation////////////////////////

type event struct {
	eventType EventType
	session   Session
}

func newEvent(e EventType, s Session) *event {
	return &event{eventType: e, session: s}
}

func (this *event) GetEventType() EventType {
	return this.eventType
}

func (this *event) GetSession() Session {
	return this.session
}

type event_session_terminated struct {
	event

	reason string
}

func newEventSessionTerminated(s Session, r string) *event_session_terminated {
	this := &event_session_terminated{}

	this.eventType = EVENT_SESSION_TERMINATED
	this.session = s
	this.reason = r

	return this
}

func (this *event_session_terminated) GetReason() string {
	return this.reason
}

type event_connect struct {
	event

	connectFlags byte
	keepAlive    uint16
	clientId     string
	willTopic    string
	willMessage  string
	userName     string
	password     []byte
}

func newEventConnect(s Session, p PacketConnect) *event_connect {
	this := &event_connect{}

	this.eventType = EVENT_CONNECT
	this.session = s

	this.connectFlags = p.GetConnectFlags()
	this.keepAlive = p.GetKeepAlive()
	this.clientId = p.GetClientId()
	this.willTopic = p.GetWillTopic()
	this.willMessage = p.GetWillMessage()
	this.userName = p.GetUserName()
	this.password = p.GetPassword()

	return this
}

func (this *event_connect) GetConnectFlags() byte {
	return this.connectFlags
}

func (this *event_connect) GetKeepAlive() uint16 {
	return this.keepAlive
}

func (this *event_connect) GetClientId() string {
	return this.clientId
}

func (this *event_connect) GetWillTopic() string {
	return this.willTopic
}

func (this *event_connect) GetWillMessage() string {
	return this.willMessage
}

func (this *event_connect) GetUserName() string {
	return this.userName
}

func (this *event_connect) GetPassword() []byte {
	return this.password
}

type event_publish struct {
	event

	msg Message
}

func newEventPublish(s Session, m Message) *event_publish {
	this := &event_publish{}

	this.eventType = EVENT_PUBLISH
	this.session = s

	this.msg = m

	return this
}

func (this *event_publish) GetMessage() Message {
	return this.msg
}

type event_subscribe struct {
	event

	packetid uint16
	topics   []string
	qos      []QOS
}

func newEventSubscribe(s Session, p uint16, t []string, q []QOS) *event_subscribe {
	this := &event_subscribe{}

	this.eventType = EVENT_SUBSCRIBE
	this.session = s

	this.packetid = p
	this.topics = t
	this.qos = q

	return this
}

func (this *event_subscribe) GetPacketId() uint16 {
	return this.packetid
}

func (this *event_subscribe) GetSubscribeTopics() []string {
	return this.topics
}

func (this *event_subscribe) GetQoSs() []QOS {
	return this.qos
}

type event_unsubscribe struct {
	event

	topics []string
}

func newEventUnsubscribe(s Session, t []string) *event_unsubscribe {
	this := &event_unsubscribe{}

	this.eventType = EVENT_UNSUBSCRIBE
	this.session = s

	this.topics = t

	return this
}

func (this *event_unsubscribe) GetUnsubscribeTopics() []string {
	return this.topics
}

type event_timeout struct {
	event

	timeoutType TimeoutType
}

func newEventTimeout(s Session, t TimeoutType) *event_timeout {
	this := &event_timeout{}

	this.eventType = EVENT_TIMEOUT
	this.session = s

	this.timeoutType = t

	return this
}

func (this *event_timeout) GetTimeoutType() TimeoutType {
	return this.timeoutType
}

type event_ioexception struct {
	event

	addr net.Addr
}

func newEventIOException(s Session, addr net.Addr) *event_ioexception {
	this := &event_ioexception{}

	this.eventType = EVENT_IOEXCEPTION
	this.session = s
	this.addr = addr

	return this
}

func (this *event_ioexception) GetRemoteAddr() net.Addr {
	return this.addr
}
