package emails

import (
	"context"
	"time"

	"github.com/mailgun/mailgun-go/v4"
)

func (e *Mail) SendEmailVerificationCode() (string, string, error) {
	mg := mailgun.NewMailgun(e.Domain, e.ApiKey)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

// 	templates := mg.ListTemplates(nil)
// 	var page, results []mailgun.Template
// 	if templates.Next(ctx, &page) {
// 		results = append(results, page...)
// 		for _, r := range results {
// 			if r.Name == "email-verification-code" {
// 				templateName := "email-verification-code"
// 				err := mg.DeleteTemplate(ctx, templateName)
// 				if err != nil {
// 					log.Println(err)
// 					return "", "", nil
// 				}
// 			}
// 		}
// 	}

// 	err := mg.CreateTemplate(ctx, &mailgun.Template{
// 		Name:        "email-verification-code",
// 		CreatedAt:   mailgun.RFC2822Time{},
// 		Version: mailgun.TemplateVersion{
// 			Template: `
// 			<html>
// 	<body>
// 		<p>Hi,</p>
// 		<p>Thanks for getting started with Bookmarked.</p>
// 		<p>Your verification code is <span style="font-weight: bold; text-decoration: underline">{{.code}}</span></p>
// 		<p>Enter this code into your browser window to activate your account.</p>
// 		<p>We are glad you're here!</p>
// 		<p>The Bookmarked Team.</p>
// 	</body>
// </html>


// 			`,
// 			Engine:    mailgun.TemplateEngineGo,
// 			Tag:       "auth",
// 			Active:    true,
// 			CreatedAt: mailgun.RFC2822Time(time.Now()),
// 		},
// 	})
// 	if err != nil {
// 		log.Println(err)
// 		return "", "", err
// 	}

	t, err := mg.GetTemplate(context.Background(), "email-verification-code")
	if err != nil {
		return "", "", err
	}

	time.Sleep(time.Second * 1)

	message := mg.NewMessage(e.Sender, e.Subject, "", e.Recipients...)
	message.SetTemplate(t.Name)

	message.AddVariable("code", e.Code)

	msg, id, err := mg.Send(ctx, message)
	if err != nil {
		return "", "", err
	}

	return msg, id, err
}
