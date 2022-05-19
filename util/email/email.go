package email

import "crypto/tls"

import gomail "gopkg.in/mail.v2"

func SendEmail(toAddress string, subject string, message string) error {
	m := gomail.NewMessage()

	// Set E-Mail sender
	m.SetHeader("From", "gturner.au@gmail.com")

	// Set E-Mail receivers
	m.SetHeader("To", toAddress)

	// Set E-Mail subject
	m.SetHeader("Subject", subject)

	// Set E-Mail body. You can set plain text or html with text/html
	m.SetBody("text/html", message)

	// Settings for SMTP server
	d := gomail.NewDialer("smtp.gmail.com", 587, "gturner.au@gmail.com", "exlxvgauubmdugzy")

	// This is only needed when SSL/TLS certificate is not valid on server.
	// In production this should be set to false.
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	// Now send E-Mail
	if err := d.DialAndSend(m); err != nil {
		return err
	}

	return nil
}
