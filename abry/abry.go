package main

import (
    "bufio"
    "fmt"
    "os"
    "strings"
)

const abbreviationsFile = "~/.config/fish/functions/__fish_abbreviations.fish"
const abbreviationCommand = "abbr"

func check(e error) {
    if e != nil {
        panic(e)
    }
}

func expandHome(path string) (string) {
    home := os.Getenv("HOME")
    return strings.Replace(path, "~", home, 1)
}

func printUsage() {
    fmt.Printf("usage: abry <abbreviation> <abbreviation-definition>\n")
}

func getTokenWithIndex(line string, tokenIndex int) (string) {
    trimmedLine := strings.TrimSpace(line)
    tokens := strings.Split(trimmedLine, " ")
    return tokens[tokenIndex]
}

func getFirstToken(line string) (string) {
    return getTokenWithIndex(line, 0)
}

func getAbbrevName(line string) (string) {
    return getTokenWithIndex(line, 2)
}

func isAbbreviationLine(line string) (bool) {
    firstToken := getFirstToken(line)
    if firstToken == abbreviationCommand {
        return true
    }
    return false
}

func printAbbreviation(abbrName string, abbrPhrase string) {
    fmt.Printf("    abbr --add %s %s\n", abbrName, abbrPhrase)
}

func shouldPrintNewAbbreviation(abbrName string, existingAbbrev string, alreadyPrinted bool) (bool) {
    if !alreadyPrinted && existingAbbrev > abbrName {
        return true
    }
    return false
}

func main() {
    args := os.Args
    numberOfArgs := len(args)
    if numberOfArgs < 3 {
        printUsage()
        os.Exit(1)
    }
    abbrName := args[1]
    abbrPhrase := strings.Join(args[2:], " ")
    expanded := expandHome(abbreviationsFile)
    f, err := os.Open(expanded)
    check(err)
    reader := bufio.NewReader(f)
    check(err)
    found := false
    for {
        line, err := reader.ReadString('\n')
        if isAbbreviationLine(line) {
            existingAbbrev := getAbbrevName(line)
            if shouldPrintNewAbbreviation(abbrName, existingAbbrev, found) {
                printAbbreviation(abbrName, abbrPhrase)
                found = true
            }
            fmt.Printf("%s", line)
        } else {
            fmt.Printf("%s", line)
        }
        if err != nil {
            break
        }
    }
}
