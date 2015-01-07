package mqtt

////////////////////Interface//////////////////////////////

type SessionState int

type Session interface {
	GetRetransmitTimer() int
	SetRetransmitTimer(retransmitTimer int)

	GetState() SessionState

	Terminate()

	GetAppData() interface{}
	SetAppData(interface{})
}

////////////////////Implementation////////////////////////

type session struct {
}

func (this *session) GetRetransmitTimer() int {
	return 0
}

func (this *session) SetRetransmitTimer(retransmitTimer int) {

}

func (this *session) GetState() SessionState {
	return 0
}

func (this *session) Terminate() {

}

func (this *session) GetAppData() interface{} {
	return nil
}

func (this *session) SetAppData(interface{}) {

}
