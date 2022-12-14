package main

import (
	"bytes"
	"fmt"
	"github.com/vanng822/go-premailer/premailer"
	mail "github.com/xhit/go-simple-mail/v2"
	"html/template"
	"sync"
	"time"
)

type Mail struct {
	Domain      string
	Host        string
	Port        int
	Username    string
	Password    string
	Encryption  string
	FromAddress string
	FromName    string
	WaitGroup   *sync.WaitGroup
	MailerChan  chan Message
	ErrorChan   chan error
	DoneChan    chan bool
}

type Message struct {
	FromAddress    string
	FromName       string
	ToAddress      string
	Subject        string
	Attachments    []string
	AttachmentsMap map[string]string
	Data           interface{}
	DataMap        map[string]interface{}
	Template       string
}

func (a *App) listenForMail() {
	for {
		select {
		case msg := <-a.Mailer.MailerChan:
			go a.Mailer.SendMail(msg, a.Mailer.ErrorChan)
		case err := <-a.Mailer.ErrorChan:
			a.ErrorLog.Println(err)
		case <-a.Mailer.DoneChan:
			return
		}
	}
}

func (m *Mail) SendMail(msg Message, errorChan chan error) {
	defer m.WaitGroup.Done()

	if msg.Template == "" {
		msg.Template = "mail"
	}

	if msg.FromAddress == "" {
		msg.FromAddress = m.FromAddress
	}

	if msg.FromName == "" {
		msg.FromName = m.FromName
	}

	if len(msg.DataMap) == 0 {
		msg.DataMap = map[string]interface{}{
			"message": msg.Data,
		}
	}

	formattedMessage, err := m.buildHTMLMessage(msg)
	if err != nil {
		errorChan <- err
		return
	}

	plainMessage, err := m.buildPlainTextMessage(msg)
	if err != nil {
		errorChan <- err
		return
	}

	server := mail.NewSMTPClient()
	server.Host = m.Host
	server.Port = m.Port
	server.Username = m.Username
	server.Password = m.Password
	server.Encryption = m.getEncryption(m.Encryption)
	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	smtpClient, err := server.Connect()
	if err != nil {
		errorChan <- err
	}

	email := mail.NewMSG()
	email.SetFrom(msg.FromAddress).AddTo(msg.ToAddress).SetSubject(msg.Subject)
	email.SetBody(mail.TextPlain, plainMessage)
	email.AddAlternative(mail.TextHTML, formattedMessage)

	if len(msg.Attachments) > 0 {
		for _, attachment := range msg.Attachments {
			email.AddAttachment(attachment)
		}
	}

	if len(msg.AttachmentsMap) > 0 {
		for key, value := range msg.AttachmentsMap {
			email.AddAttachment(value, key)
		}
	}

	err = email.Send(smtpClient)
	if err != nil {
		errorChan <- err
	}
}

func (m *Mail) buildHTMLMessage(msg Message) (string, error) {
	templateToRender := fmt.Sprintf("./cmd/web/templates/%s.html.gohtml", msg.Template)

	t, err := template.New("email-html").ParseFiles(templateToRender)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	err = t.ExecuteTemplate(&tpl, "body", msg.DataMap)
	if err != nil {
		return "", err
	}

	formattedMessage := tpl.String()
	formattedMessage, err = m.inlineCSS(formattedMessage)
	if err != nil {
		return "", err
	}

	return formattedMessage, nil
}

func (m *Mail) buildPlainTextMessage(msg Message) (string, error) {
	templateToRender := fmt.Sprintf("./cmd/web/templates/%s.plain.gohtml", msg.Template)

	t, err := template.New("email-html").ParseFiles(templateToRender)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	err = t.ExecuteTemplate(&tpl, "body", msg.DataMap)
	if err != nil {
		return "", err
	}

	return tpl.String(), nil
}

func (m *Mail) inlineCSS(s string) (string, error) {
	prem, err := premailer.NewPremailerFromString(s, &premailer.Options{
		RemoveClasses:     false,
		CssToAttributes:   false,
		KeepBangImportant: true,
	})
	if err != nil {
		return "", err
	}

	html, err := prem.Transform()
	if err != nil {
		return "", err
	}

	return html, nil
}

func (m *Mail) getEncryption(e string) mail.Encryption {
	switch e {
	case "tls":
		return mail.EncryptionSTARTTLS
	case "ssl":
		return mail.EncryptionSSLTLS
	case "none":
		return mail.EncryptionNone
	default:
		return mail.EncryptionSTARTTLS
	}
}
