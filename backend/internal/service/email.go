package service

import (
	"fmt"
	"log"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// EmailService defines the interface for sending emails.
type EmailService interface {
	// SendVerificationEmail sends an email verification link to the subscriber.
	SendVerificationEmail(to, token string) error
	// SendWelcomeEmail sends a welcome email after verification.
	SendWelcomeEmail(to string) error
	// SendNewPostNotification sends notification about a new blog post.
	SendNewPostNotification(to, postTitle, postURL string) error
}

// emailService implements EmailService using SendGrid.
type emailService struct {
	client   *sendgrid.Client
	fromEmail string
	fromName  string
	baseURL   string
}

// NewEmailService creates a new EmailService instance.
// apiKey: SendGrid API key
// fromEmail: Sender email address (must be verified in SendGrid)
// fromName: Sender name
// baseURL: Base URL of your website (e.g., "https://blog.example.com")
func NewEmailService(apiKey, fromEmail, fromName, baseURL string) EmailService {
	return &emailService{
		client:    sendgrid.NewSendClient(apiKey),
		fromEmail: fromEmail,
		fromName:  fromName,
		baseURL:   baseURL,
	}
}

// SendVerificationEmail sends an email verification link to the subscriber.
func (s *emailService) SendVerificationEmail(to, token string) error {
	verifyURL := fmt.Sprintf("%s/api/newsletter/verify/%s", s.baseURL, token)

	from := mail.NewEmail(s.fromName, s.fromEmail)
	subject := "Xác nhận đăng ký nhận bài viết mới"
	toEmail := mail.NewEmail("", to)

	htmlContent := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; padding: 30px; text-align: center; border-radius: 10px 10px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border-radius: 0 0 10px 10px; }
        .button { display: inline-block; padding: 12px 30px; background: #667eea; color: white; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .footer { text-align: center; margin-top: 20px; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🎉 Chào mừng bạn!</h1>
        </div>
        <div class="content">
            <p>Cảm ơn bạn đã đăng ký nhận bài viết mới từ blog của chúng tôi!</p>
            <p>Để hoàn tất đăng ký, vui lòng click vào nút bên dưới để xác nhận email của bạn:</p>
            <div style="text-align: center;">
                <a href="%s" class="button">Xác nhận email</a>
            </div>
            <p style="font-size: 14px; color: #666;">Hoặc copy link này vào trình duyệt:<br>
            <a href="%s">%s</a></p>
            <p style="margin-top: 30px;">Nếu bạn không đăng ký nhận email này, vui lòng bỏ qua.</p>
        </div>
        <div class="footer">
            <p>© 2026 Personal Blog. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`, verifyURL, verifyURL, verifyURL)

	plainTextContent := fmt.Sprintf(`
Chào mừng bạn!

Cảm ơn bạn đã đăng ký nhận bài viết mới từ blog của chúng tôi!

Để hoàn tất đăng ký, vui lòng truy cập link sau để xác nhận email:
%s

Nếu bạn không đăng ký nhận email này, vui lòng bỏ qua.

---
© 2026 Personal Blog
`, verifyURL)

	message := mail.NewSingleEmail(from, subject, toEmail, plainTextContent, htmlContent)
	response, err := s.client.Send(message)
	if err != nil {
		log.Printf("ERROR: Failed to send verification email to %s: %v", to, err)
		return fmt.Errorf("failed to send email: %w", err)
	}

	if response.StatusCode >= 400 {
		log.Printf("ERROR: SendGrid returned status %d for email to %s: %s", response.StatusCode, to, response.Body)
		return fmt.Errorf("sendgrid error: status %d", response.StatusCode)
	}

	log.Printf("INFO: Verification email sent to %s (status: %d)", to, response.StatusCode)
	return nil
}

// SendWelcomeEmail sends a welcome email after verification.
func (s *emailService) SendWelcomeEmail(to string) error {
	from := mail.NewEmail(s.fromName, s.fromEmail)
	subject := "Đăng ký thành công! 🎉"
	toEmail := mail.NewEmail("", to)

	htmlContent := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 30px; text-align: center; border-radius: 10px 10px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border-radius: 0 0 10px 10px; }
        .footer { text-align: center; margin-top: 20px; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>✅ Email đã được xác nhận!</h1>
        </div>
        <div class="content">
            <p>Chúc mừng! Email của bạn đã được xác nhận thành công.</p>
            <p>Từ giờ, bạn sẽ nhận được thông báo mỗi khi có bài viết mới trên blog.</p>
            <p>Cảm ơn bạn đã theo dõi! 🙏</p>
        </div>
        <div class="footer">
            <p>© 2026 Personal Blog. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`

	plainTextContent := `
Email đã được xác nhận!

Chúc mừng! Email của bạn đã được xác nhận thành công.

Từ giờ, bạn sẽ nhận được thông báo mỗi khi có bài viết mới trên blog.

Cảm ơn bạn đã theo dõi!

---
© 2026 Personal Blog
`

	message := mail.NewSingleEmail(from, subject, toEmail, plainTextContent, htmlContent)
	response, err := s.client.Send(message)
	if err != nil {
		log.Printf("ERROR: Failed to send welcome email to %s: %v", to, err)
		return fmt.Errorf("failed to send email: %w", err)
	}

	if response.StatusCode >= 400 {
		log.Printf("ERROR: SendGrid returned status %d for email to %s: %s", response.StatusCode, to, response.Body)
		return fmt.Errorf("sendgrid error: status %d", response.StatusCode)
	}

	log.Printf("INFO: Welcome email sent to %s (status: %d)", to, response.StatusCode)
	return nil
}

// SendNewPostNotification sends notification about a new blog post.
func (s *emailService) SendNewPostNotification(to, postTitle, postURL string) error {
	from := mail.NewEmail(s.fromName, s.fromEmail)
	subject := fmt.Sprintf("📝 Bài viết mới: %s", postTitle)
	toEmail := mail.NewEmail("", to)

	htmlContent := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; padding: 30px; text-align: center; border-radius: 10px 10px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border-radius: 0 0 10px 10px; }
        .button { display: inline-block; padding: 12px 30px; background: #667eea; color: white; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .footer { text-align: center; margin-top: 20px; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>📝 Bài viết mới!</h1>
        </div>
        <div class="content">
            <h2>%s</h2>
            <p>Chúng tôi vừa đăng một bài viết mới mà bạn có thể quan tâm.</p>
            <div style="text-align: center;">
                <a href="%s" class="button">Đọc bài viết</a>
            </div>
        </div>
        <div class="footer">
            <p>© 2026 Personal Blog. All rights reserved.</p>
            <p><a href="%s/api/newsletter/unsubscribe">Hủy đăng ký</a></p>
        </div>
    </div>
</body>
</html>
`, postTitle, postURL, s.baseURL)

	plainTextContent := fmt.Sprintf(`
Bài viết mới: %s

Chúng tôi vừa đăng một bài viết mới mà bạn có thể quan tâm.

Đọc bài viết tại: %s

---
© 2026 Personal Blog
Hủy đăng ký: %s/api/newsletter/unsubscribe
`, postTitle, postURL, s.baseURL)

	message := mail.NewSingleEmail(from, subject, toEmail, plainTextContent, htmlContent)
	response, err := s.client.Send(message)
	if err != nil {
		log.Printf("ERROR: Failed to send new post notification to %s: %v", to, err)
		return fmt.Errorf("failed to send email: %w", err)
	}

	if response.StatusCode >= 400 {
		log.Printf("ERROR: SendGrid returned status %d for email to %s: %s", response.StatusCode, to, response.Body)
		return fmt.Errorf("sendgrid error: status %d", response.StatusCode)
	}

	log.Printf("INFO: New post notification sent to %s (status: %d)", to, response.StatusCode)
	return nil
}
