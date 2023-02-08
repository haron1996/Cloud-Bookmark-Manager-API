package mailjet

type Mail struct {
	Email string `json:"to_email"`
	Name  string `json:"to_name"`
	Code  string `json:"code"`
}

func NewMail(email, name, code string) *Mail {
	return &Mail{
		Email: email,
		Name:  name,
		Code:  code,
	}
}
