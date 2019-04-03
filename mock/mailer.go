package mock

type Mailer struct {
	InitFn      func(config []byte) error
	InitInvoked bool

	SendMailFn      func(subject, to, html, text string) error
	SendMailInvoked bool
}

func (m *Mailer) Init(config []byte) error {
	m.InitInvoked = true
	return m.InitFn(config)
}

func (m *Mailer) SendMail(subject, to, html, text string) error {
	m.SendMailInvoked = true
	return m.SendMailFn(subject, to, html, text)
}
