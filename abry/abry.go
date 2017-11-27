package main

import (
	"bufio"
	"errors"
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

func expandHome(path string) string {
	home := os.Getenv("HOME")
	return strings.Replace(path, "~", home, 1)
}

func printUsage() {
	fmt.Printf("usage: abry <abbreviation> <abbreviation-definition>\n")
}

func getAsTokens(line string) []string {
	trimmedLine := strings.TrimSpace(line)
	return strings.Split(trimmedLine, " ")
}
func getTokenWithIndex(line string, tokenIndex int) string {
	tokens := getAsTokens(line)
	return tokens[tokenIndex]
}

func getFirstToken(line string) string {
	return getTokenWithIndex(line, 0)
}

func getAbbrevName(line string) string {
	return getTokenWithIndex(line, 2)
}

func getAbbrevPhrase(line string) string {
	tokens := getAsTokens(line)
	phraseTokens := tokens[3:]
	return strings.Join(phraseTokens, " ")
}

func isAbbreviationLine(line string) bool {
	firstToken := getFirstToken(line)
	if firstToken == abbreviationCommand {
		return true
	}
	return false
}

func getAbbrevCommand(abbrName string, abbrPhrase string) string {
	return fmt.Sprintf("    abbr --add %s '%s'\n", abbrName, abbrPhrase)
}

func maybeWriteNewAbbreviation(line string, abbrName string, abbrPhrase string, writer *bufio.Writer) (bool, error) {
	existingAbbrev := getAbbrevName(line)
	if existingAbbrev > abbrName {
		abbrevCommand := getAbbrevCommand(abbrName, abbrPhrase)
		writer.WriteString(abbrevCommand)
		return true, nil
	} else if existingAbbrev == abbrName {
		existingAbbrevPhrase := getAbbrevPhrase(line)
		fmt.Printf("Abbreviation `%s` already exists with definition `%s`\n", abbrName, existingAbbrevPhrase)
		return false, errors.New("already exists")
	}
	return false, nil
}

func addAbbreviation(abbrName string, abbrPhrase string) bool {
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
	var writeErr error
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		if isAbbreviationLine(line) && !found {
			found, writeErr = maybeWriteNewAbbreviation(line, abbrName, abbrPhrase, writer)
		}
		if writeErr != nil {
			return false
		}
		writer.WriteString(line)
	}

	writer.Flush()

	return found
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

	addOk := addAbbreviation(abbrName, abbrPhrase)
	if addOk {
		os.Exit(0)
	}
	os.Exit(2)
}
