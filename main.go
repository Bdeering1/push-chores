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
const WeeklyFile = "weekly.ht"

var dry,
    force,
    testOnly,
    verbose,
    yesToAll *bool
var userRotation *int

var contacts map[string]Person
var tasks,
    people,
    recipients []string

func main() {
    dry          = flag.Bool("d", false, "do not send email")
    force        = flag.Bool("f", false, "skip checks and send email")
    testOnly     = flag.Bool("t", false, "send test email without affecting rotation")
    verbose      = flag.Bool("v", false, "print rotation information")
    yesToAll     = flag.Bool("y", false, "auto confirm all prompts")
    userRotation = flag.Int("r", -1, "manually set rotation")
    flag.Parse()

    readContacts()
    readDataFile(WeeklyFile)

    rtnFileName := rotationFileName(WeeklyFile)
    lastCycle, rotation := readRotationFile(rtnFileName)

    t := getTime()
    _, thisCycle := t.ISOWeek()

    if *verbose {
        fmt.Printf("Stored week: %d\nCurrent week: %d\n", lastCycle, thisCycle)
        fmt.Println("Rotation:", rotation)
    }

    if *dry { return }
    if *force || thisCycle != lastCycle {
        if !*testOnly { rotation++ }
        if *userRotation != -1 { rotation = *userRotation }
        login()

        content := MailContent{
            subject: "Weekly Chore Rotation",
            body: craftMessage(t, rotation, people, tasks),
        }
        recipients = append(recipients, contacts["automail"].email)
        send(content, recipients, *yesToAll)
    } else {
        fmt.Println("Already notified for this cycle. Use -f to force.")
    }

    if *testOnly { return }
    rotationFile, err := os.Create(rtnFileName); check(err)
    defer rotationFile.Close()
    s := fmt.Sprintf("%d %d", thisCycle, rotation)
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

    msg := fmt.Sprintf("%s - %s<br/><br/>Chores this week:<br/>", getTimeString(t), getTimeString(t.AddDate(0, 0, 6)))
    for i, p := range people {
        msg += fmt.Sprintf("%s - %s<br/>", p, chores[(i + rot) % len(chores)])
    }
    return msg + fmt.Sprintf("<br/>%s,<br/>Chore Bot", signoffs[rot % len(signoffs)])
}

func readDataFile(dFileName string) {
    tasks = []string{}
    people = []string{}
    recipients = []string{}

    dataFile, err := os.Open(dFileName); check(err)
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
        recipients = append(recipients, contacts[person].email)
    }
    for sc.Scan() {
        tasks = append(tasks, normalize(sc.Text()))
    }
}

func readContacts() {
    contacts = map[string]Person{}
    contactFile, err := os.Open(ContactFile); check(err)
    defer contactFile.Close()

    for sc := bufio.NewScanner(contactFile); sc.Scan(); {
        if len(sc.Text()) < 8 { continue }
        n, e := splitTuple(sc.Text(), " ")
        contacts[n] = Person{name: n, email: e}
    }
}

func readRotationFile(rtnFileName string) (last int, rot int) {
    _, err := os.Stat(rtnFileName)
    if !os.IsNotExist(err) {
        b, err := os.ReadFile(rtnFileName); check(err)

        a := strings.Split(string(b), " ")
        last, err = strconv.Atoi(a[0]); check(err)
        rot, err = strconv.Atoi(a[1]); check(err)
    }
    return
}

func getTime() time.Time {
    est, _ := time.LoadLocation("EST")
    return time.Now().In(est)
}

func getTimeString(t time.Time) string {
    return fmt.Sprintf("%s, %s %d, %d", t.Weekday(), t.Month(), t.Day(), t.Year())
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
