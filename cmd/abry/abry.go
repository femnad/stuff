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

func maybeWriteNewAbbreviation(writer *bufio.Writer, line, abbrName, abbrPhrase string) (bool, error) {
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

func maybeCreateFile(fileName string) bool {
	if doesNotExist(fileName) {
		log.Printf("Creating file %s", fileName)
		file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, defaultFileMode)
		check(err)
		err = file.Close()
		check(err)
		return true
	}
	return false
}

func getFileOfType(fileType string) string {
	baseFileName, ok := abbreviationFiles[fileType]
	if !ok {
		log.Fatalf("No abbreviation file of type %s", fileType)
	}
	return expandHome(abbreviationPrefix + baseFileName)
}

func writeAbbrevWhereSuitable(reader *bufio.Reader, writer *bufio.Writer, abbrName, abbrPhrase string) {
	var writeErr error
	found := false

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		if (isAbbreviationLine(line) && !found) || line == "" {
			found, writeErr = maybeWriteNewAbbreviation(writer, line, abbrName, abbrPhrase)
		}
		check(writeErr)
		writer.WriteString(line)
	}
	if !found {
		writeAbbreviation(writer, abbrName, abbrPhrase)
	}
}

func addAbbreviation(fileType, abbrName, abbrPhrase string) {
	abbreviationsFile := getFileOfType(fileType)

	newlyCreated := maybeCreateFile(abbreviationsFile)

	abbrevsFile, err := os.Open(abbreviationsFile)
	check(err)
	defer abbrevsFile.Close()

	reader := bufio.NewReader(abbrevsFile)
	check(err)

	tempFile, err := os.Create(tempFile)
	check(err)
	defer tempFile.Close()

	writer := bufio.NewWriter(tempFile)

	if newlyCreated {
		writeAbbreviation(writer, abbrName, abbrPhrase)
	} else {
		writeAbbrevWhereSuitable(reader, writer, abbrName, abbrPhrase)
	}

	writer.Flush()
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

func copyTempFileToOriginal(fileType string) {
	temp, err := os.Open(tempFile)
	check(err)

	abbreviationsFile := getFileOfType(fileType)
	updatee, err := os.OpenFile(abbreviationsFile, os.O_RDWR, defaultFileMode)
	check(err)

	_, err = io.Copy(updatee, temp)
	check(err)

	temp.Close()
	updatee.Close()

	err = os.Remove(tempFile)
	check(err)
}

func main() {
	fileType, abbrName, abbrPhrase := getTypeAbbrevAndCommand()

	addAbbreviation(fileType, abbrName, abbrPhrase)
	copyTempFileToOriginal(fileType)

	os.Exit(2)
}
