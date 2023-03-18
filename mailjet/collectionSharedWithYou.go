package mailjet

import (
	"fmt"
	"log"
	"strings"

	"github.com/kwandapchumba/go-bookmark-manager/util"
	"github.com/mailjet/mailjet-apiv3-go/v4"
)

type collectonHasBeenSharedWithYou struct {
	EmailSharedWith        string
	NameOfSharer           string
	EmailOfSharer          string
	NameOfCollectionShared string
	IdOfCollectionShared   string
}

func NewCollectionHasBeenSharedWithYou(email_shared_with, name_of_sharer, email_of_sharer, name_of_collection_shared, id_of_collection_shared string) *collectonHasBeenSharedWithYou {
	return &collectonHasBeenSharedWithYou{
		EmailSharedWith:        email_shared_with,
		NameOfSharer:           name_of_sharer,
		EmailOfSharer:          email_of_sharer,
		NameOfCollectionShared: name_of_collection_shared,
		IdOfCollectionShared:   id_of_collection_shared,
	}
}

func (c collectonHasBeenSharedWithYou) SendACollectionHasBeenSharedWithYouMail() {
	config, err := util.LoadConfig(".")
	if err != nil {
		panic(err)
	}

	client := mailjet.NewMailjetClient(config.MailJetApiKey, config.MailJetSecretKey)

	messagesInfo := []mailjet.InfoMessagesV31{
		{
			From: &mailjet.RecipientV31{
				Email: "accounts@linkspace.space",
				Name:  c.NameOfSharer,
			},
			To: &mailjet.RecipientsV31{
				mailjet.RecipientV31{
					Email: c.EmailSharedWith,
					Name:  strings.Split(c.EmailSharedWith, "@")[0],
				},
			},
			Subject:  fmt.Sprintf(`%s has shared a collection with you 👏`, c.NameOfSharer),
			HTMLPart: fmt.Sprintf(`<p>Hey %s</p><p><span style="text-transform: capitalize;">%s</span>(%s) has shared a links collection (%s) with you.</p><a href="http://localhost:5173/appv1/my_links/%s">Click here to view it.</a><p>Regards,</P><p><a href="beta.linkspace.space">Linkspace</a> Team.</p>`, strings.Split(c.EmailSharedWith, "@")[0], c.NameOfSharer, c.EmailOfSharer, c.NameOfCollectionShared, c.IdOfCollectionShared),
		},
	}
	messages := mailjet.MessagesV31{Info: messagesInfo}

	_, err = client.SendMailV31(&messages)
	if err != nil {
		log.Panicf("could not send invite email: %v", err)
	}
}
