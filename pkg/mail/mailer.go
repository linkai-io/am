package mail

type Mailer interface {
	Init(config []byte) error
	SendMail(subect, to, html, text string) error
}
