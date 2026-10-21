package auth

import (
	"fmt"

	"github.com/gosha5803/go-crud/initializers"
	"gopkg.in/gomail.v2"
)

type MailService struct {
	applicationConfig *initializers.ApplicationConfig
	SMTPConfig *initializers.SMTPConfig
}

func NewMailService(applicationConfig *initializers.ApplicationConfig, SMTPConfig *initializers.SMTPConfig) *MailService {
	return &MailService{applicationConfig: applicationConfig, SMTPConfig: SMTPConfig}
}

func (s *MailService) SendVerificationMail(emailTo string, verificationToken string) error {
	// 1. Формируем ссылку с токеном
	activateURL := fmt.Sprintf(
		"%s:%d/auth/activate?token=%s",
		s.applicationConfig.AppHost,
		s.applicationConfig.AppPort,
		verificationToken,
	)

	m := gomail.NewMessage()
	htmlBody := verificationEmailBody(activateURL)

	m.SetBody("text/html", htmlBody)
	m.SetHeader("From", s.SMTPConfig.SMTPUsername)
	m.SetHeader("To", emailTo)
	m.SetHeader("Subject", "Подтверждение регистрации")
	m.SetBody("text/plain", fmt.Sprintf("Подтвердите email: %s", activateURL))
	m.AddAlternative("text/html", htmlBody)

	d := gomail.NewDialer(
		s.SMTPConfig.SMTPHost,
		s.SMTPConfig.SMTPPort,
		s.SMTPConfig.SMTPUsername,
		s.SMTPConfig.SMTPPassword,
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
