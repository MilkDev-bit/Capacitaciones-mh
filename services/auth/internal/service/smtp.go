package service

import "net/smtp"

// sendSMTP es un wrapper delgado sobre net/smtp.SendMail para facilitar el test.
func sendSMTP(addr, user, pass, from string, to []string, msg []byte) error {
	auth := smtp.PlainAuth("", user, pass, extractHost(addr))
	return smtp.SendMail(addr, auth, from, to, msg)
}

func extractHost(addr string) string {
	for i, c := range addr {
		if c == ':' {
			return addr[:i]
		}
	}
	return addr
}
