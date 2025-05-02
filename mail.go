package main

import (
    "bufio"
    "fmt"
    "log"
    "os"
    "time"

    "github.com/joho/godotenv"
    "gopkg.in/gomail.v2"
)

type MailContent struct {
    subject string
    message string
}

var sender gomail.SendCloser

func login() {
    err := godotenv.Load(); check(err)
    user := tryGetEnv("MAIL_USER")
    pass := tryGetEnv("MAIL_PASS")

    d := gomail.NewDialer("smtp.gmail.com", 465, user, pass);
    sender, err = d.Dial(); check(err)
}

func send(cont MailContent, to []string) {
    if len(to) == 0 { log.Fatalf("Cannot send mail: recipient list is empty") }
    user := tryGetEnv("MAIL_USER")

    m := gomail.NewMessage()
    m.SetHeader("From", user)
    m.SetHeader("To", to...)
    m.SetHeader("Subject", cont.subject)

    est, _ := time.LoadLocation("EST")
    t := time.Now().In(est)

    body := fmt.Sprintf("%s, %s %d, %d<br/><br/>%s<br/>Autobot - bot things<br/>Bryn - human things<br/><br/>Yours truly,<br/>Chore Bot", t.Weekday(), t.Month(), t.Day(), t.Year(), cont.message)
    m.SetBody("text/html", body)

    fmt.Printf("\nTo: %v\nSubject: %s\nMessage:\n%s\n\n", to, cont.subject, body)
    fmt.Print("Send? ")

    sc := bufio.NewScanner(os.Stdin)
    sc.Scan(); input := normalize(sc.Text())
    if input != "y" && input != "yes" {
        fmt.Println("Nothing sent")
        return
    }

    err := gomail.Send(sender, m); check(err)
    fmt.Println("Success")
}

func tryGetEnv(name string) string {
    v := os.Getenv(name)
    if v == "" { log.Fatalf("No value for %s", v) }
    return v
}
