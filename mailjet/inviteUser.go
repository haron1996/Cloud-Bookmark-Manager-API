package mailjet

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/kwandapchumba/go-bookmark-manager/util"
	"github.com/mailjet/mailjet-apiv3-go/v4"
)

type inviteUserMail struct {
	InvitedEmail   string    `json:"invited_email"`
	InvitedByName  string    `json:"invited_by_name"`
	InvitedByMail  string    `json:"invited_by_mail"`
	Token          string    `json:"token"`
	Expiry         time.Time `json:"expiry"`
	CollectionName string    `json:"collection_name"`
}

func NewInviteUserMail(inviteeName, inviteeMail, invited, token, collectionName string, expiry time.Time) *inviteUserMail {
	return &inviteUserMail{
		InvitedEmail:   invited,
		InvitedByName:  inviteeName,
		InvitedByMail:  inviteeMail,
		Token:          token,
		Expiry:         expiry,
		CollectionName: collectionName,
	}
}

func (i inviteUserMail) SendInviteUserEmail() {
	config, err := util.LoadConfig(".")
	if err != nil {
		panic(err)
	}

	client := mailjet.NewMailjetClient(config.MailJetApiKey, config.MailJetSecretKey)

	messagesInfo := []mailjet.InfoMessagesV31{
		{
			From: &mailjet.RecipientV31{
				Email: "team@linkspace.space",
				Name:  i.InvitedByName,
			},
			To: &mailjet.RecipientsV31{
				mailjet.RecipientV31{
					Email: i.InvitedEmail,
					Name:  strings.Split(i.InvitedEmail, "@")[0],
				},
			},
			Subject:  fmt.Sprintf(`%s has shared a collection with you 👏`, i.InvitedByName),
			HTMLPart: fmt.Sprintf(`<p>Hey %s</p><p><span style="text-transform: capitalize;">%s</span>(%s) has shared a links collection (%s) with you.</p><a href="http://localhost:5173/accounts/accept_invite?token=%s">Click here to join them.</a><p>Your invite expires on %dth %s %d so <a href="http://localhost:5173/accounts/accept_invite?token=%s">join %s now</a>.</p><p>Regards,</P><p><a href="beta.linkspace.space">Linkspace</a> Team.</p>`, strings.Split(i.InvitedEmail, "@")[0], i.InvitedByName, i.InvitedByMail, i.CollectionName, i.Token, i.Expiry.Day(), i.Expiry.Month(), i.Expiry.Year(), i.Token, i.InvitedByName),
		},
	}
	messages := mailjet.MessagesV31{Info: messagesInfo}

	_, err = client.SendMailV31(&messages)
	if err != nil {
		log.Panicf("could not send invite email: %v", err)
	}
}
