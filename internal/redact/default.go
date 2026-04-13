package redact

// defaultSensitiveKeys is the built-in list of key patterns treated as sensitive.
var defaultSensitiveKeys = []string{
	"password",
	"passwd",
	"secret",
	"token",
	"api_key",
	"apikey",
	"private_key",
	"privatekey",
	"auth",
	"credential",
	"credentials",
	"access_key",
	"accesskey",
	"client_secret",
	"clientsecret",
	"signing_key",
}

// NewDefault returns a Redactor pre-configured with common sensitive key names.
func NewDefault() *Redactor {
	return New(defaultSensitiveKeys)
}
