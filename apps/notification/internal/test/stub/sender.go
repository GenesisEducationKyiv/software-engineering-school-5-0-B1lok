package stub

type Sender struct {
	SendFunc func(templateName, to, subject string, data any) error
}

func (s *Sender) Send(templateName, to, subject string, data any) error {
	if s.SendFunc != nil {
		return s.SendFunc(templateName, to, subject, data)
	}
	return nil
}
