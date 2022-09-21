package main

func (a *App) sendEmail(msg Message) {
	a.WaitGroup.Add(1)
	a.Mailer.MailerChan <- msg
}
