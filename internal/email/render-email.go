package email

import (
	"bytes"
	_ "embed"
	"html/template"
)

//go:embed templates/email.html
var verifyEmailHTML string

type VerifyEmailCode struct {
	Code string
}

func BuildEmailLetter(code string) (string, error) {
	tmpl, err := template.New("email").Parse(verifyEmailHTML)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, VerifyEmailCode{Code: code}); err != nil {
		return "", err
	}
	return buf.String(), nil
}
