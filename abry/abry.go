package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

const (
	abbreviationCommand = "abbr"
	abbreviationPrefix  = "~/.config/fish/functions/"
	defaultFileMode     = 0644
	publicFile          = "__fish_abbreviations.fish"
	privateFile         = "__self_abbreviations.fish"
	tempFile            = "/tmp/fish_abbrevs.fish"
)

var abbreviationFiles = map[string]string{
	"public":  publicFile,
	"private": privateFile,
}

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
	return firstToken == abbreviationCommand
}

func getAbbrevCommand(abbrName string, abbrPhrase string) string {
	return fmt.Sprintf("abbr --add %s '%s'\n", abbrName, abbrPhrase)
}

func writeAbbreviation(writer *bufio.Writer, abbrName, abbrPhrase string) {
	abbrevCommand := getAbbrevCommand(abbrName, abbrPhrase)
	writer.WriteString(abbrevCommand)
}

func maybeWriteNewAbbreviation(line string, abbrName string, abbrPhrase string, writer *bufio.Writer) (bool, error) {
	existingAbbrev := getAbbrevName(line)
	if existingAbbrev > abbrName {
		writeAbbreviation(writer, abbrName, abbrPhrase)
		return true, nil
	} else if existingAbbrev == abbrName {
		existingAbbrevPhrase := getAbbrevPhrase(line)
		fmt.Printf("Abbreviation `%s` already exists with definition `%s`\n", abbrName, existingAbbrevPhrase)
		return false, errors.New("already exists")
	}
	return false, nil
}

func doesNotExist(fileName string) bool {
	_, err := os.Stat(fileName)
	return os.IsNotExist(err)
}

func maybeCreateFile(fileName string) {
	if doesNotExist(fileName) {
		log.Printf("Creating file %s", fileName)
		file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, defaultFileMode)
		check(err)
		err = file.Close()
		check(err)
	}
}

func maybeInitialiseNewFile(line string, err error, abbrName, abbrPhrase string, writer *bufio.Writer) {
	if line == "" && err == io.EOF {
		writeAbbreviation(writer, abbrName, abbrPhrase)
	}
}

func addAbbreviation(fileType, abbrName, abbrPhrase string) bool {
	baseFileName, ok := abbreviationFiles[fileType]
	if !ok {
		log.Fatalf("No abbreviation file of type %s", fileType)
	}
	abbreviationsFile := expandHome(abbreviationPrefix + baseFileName)

	maybeCreateFile(abbreviationsFile)

	abbrevsFile, err := os.Open(abbreviationsFile)
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
			maybeInitialiseNewFile(line, err, abbrName, abbrPhrase, writer)
			break
		}
		if (isAbbreviationLine(line) && !found) || line == "" {
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

func getTypeAbbrevAndCommand() (string, string, string) {
	fileType := flag.String("file", "public", "Type of abbreviation file")
	flag.Parse()
	abbrevAndcommandList := flag.Args()
	if len(abbrevAndcommandList) < 2 {
		panic("Need an abbreviation and at least one phrase")
	}
	abbreviation := abbrevAndcommandList[0]
	commands := abbrevAndcommandList[1:]
	command := strings.Join(commands, " ")
	return *fileType, abbreviation, command
}

func main() {
	fileType, abbrName, abbrPhrase := getTypeAbbrevAndCommand()

	addOk := addAbbreviation(fileType, abbrName, abbrPhrase)
	if addOk {
		os.Exit(0)
	}
	os.Exit(2)
}
