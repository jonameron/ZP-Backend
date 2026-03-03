package services

import (
	"fmt"
	"os"

	"github.com/resend/resend-go/v2"
)

type EmailService struct {
	client    *resend.Client
	fromEmail string
}

func NewEmailService() *EmailService {
	apiKey := os.Getenv("RESEND_API_KEY")
	from := os.Getenv("FROM_EMAIL")
	if from == "" {
		from = "ZeitPass <noreply@zp.11data.ai>"
	}
	return &EmailService{
		client:    resend.NewClient(apiKey),
		fromEmail: from,
	}
}

func (s *EmailService) SendMagicLink(toEmail, token string) error {
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "https://zp.11data.ai"
	}

	link := fmt.Sprintf("%s/auth/verify?token=%s", frontendURL, token)

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head><meta charset="utf-8"></head>
<body style="font-family: 'Inter', sans-serif; background-color: #FAF7F2; padding: 40px 20px;">
  <div style="max-width: 480px; margin: 0 auto; background: #FFFFFF; border-radius: 12px; padding: 40px; box-shadow: 0 2px 8px rgba(0,0,0,0.06);">
    <h1 style="font-family: 'Playfair Display', serif; color: #2D3B2D; font-size: 24px; margin-bottom: 8px;">ZeitPass</h1>
    <p style="color: #6B7B6B; font-size: 14px; margin-bottom: 32px;">Curated experiences in Munich</p>
    <p style="color: #2D3B2D; font-size: 16px; line-height: 1.6;">Click the button below to sign in to your account. This link expires in 15 minutes.</p>
    <a href="%s" style="display: inline-block; background-color: #2D3B2D; color: #FAF7F2; text-decoration: none; padding: 14px 32px; border-radius: 8px; font-size: 16px; font-weight: 500; margin: 24px 0;">Sign in to ZeitPass</a>
    <p style="color: #9CA89C; font-size: 13px; margin-top: 32px;">If you didn't request this link, you can safely ignore this email.</p>
  </div>
</body>
</html>`, link)

	_, err := s.client.Emails.Send(&resend.SendEmailRequest{
		From:    s.fromEmail,
		To:      []string{toEmail},
		Subject: "Your ZeitPass sign-in link",
		Html:    html,
	})
	return err
}

func (s *EmailService) SendBookingConfirmed(toEmail, firstName, eventTitle, date, time string) error {
	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head><meta charset="utf-8"></head>
<body style="font-family: 'Inter', sans-serif; background-color: #FAF7F2; padding: 40px 20px;">
  <div style="max-width: 480px; margin: 0 auto; background: #FFFFFF; border-radius: 12px; padding: 40px; box-shadow: 0 2px 8px rgba(0,0,0,0.06);">
    <h1 style="font-family: 'Playfair Display', serif; color: #2D3B2D; font-size: 24px; margin-bottom: 8px;">ZeitPass</h1>
    <p style="color: #6B7B6B; font-size: 14px; margin-bottom: 32px;">Kuratierte Erlebnisse in München</p>
    <p style="color: #2D3B2D; font-size: 16px; line-height: 1.6;">Hallo %s,</p>
    <p style="color: #2D3B2D; font-size: 16px; line-height: 1.6;">Deine Buchung wurde bestätigt! 🎉</p>
    <div style="background: #FAF7F2; border-radius: 8px; padding: 20px; margin: 24px 0;">
      <p style="color: #2D3B2D; font-size: 16px; font-weight: 600; margin: 0 0 8px;">%s</p>
      <p style="color: #6B7B6B; font-size: 14px; margin: 0;">%s · %s</p>
    </div>
    <p style="color: #2D3B2D; font-size: 14px; line-height: 1.6;">Du findest alle Details in deinem <a href="https://zp.11data.ai/profile" style="color: #2D3B2D; font-weight: 600;">Profil</a>.</p>
    <p style="color: #9CA89C; font-size: 13px; margin-top: 32px;">Wir freuen uns auf dich!</p>
  </div>
</body>
</html>`, firstName, eventTitle, date, time)

	_, err := s.client.Emails.Send(&resend.SendEmailRequest{
		From:    s.fromEmail,
		To:      []string{toEmail},
		Subject: fmt.Sprintf("Buchung bestätigt: %s", eventTitle),
		Html:    html,
	})
	return err
}

