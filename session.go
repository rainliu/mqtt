package mqtt

////////////////////Interface//////////////////////////////

type SessionState int

const (
	SESSION_STATE_CREATED SessionState = iota
	SESSION_STATE_ACTIVE
	SESSION_STATE_TIMEOUT
	SESSION_STATE_TERMINATED
)

type Session interface {
	GetRetransmitTimer() int
	SetRetransmitTimer(retransmitTimer int)

	GetState() SessionState
	Error() string
	Terminate(err error)

	GetAppData() interface{}
	SetAppData(interface{})
}

////////////////////Implementation////////////////////////

type session struct {
	state           SessionState
	err             error
	ch              chan bool
	appData         interface{}
	retransmitTimer int
}

func (this *session) GetRetransmitTimer() int {
	return this.retransmitTimer
}

func (this *session) SetRetransmitTimer(retransmitTimer int) {
	this.retransmitTimer = retransmitTimer
}

func (this *session) GetState() SessionState {
	return this.state
}

func (this *session) Error() string {
	return this.err.Error()
}

func (this *session) Terminate(err error) {
	this.state = SESSION_STATE_TERMINATED
	this.err = err
	close(this.ch)
}

func (this *session) GetAppData() interface{} {
	return this.appData
}

func (this *session) SetAppData(appData interface{}) {
	this.appData = appData
}
