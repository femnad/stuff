package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/femnad/mare"
)

const (
	historyFileName = "~/.local/share/fish/fish_history"
)

func getHistoryCommandLines(cmd string) string {
	now := time.Now()
	unixNow := now.Unix()
	historyLines := fmt.Sprintf("- cmd: %s\n  when: %d\n", cmd, unixNow)
	return historyLines
}

func appendToHistory(cmd string) {
	historyFilePath := mare.ExpandUser(historyFileName)
	historyFile, err := os.OpenFile(historyFilePath, os.O_RDWR|os.O_APPEND, 0644)
	defer historyFile.Close()

	mare.PanicIfErr(err)

	writer := bufio.NewWriter(historyFile)
	historyLines := getHistoryCommandLines(cmd)
	_, err = writer.WriteString(historyLines)

	mare.PanicIfErr(err)

	writer.Flush()
}

func printUsage() {
	progName := os.Args[0]
	fmt.Printf("usage: %s <command>", progName)
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
	args := os.Args[1:]
	command := strings.Join(args, " ")
	appendToHistory(command)
}
