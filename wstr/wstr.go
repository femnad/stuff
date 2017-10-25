package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const (
	credentialsFile = "~/.aws/credentials"
	name            = "wstr"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func getAWSCredentialsFile() string {
	home := os.Getenv("HOME")
	return strings.Replace(credentialsFile, "~", home, 1)
}

func appendProfileToCredentials(profileName string, id string, secret string) {
	credentialsFile := getAWSCredentialsFile()
	file, err := os.OpenFile(credentialsFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	check(err)
	defer file.Close()

	profileLine := fmt.Sprintf("[%s]\n", profileName)
	idLine := fmt.Sprintf("aws_access_key_id = %s\n", id)
	secretLine := fmt.Sprintf("aws_secret_access_key = %s\n", secret)

	writer := bufio.NewWriter(file)
	writer.WriteString(profileLine)
	writer.WriteString(idLine)
	writer.WriteString(secretLine)
	writer.Flush()
}

func main() {
	args := os.Args[1:]
	if len(args) != 3 {
		fmt.Printf("usage: %s <profile-name> <id> <secret>\n", name)
		os.Exit(1)
	}

	profileName := args[0]
	id := args[1]
	secret := args[2]

	appendProfileToCredentials(profileName, id, secret)
}
