package main

import (
    "bufio"
    "flag"
    "fmt"
    "log"
    "os"
    "path/filepath"
    "strconv"
    "strings"
    "time"
)

type Person struct {
    name  string
    email string
}

const ContactFile = "contacts.ht"
const DataFile = "weekly.ht"

func main() {
    force := flag.Bool("f", false, "skip checks and send email")
    fRotation := flag.Int("r", -1, "manually set rotation")
    testOnly := flag.Bool("t", false, "send test email without affecting rotation")
    flag.Parse()

    contacts := map[string]Person{}
    contactFile, err := os.Open(ContactFile); check(err)
    defer contactFile.Close()
    for sc := bufio.NewScanner(contactFile); sc.Scan(); {
        n, e := splitTuple(sc.Text(), " ")
        contacts[n] = Person{name: n, email: e}
    }

    chores := []string{}
    people := []string{}
    to := []string{}
    dataFile, err := os.Open(DataFile); check(err)
    defer dataFile.Close()
    sc := bufio.NewScanner(dataFile)
    for sc.Scan() && sc.Text() != "---" {
        person := normalize(sc.Text());
        people = append(people, person)
        if _, ok := contacts[person]; !ok {
            if *force { continue }
            log.Fatalf("%s not in contacts", person)
        }
        if *testOnly { continue }
        to = append(to, contacts[person].email)
    }
    for sc.Scan() {
        chores = append(chores, normalize(sc.Text()))
    }

    lastWeek, rotation := 0, 0
    rtnFileName := rotationFileName(DataFile)
    _, err = os.Stat(rtnFileName)
    if !os.IsNotExist(err) {
        b, err := os.ReadFile(rtnFileName); check(err)

        a := strings.Split(string(b), " ")
        lastWeek, err = strconv.Atoi(a[0]); check(err)
        rotation, err = strconv.Atoi(a[1]); check(err)
    }

    est, _ := time.LoadLocation("EST")
    t := time.Now().In(est)
    _, week := t.ISOWeek()

    if *force || week != lastWeek {
        if !*testOnly { rotation++ }
        if *fRotation != -1 { rotation = *fRotation }
        login()

        content := MailContent{
            subject: "Weekly Chore Rotation",
            body: craftMessage(t, rotation, people, chores),
        }
        to = append(to, contacts["automail"].email)
        send(content, to)
    } else {
        fmt.Println("Already notified for this week. Use -f to force.")
    }

    if *testOnly { return }
    rotationFile, err := os.Create(rtnFileName); check(err)
    defer rotationFile.Close()
    s := fmt.Sprintf("%d %d", week, rotation)
    _, err = rotationFile.WriteString(s); check(err)
}

func craftMessage(t time.Time, rot int, people []string, chores []string) string {
    signoffs := []string{
        "Yours truly",
        "Sincerely",
        "Warm regards",
        "All the best",
        "Respectfully",
        "Cheers",
        "Cordially",
        "Have a great week",
        "Keep up the good work",
    }

    msg := fmt.Sprintf("%s, %s %d, %d<br/><br/>Chores this week:<br/>", t.Weekday(), t.Month(), t.Day(), t.Year())
    for i, p := range people {
        msg += fmt.Sprintf("%s - %s<br/>", p, chores[(i + rot) % len(chores)])
    }
    return msg + fmt.Sprintf("<br/>%s,<br/>Chore Bot", signoffs[rot % len(signoffs)])
}

func splitTuple(s, sep string) (string, string) {
    a := strings.Split(s, sep)
    return normalize(a[0]), normalize(a[1])
}

func normalize(s string) string {
    return strings.ToLower(strings.TrimSpace(s))
}

func rotationFileName(f string) string {
    return fmt.Sprintf(".rotations/%s-rotation.ht", fileStem(f))
}

func fileStem(f string) string {
    return f[:len(f) - len(filepath.Ext(f))]
}

func check(err error) {
    if err != nil { log.Fatalf("error: %v", err) }
}
