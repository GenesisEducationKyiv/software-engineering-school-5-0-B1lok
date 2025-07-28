package stubs

type RecorderStub struct {
	IncActiveSubscriptionsFn func()
	DecActiveSubscriptionsFn func()
}

func NewRecorderStub() *RecorderStub {
	return &RecorderStub{
		IncActiveSubscriptionsFn: nil,
		DecActiveSubscriptionsFn: nil,
	}
}

func (s *RecorderStub) IncActiveSubscriptions() {
	if s.IncActiveSubscriptionsFn != nil {
		s.IncActiveSubscriptionsFn()
	}
}

func (s *RecorderStub) DecActiveSubscriptions() {
	if s.DecActiveSubscriptionsFn != nil {
		s.DecActiveSubscriptionsFn()
	}
}
