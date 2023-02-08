package mailjet

import (
	"fmt"
	"log"

	"github.com/kwandapchumba/go-bookmark-manager/util"
	"github.com/mailjet/mailjet-apiv3-go/v4"
)

func (m Mail) SendEmailVificationMail() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Println(err)
		return
	}

	client := mailjet.NewMailjetClient(config.MailJetApiKey, config.MailJetSecretKey)

	messagesInfo := []mailjet.InfoMessagesV31{
		{
			From: &mailjet.RecipientV31{
				Email: "support@bookmarkbucket.com",
				Name:  "Team at Bookmark Bucket",
			},
			To: &mailjet.RecipientsV31{
				mailjet.RecipientV31{
					Email: m.Email,
					Name:  m.Name,
				},
			},
			Subject:  "Confirm your email address",
			TextPart: "Confirm your email address",
			HTMLPart: fmt.Sprintf(`<p>Hi %s.</p><p>Your verification code is <b>%s</b>.</p><p>Use this code to confirm your email address.</p><p>Regards.</p><p>The bookmark bucket team.</p>`, m.Name, m.Code),
		},
	}
	messages := mailjet.MessagesV31{Info: messagesInfo}

	_, err = client.SendMailV31(&messages)
	if err != nil {
		log.Fatal(err)
	}

}