func (s *EmailService) SendReminder(toEmail, firstName, eventTitle, date, timeStr, neighbourhood string, daysUntil int) error {
	urgency := "in 7 Tagen"
	subjectPrefix := "Erinnerung"
	if daysUntil <= 1 {
		urgency = "morgen"
		subjectPrefix = "Morgen"
	}

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head><meta charset="utf-8"></head>
<body style="font-family: 'Inter', sans-serif; background-color: #FAF7F2; padding: 40px 20px;">
  <div style="max-width: 480px; margin: 0 auto; background: #FFFFFF; border-radius: 12px; padding: 40px; box-shadow: 0 2px 8px rgba(0,0,0,0.06);">
    <h1 style="font-family: 'Playfair Display', serif; color: #2D3B2D; font-size: 24px; margin-bottom: 8px;">ZeitPass</h1>
    <p style="color: #6B7B6B; font-size: 14px; margin-bottom: 32px;">Kuratierte Erlebnisse in München</p>
    <p style="color: #2D3B2D; font-size: 16px; line-height: 1.6;">Hallo %s,</p>
    <p style="color: #2D3B2D; font-size: 16px; line-height: 1.6;">Dein Erlebnis findet %s statt!</p>
    <div style="background: #FAF7F2; border-radius: 8px; padding: 20px; margin: 24px 0;">
      <p style="color: #2D3B2D; font-size: 16px; font-weight: 600; margin: 0 0 8px;">%s</p>
      <p style="color: #6B7B6B; font-size: 14px; margin: 0;">%s · %s</p>
      <p style="color: #6B7B6B; font-size: 14px; margin: 4px 0 0;">📍 %s</p>
    </div>
    <a href="https://zp.11data.ai/profile" style="display: inline-block; background-color: #2D3B2D; color: #FAF7F2; text-decoration: none; padding: 14px 32px; border-radius: 8px; font-size: 16px; font-weight: 500; margin: 16px 0;">Details ansehen</a>
    <p style="color: #9CA89C; font-size: 13px; margin-top: 32px;">Wir freuen uns auf dich!</p>
  </div>
</body>
</html>`, firstName, urgency, eventTitle, date, timeStr, neighbourhood)

	_, err := s.client.Emails.Send(&resend.SendEmailRequest{
		From:    s.fromEmail,
		To:      []string{toEmail},
		Subject: fmt.Sprintf("%s: %s", subjectPrefix, eventTitle),
		Html:    html,
	})
	return err
}

func (s *EmailService) SendFeedbackRequest(toEmail, firstName, eventTitle string) error {
	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head><meta charset="utf-8"></head>
<body style="font-family: 'Inter', sans-serif; background-color: #FAF7F2; padding: 40px 20px;">
  <div style="max-width: 480px; margin: 0 auto; background: #FFFFFF; border-radius: 12px; padding: 40px; box-shadow: 0 2px 8px rgba(0,0,0,0.06);">
    <h1 style="font-family: 'Playfair Display', serif; color: #2D3B2D; font-size: 24px; margin-bottom: 8px;">ZeitPass</h1>
    <p style="color: #6B7B6B; font-size: 14px; margin-bottom: 32px;">Kuratierte Erlebnisse in München</p>
    <p style="color: #2D3B2D; font-size: 16px; line-height: 1.6;">Hallo %s,</p>
    <p style="color: #2D3B2D; font-size: 16px; line-height: 1.6;">Wie war <strong>%s</strong>? Wir würden uns über dein Feedback freuen!</p>
    <a href="https://zp.11data.ai/profile" style="display: inline-block; background-color: #2D3B2D; color: #FAF7F2; text-decoration: none; padding: 14px 32px; border-radius: 8px; font-size: 16px; font-weight: 500; margin: 24px 0;">Feedback geben</a>
    <p style="color: #9CA89C; font-size: 13px; margin-top: 32px;">Dein Feedback hilft uns, noch bessere Erlebnisse für dich zu kuratieren.</p>
  </div>
</body>
</html>`, firstName, eventTitle)

	_, err := s.client.Emails.Send(&resend.SendEmailRequest{
		From:    s.fromEmail,
		To:      []string{toEmail},
		Subject: fmt.Sprintf("Wie war's? %s", eventTitle),
		Html:    html,
	})
	return err
}

