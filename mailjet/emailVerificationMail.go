package mailjet

import (
	"fmt"
	"log"
	"os"

	"github.com/mailjet/mailjet-apiv3-go/v4"
)

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

func (m Mail) SendEmailVificationMail() {
	// config, err := util.LoadConfig(".")
	// if err != nil {
	// 	log.Println(err)
	// 	return
	// }

	client := mailjet.NewMailjetClient(os.Getenv("mailJetApiKey"), os.Getenv("mailJetSecretKey"))

	messagesInfo := []mailjet.InfoMessagesV31{
		{
			From: &mailjet.RecipientV31{
				Email: "haron@bookmarkbucket.com",
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

	_, err := client.SendMailV31(&messages)
	if err != nil {
		log.Fatal(err)
	}
}
