package main

import (
	"log"
	"time"
	"os"
	"html/template"
	"bytes"
	"context"
	"net/smtp"
	"net/textproto"
	"github.com/jordan-wright/email"
)

type MailClient struct {
	pool *email.Pool
	schedule chan email.Email
	done chan int
}

type BodyMail struct {
	User Order
	Products []Product
	Total int
}

func (app *Application) InitMail() (*MailClient, error) {
	host := "smtp.gmail.com"
	port := "587"
	username := os.Getenv("MAIL_USER")
	password := os.Getenv("MAIL_PASS")
	auth := smtp.PlainAuth("", username, password, host)
	pool, err := email.NewPool(host+":"+port, 4, auth)
	client := MailClient{pool, make(chan email.Email), make(chan int)}
	go client.ProcessMail()
	return &client, err
}

func (client *MailClient) ProcessMail() {
	for {
		select {
		case mail := <-client.schedule:
			client.pool.Send(&mail, time.Second*5)
		case <- client.done:
			client.pool.Close()
		}
	}
}

func (client *MailClient) SendMail(ctx context.Context, ch chan int) {
	log.Println("Sending")
	ord, _ := ctx.Value("data").(Order)
	log.Println(ctx.Value("total"))
	prod, _ := ctx.Value("prod").([]Product)
	total, _ := ctx.Value("total").(int)
	data := BodyMail{ord, prod, total}
	log.Println(data)
	ctx, cancel := context.WithCancel(ctx)
	temp, err := template.ParseFiles("./Front/html/mail_template.html")
	if err != nil {
		log.Println(err)
		cancel()
	}
	subject := "Thanks for purchase!"
	//headers := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	var body bytes.Buffer
	temp.Execute(&body, data)
	if err != nil {
		cancel()
		log.Println(err)
		return
	}
	mail := &email.Email {
		To: []string{ord.Client.Email},
		From: os.Getenv("MAIL_USER"),
		Subject: subject,
		Headers: textproto.MIMEHeader{},
		HTML: body.Bytes(),
	}
	client.schedule<-*mail
	ch <- 1
	//err = smtp.SendMail(mail.host+":"+mail.port, mail.auth, mail.username, []string{ord.Client.Email}, body.Bytes())
}