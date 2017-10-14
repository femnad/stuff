package main

import (
    "bufio"
    "fmt"
    "os"
    "strings"
)

const abbreviationsFile = "~/.config/fish/functions/__fish_abbreviations.fish"
const abbreviationCommand = "abbr"
const tempFile = "/tmp/fish_abbrevs.fish"

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

func getAbbrevCommand(abbrName string, abbrPhrase string) (string) {
    return fmt.Sprintf("    abbr --add %s %s\n", abbrName, abbrPhrase)
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
    abbrevsFile, err := os.Open(expanded)
    check(err)
    defer abbrevsFile.Close()

    reader := bufio.NewReader(abbrevsFile)
    check(err)

    tempFile, err := os.Create(tempFile)
    check(err)
    defer tempFile.Close()

    writer := bufio.NewWriter(tempFile)

    found := false
    for {
        line, err := reader.ReadString('\n')
        if isAbbreviationLine(line) {
            existingAbbrev := getAbbrevName(line)
            if shouldPrintNewAbbreviation(abbrName, existingAbbrev, found) {
                abbrevCommand := getAbbrevCommand(abbrName, abbrPhrase)
                writer.WriteString(abbrevCommand)
                found = true
            }
            writer.WriteString(line)
        } else {
            writer.WriteString(line)
        }
        if err != nil {
            break
        }
    }

    writer.Flush()
}
