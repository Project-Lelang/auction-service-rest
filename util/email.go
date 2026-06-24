package util

import (
	"fmt"
	"log"
	"net/smtp"
	"strings"
	"time"

	"auction-service/global"
)

func SendOtpEmail(to string, otp string, expiresIn time.Duration) error {
	config := global.GetConfig()
	emailConfig := config.Email
	if !emailConfig.Enabled || emailConfig.Host == "" || emailConfig.FromAddress == "" {
		if global.IsProduction() {
			return fmt.Errorf("email delivery is not configured")
		}
		log.Printf("OTP email delivery disabled; otp for %s is %s", to, otp)
		return nil
	}

	fromName := emailConfig.FromName
	if fromName == "" {
		fromName = config.AppName
	}

	subject := fmt.Sprintf("%s verification code", config.AppName)
	body := fmt.Sprintf(
		"Your %s verification code is %s.\n\nThis code expires in %.0f minutes.\nIf you did not request this code, you can ignore this email.",
		config.AppName,
		otp,
		expiresIn.Minutes(),
	)
	message := strings.Join([]string{
		fmt.Sprintf("From: %s <%s>", fromName, emailConfig.FromAddress),
		fmt.Sprintf("To: %s", to),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}, "\r\n")

	address := fmt.Sprintf("%s:%d", emailConfig.Host, emailConfig.Port)
	var auth smtp.Auth
	if emailConfig.Username != "" || emailConfig.Password != "" {
		auth = smtp.PlainAuth("", emailConfig.Username, emailConfig.Password, emailConfig.Host)
	}
	return smtp.SendMail(address, auth, emailConfig.FromAddress, []string{to}, []byte(message))
}
