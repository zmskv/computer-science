package interfaces

type SMTPClient interface {
	SendEmail(email string, subject string, body string) error
}
