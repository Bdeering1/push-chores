package main

import (
    "bufio"
    "fmt"
    "log"
    "os"
    "strconv"
    "strings"
    "time"
)

type Person struct {
    name  string
    email string
}

const DataFile = "data.txt"
const RotationFile = "rotation.txt"

func main() {
    dataFile, err := os.Open(DataFile); check(err)
    defer dataFile.Close()

    people := []Person{}
    chores := []string{}

    sc := bufio.NewScanner(dataFile)
    for sc.Scan() && sc.Text() != "" {
        a := strings.Split(sc.Text(), " ")
        people = append(people, Person{name: a[0], email: a[1]})
    }
    for sc.Scan() {
        chores = append(chores, strings.TrimSpace(sc.Text()))
    }

    lastWeek, rotation := 0, 0
    _, err = os.Stat(RotationFile)
    if !os.IsNotExist(err) {
        b, err := os.ReadFile(RotationFile); check(err)

        s := string(b)
        a := strings.Split(s, " ")
        lastWeek, err = strconv.Atoi(a[0]); check(err)
        rotation, err = strconv.Atoi(a[1]); check(err)
    }

    est, _ := time.LoadLocation("EST")
    _, week := time.Now().In(est).ISOWeek()

    if week != lastWeek {
        rotation++
        // notify
    }

    rotationFile, err := os.Create(RotationFile); check(err)
    defer rotationFile.Close()
    s := fmt.Sprintf("%d %d", week, rotation)
    _, err = rotationFile.WriteString(s); check(err)
}

func check(err error) {
    if err != nil {
        log.Fatalf("error: %v", err)
    }
}
