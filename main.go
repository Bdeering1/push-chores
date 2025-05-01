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
const DataFile = "test.ht"

func main() {
    force := flag.Bool("f", false, "skip checks and send email")

    contacts := map[string]Person{}
    contactFile, err := os.Open(ContactFile); check(err)
    defer contactFile.Close()
    for sc := bufio.NewScanner(contactFile); sc.Scan(); {
        n, e := splitTuple(sc.Text(), " ")
        contacts[n] = Person{name: n, email: e}
    }

    chores := []string{}
    people := []string{}
    dataFile, err := os.Open(DataFile); check(err)
    defer dataFile.Close()
    sc := bufio.NewScanner(dataFile)
    for sc.Scan() && sc.Text() != "---" {
        person := normalize(sc.Text());
        checkDict(contacts, person, "'%s' not in contacts")
        people = append(people, person)
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
    _, week := time.Now().In(est).ISOWeek()

    if *force || week != lastWeek {
        rotation++
        login()
        content := MailContent{
            subject: "Test Message",
            body:    "Hello, this is chore bot.<br/>",
        }
        send(content, []string{ contacts[people[0]].email })
    } else {
        fmt.Println("Already notified for this week. Use -f to force.")
    }

    rotationFile, err := os.Create(rtnFileName); check(err)
    defer rotationFile.Close()
    s := fmt.Sprintf("%d %d", week, rotation)
    _, err = rotationFile.WriteString(s); check(err)
}

func splitTuple(s, sep string) (string, string) {
    a := strings.Split(s, sep)
    return normalize(a[0]), normalize(a[1])
}

func normalize(s string) string {
    return strings.ToLower(strings.TrimSpace(s))
}

func rotationFileName(f string) string {
    return fmt.Sprintf(".%s-rotation.ht", fileStem(f))
}

func fileStem(f string) string {
    return f[:len(f) - len(filepath.Ext(f))]
}

func checkDict[V any](d map[string]V, k, msg string) {
    if _, ok := d[k]; !ok { log.Fatalf(msg, k) }
}

func check(err error) {
    if err != nil { log.Fatalf("error: %v", err) }
}
