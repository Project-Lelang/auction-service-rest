package util

import (
	"fmt"
	"html"
	"log"
	"net/smtp"
	"strings"
	"time"

	"auction-service/global"
)

func SendOtpEmail(to string, otp string, expiresIn time.Duration) error {
	config := global.GetConfig()
	subject := fmt.Sprintf("%s verification code", config.AppName)
	body := verificationEmailHTML(config.AppName, "Verify your email", "Use this code to finish creating your account.", otp, expiresIn)
	fallbackLog := fmt.Sprintf("OTP email delivery disabled; otp for %s is %s", to, otp)
	return sendHTMLEmail(to, subject, body, fallbackLog)
}

func SendForgotPasswordEmail(to string, otp string, expiresIn time.Duration) error {
	config := global.GetConfig()
	subject := fmt.Sprintf("%s password reset code", config.AppName)
	body := verificationEmailHTML(config.AppName, "Reset your password", "Use this code to reset your password.", otp, expiresIn)
	fallbackLog := fmt.Sprintf("Forgot password email delivery disabled; otp for %s is %s", to, otp)
	return sendHTMLEmail(to, subject, body, fallbackLog)
}

func sendHTMLEmail(to string, subject string, htmlBody string, disabledLogMessage string) error {
	config := global.GetConfig()
	emailConfig := config.Email
	if !emailConfig.Enabled || emailConfig.Host == "" || emailConfig.FromAddress == "" {
		if global.IsProduction() {
			return fmt.Errorf("email delivery is not configured")
		}
		log.Print(disabledLogMessage)
		return nil
	}

	fromName := emailConfig.FromName
	if fromName == "" {
		fromName = config.AppName
	}

	message := strings.Join([]string{
		fmt.Sprintf("From: %s <%s>", fromName, emailConfig.FromAddress),
		fmt.Sprintf("To: %s", to),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=UTF-8",
		"",
		htmlBody,
	}, "\r\n")

	address := fmt.Sprintf("%s:%d", emailConfig.Host, emailConfig.Port)
	var auth smtp.Auth
	if emailConfig.Username != "" || emailConfig.Password != "" {
		auth = smtp.PlainAuth("", emailConfig.Username, emailConfig.Password, emailConfig.Host)
	}
	return smtp.SendMail(address, auth, emailConfig.FromAddress, []string{to}, []byte(message))
}

func verificationEmailHTML(appName string, title string, intro string, otp string, expiresIn time.Duration) string {
	safeAppName := html.EscapeString(appName)
	safeTitle := html.EscapeString(title)
	safeIntro := html.EscapeString(intro)
	safeOtp := html.EscapeString(otp)
	minutes := int(expiresIn.Minutes())

	return fmt.Sprintf(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>%s</title>
</head>
<body style="margin:0;background:#f5f3ef;font-family:Arial,Helvetica,sans-serif;color:#202124;">
  <table role="presentation" width="100%%" cellspacing="0" cellpadding="0" style="background:#f5f3ef;padding:32px 16px;">
    <tr>
      <td align="center">
        <table role="presentation" width="100%%" cellspacing="0" cellpadding="0" style="max-width:560px;background:#ffffff;border:1px solid #e4ded4;border-radius:8px;overflow:hidden;">
          <tr>
            <td style="padding:28px 28px 18px;border-bottom:1px solid #ece7df;">
              <div style="font-size:14px;font-weight:700;letter-spacing:.08em;text-transform:uppercase;color:#8a5a28;">%s</div>
              <h1 style="margin:14px 0 0;font-size:24px;line-height:1.3;color:#202124;">%s</h1>
            </td>
          </tr>
          <tr>
            <td style="padding:28px;">
              <p style="margin:0 0 18px;font-size:16px;line-height:1.6;color:#3c4043;">%s</p>
              <div style="margin:24px 0;padding:18px 20px;background:#f9f5ee;border:1px solid #eadfce;border-radius:8px;text-align:center;">
                <div style="font-size:13px;color:#6b6258;margin-bottom:8px;">Your verification code</div>
                <div style="font-size:34px;font-weight:700;letter-spacing:8px;color:#1f1a14;">%s</div>
              </div>
              <p style="margin:0 0 8px;font-size:14px;line-height:1.6;color:#5f6368;">This code expires in %d minutes.</p>
              <p style="margin:0;font-size:14px;line-height:1.6;color:#5f6368;">If you did not request this email, you can safely ignore it.</p>
            </td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>`, safeTitle, safeAppName, safeTitle, safeIntro, safeOtp, minutes)
}
