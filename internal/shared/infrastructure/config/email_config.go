package config 

type EmailConfig struct {
	APIKey   string // Resend API key
	From     string // verified sender address, e.g. "onboarding@resend.dev"
	FromName string
}

// LoadEmailConfig loads the Resend (https://resend.com) credentials. The API
// key is required: invitations cannot be delivered without it.
func LoadEmailConfig() (*EmailConfig, error) {
	apiKey, err := requireEnv("RESEND_API_KEY")
	if err != nil {
		return nil, err
	}

	return &EmailConfig{
		APIKey:   apiKey,
		From:     getEnv("EMAIL_FROM", "onboarding@resend.dev"),
		FromName: getEnv("EMAIL_FROM_NAME", "Requisition System"),
	}, nil
}