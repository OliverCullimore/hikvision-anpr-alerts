package models

import (
	"errors"
	"fmt"
	"github.com/containrrr/shoutrrr"
	"github.com/containrrr/shoutrrr/pkg/types"
	"github.com/matcornic/hermes/v2"
	"net/url"
	"time"
)

// Email struct
type Email struct {
	To      string
	Subject string
	Body    hermes.Body
}

// Send Email
func (e *Email) Send(env *Env) error {
	if env.Config.SMTPHost != "" && env.Config.SMTPFrom != "" {
		// Configure hermes by setting the header & footer of e-mails
		h := hermes.Hermes{
			Product: hermes.Product{
				Name:      "Hikvision ANPR Alerts",
				Link:      env.Config.ExternalURL,
				Copyright: fmt.Sprintf("Copyright © %s Hikvision ANPR Alerts. All rights reserved.", time.Now().Format("2006")),
			},
		}
		// Generate an HTML email
		emailBody, err := h.GenerateHTML(hermes.Email{Body: e.Body})
		if err != nil {
			return err
		}
		// Create sender
		senderURL := fmt.Sprintf("%s:%s", env.Config.SMTPHost, env.Config.SMTPPort)
		if env.Config.SMTPUser != "" {
			if env.Config.SMTPPass != "" {
				senderURL = fmt.Sprintf("%s:%s@%s", env.Config.SMTPUser, env.Config.SMTPPass, senderURL)
			} else {
				senderURL = fmt.Sprintf("%s@%s", env.Config.SMTPUser, senderURL)
			}
		}
		senderURL = fmt.Sprintf("smtp://%s/?usehtml=yes&auth=%s&fromaddress=%s&fromname=%s&subject=%s&toaddresses=%s", senderURL, url.QueryEscape(env.Config.SMTPAuth), url.QueryEscape(env.Config.SMTPFrom), url.QueryEscape("Hikvision ANPR Alerts"), url.QueryEscape(e.Subject), url.QueryEscape(e.To))
		sender, err := shoutrrr.NewSender(env.Logger, senderURL)
		if err != nil {
			return err
		}
		// Send email instantly
		sender.Send(emailBody, (*types.Params)(&map[string]string{"title": e.Subject}))
	} else {
		return errors.New("no SMTP configured")
	}
	return nil
}
