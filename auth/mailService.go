package auth

import (
	"fmt"

	"gopkg.in/gomail.v2"
)

type MailService struct {
	config MailConfig
}

func NewMailService(config MailConfig) *MailService {
	return &MailService{config: config}
}

func (s *MailService) SendVerificationMail(emailTo string, verificationToken string) error {
	// 1. Формируем ссылку с токеном
	activateURL := fmt.Sprintf(
		"%s:%s/auth/activate?token=%s",
		s.config.appHost,
		s.config.appPort,
		verificationToken,
	)

	m := gomail.NewMessage()
	htmlBody := verificationEmailBody(activateURL)

	m.SetBody("text/html", htmlBody)
	m.SetHeader("From", s.config.mailUserName)
	m.SetHeader("To", emailTo)
	m.SetHeader("Subject", "Подтверждение регистрации")
	m.SetBody("text/plain", fmt.Sprintf("Подтвердите email: %s", activateURL))
	m.AddAlternative("text/html", htmlBody)

	d := gomail.NewDialer(
		s.config.mailHost,
		s.config.mailPort,
		s.config.mailUserName,
		s.config.mailPassword,
	)

	// TODO TODO: при росте нагрузки — outbox-таблица + воркер, rabbit, gorutine?
	// Контекст для прерывания сюда не передать.
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("mailService: SendVerificationMail: %w", err)
	}

	return nil
}

func verificationEmailBody(activateURL string) string {
	return fmt.Sprintf(`
        <h2>Добро пожаловать!</h2>
        <p>Пожалуйста, подтвердите ваш email, перейдя по ссылке:</p>
        <a href="%s">Подтвердить email</a>
        <p>Ссылка действительна 24 часа.</p>
    `, activateURL)
}
