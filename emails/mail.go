package emails

import (
	"log"

	"github.com/kwandapchumba/go-bookmark-manager/util"
)

type Mail struct {
	Domain        string   `json:"mailgun_domain"`
	ApiKey        string   `json:"mailgun_api_key"`
	Sender        string   `json:"sender"`
	Subject       string   `json:"subject"`
	Recipients    []string `json:"recipient"`
	Code          string   `json:"code"`
}

func NewMail(sender, subject, code string, recipients []string) (*Mail, error) {
	config, err := util.LoadConfig("/home/kibet/kibetgo")
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &Mail{
		Domain:        config.MAILGUN_DOMAIN,
		ApiKey:        config.MailgunAPIKey,
		Sender:        sender,
		Subject:       subject,
		Recipients:    recipients,
		Code:          code,
	}, nil
}
