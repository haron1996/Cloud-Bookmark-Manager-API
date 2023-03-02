package mailjet

import (
	"log"

	"github.com/kwandapchumba/go-bookmark-manager/util"
	"github.com/mailjet/mailjet-apiv3-go/v4"
)

type EmailSupportRequest struct {
	FromEmail string `json:"fromEmail"`
	FromName  string `json:"fromName"`
	Subject   string `json:"subject"`
	TextPart  string `json:"textPart"`
	// Recipients []struct {
	// 	Email string `json:"email"`
	// } `json:"recipients"`
}

func (e EmailSupportRequest) NewEmailSupportMail(fromEmail string, fromName string, subject string, message string) *EmailSupportRequest {
	return &EmailSupportRequest{
		FromEmail: fromEmail,
		FromName:  fromName,
		Subject:   subject,
		TextPart:  message,
	}
}

func (m EmailSupportRequest) EmailSupport() error {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Println(err)
		return err
	}
	client := mailjet.NewMailjetClient(config.MailJetApiKey, config.MailJetSecretKey)

	messagesInfo := []mailjet.InfoMessagesV31{
		{
			From: &mailjet.RecipientV31{
				Email: "haron@bookmarkbucket.com",
				Name:  "BookmarkBucket Support",
			},
			To: &mailjet.RecipientsV31{
				mailjet.RecipientV31{
					Email: "kwandapchumba@gmail.com",
					Name:  "Kwandapchumba",
				},
			},
			Subject:  m.Subject,
			TextPart: m.TextPart,
			Sender: &mailjet.RecipientV31{
				Email: m.FromEmail,
				Name:  m.FromName,
			},
		},
	}
	messages := mailjet.MessagesV31{Info: messagesInfo}

	_, err = client.SendMailV31(&messages)
	if err != nil {
		return err
	}

	return nil
}