func (s *EmailService) SendMysteryReveal(toEmail, firstName, eventTitle, date, timeStr, neighbourhood string) error {
	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head><meta charset="utf-8"></head>
<body style="font-family: 'Inter', sans-serif; background-color: #FAF7F2; padding: 40px 20px;">
  <div style="max-width: 480px; margin: 0 auto; background: #FFFFFF; border-radius: 12px; padding: 40px; box-shadow: 0 2px 8px rgba(0,0,0,0.06);">
    <h1 style="font-family: 'Playfair Display', serif; color: #2D3B2D; font-size: 24px; margin-bottom: 8px;">ZeitPass</h1>
    <p style="color: #6B7B6B; font-size: 14px; margin-bottom: 32px;">Kuratierte Erlebnisse in München</p>
    <p style="color: #2D3B2D; font-size: 16px; line-height: 1.6;">Hallo %s,</p>
    <p style="color: #2D3B2D; font-size: 16px; line-height: 1.6;">Dein Mystery-Erlebnis wird enthüllt! ✨</p>
    <div style="background: #FAF7F2; border-radius: 8px; padding: 20px; margin: 24px 0;">
      <p style="color: #2D3B2D; font-size: 18px; font-weight: 600; margin: 0 0 8px;">%s</p>
      <p style="color: #6B7B6B; font-size: 14px; margin: 0;">%s · %s</p>
      <p style="color: #6B7B6B; font-size: 14px; margin: 4px 0 0;">📍 %s</p>
    </div>
    <a href="https://zp.11data.ai/profile" style="display: inline-block; background-color: #2D3B2D; color: #FAF7F2; text-decoration: none; padding: 14px 32px; border-radius: 8px; font-size: 16px; font-weight: 500; margin: 16px 0;">Alle Details ansehen</a>
    <p style="color: #9CA89C; font-size: 13px; margin-top: 32px;">Viel Spaß morgen!</p>
  </div>
</body>
</html>`, firstName, eventTitle, date, timeStr, neighbourhood)

	_, err := s.client.Emails.Send(&resend.SendEmailRequest{
		From:    s.fromEmail,
		To:      []string{toEmail},
		Subject: fmt.Sprintf("Mystery enthüllt: %s ✨", eventTitle),
		Html:    html,
	})
	return err
}

func (s *EmailService) SendBookingCancelled(toEmail, firstName, eventTitle string) error {
	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head><meta charset="utf-8"></head>
<body style="font-family: 'Inter', sans-serif; background-color: #FAF7F2; padding: 40px 20px;">
  <div style="max-width: 480px; margin: 0 auto; background: #FFFFFF; border-radius: 12px; padding: 40px; box-shadow: 0 2px 8px rgba(0,0,0,0.06);">
    <h1 style="font-family: 'Playfair Display', serif; color: #2D3B2D; font-size: 24px; margin-bottom: 8px;">ZeitPass</h1>
    <p style="color: #6B7B6B; font-size: 14px; margin-bottom: 32px;">Kuratierte Erlebnisse in München</p>
    <p style="color: #2D3B2D; font-size: 16px; line-height: 1.6;">Hallo %s,</p>
    <p style="color: #2D3B2D; font-size: 16px; line-height: 1.6;">Deine Buchung für <strong>%s</strong> wurde leider storniert.</p>
    <p style="color: #2D3B2D; font-size: 14px; line-height: 1.6; margin-top: 16px;">Entdecke weitere Erlebnisse in deinem <a href="https://zp.11data.ai/events" style="color: #2D3B2D; font-weight: 600;">Event-Feed</a>.</p>
    <p style="color: #9CA89C; font-size: 13px; margin-top: 32px;">Bei Fragen sind wir für dich da.</p>
  </div>
</body>
</html>`, firstName, eventTitle)

	_, err := s.client.Emails.Send(&resend.SendEmailRequest{
		From:    s.fromEmail,
		To:      []string{toEmail},
		Subject: fmt.Sprintf("Buchung storniert: %s", eventTitle),
		Html:    html,
	})
	return err
}
