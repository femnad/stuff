package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/femnad/mare"
	"github.com/femnad/stuff/pkg/history"
)

const (
	GpgFileExtension = ".gpg"
	HistoryFile      = "~/.config/fred/fred_history"
	PasswordStore    = "~/.password-store"
)

func getDirsAndFiles(path string) ([]os.FileInfo, []os.FileInfo) {
	dirs := make([]os.FileInfo, 0)
	files := make([]os.FileInfo, 0)
	dirContent, err := ioutil.ReadDir(path)

	mare.PanicIfErr(err)

	for _, fileInfo := range dirContent {
		if fileInfo.IsDir() {
			dirs = append(dirs, fileInfo)
		} else {
			files = append(files, fileInfo)
		}
	}
	return dirs, files
}

func getAbsoluteNamesOfFiles(files []os.FileInfo, leadingPath string) []string {
	fileNames := mare.MapFileInfo(files, os.FileInfo.Name)
	prependPath := func(path string) string {
		return fmt.Sprintf("%s/%s", leadingPath, path)
	}
	return mare.Map(fileNames, prependPath)
}

func getDirsAndFilesAsStrings(path string) ([]string, []string) {
	dirs, files := getDirsAndFiles(path)
	dirNames := getAbsoluteNamesOfFiles(dirs, path)
	fileNames := getAbsoluteNamesOfFiles(files, path)
	return dirNames, fileNames
}

func isGpgFile(path string) bool {
	return strings.HasSuffix(path, GpgFileExtension)
}

func recursivelyListGpgFiles(path string) []string {
	dirs, files := getDirsAndFilesAsStrings(path)
	gpgFiles := mare.Filter(files, isGpgFile)
	deeperFiles := mare.FlatMap(dirs, recursivelyListGpgFiles)
	return append(gpgFiles, deeperFiles...)
}

func getRemoveStorePathPrefixFn(storePath string) func(string) string {
	return func(path string) string {
		return strings.TrimPrefix(path, storePath)
	}
}

func removeLeadingSlashAndExtension(name string) string {
	nameWithExtensionRemoved := strings.TrimSuffix(name, GpgFileExtension)
	return strings.TrimLeft(nameWithExtensionRemoved, "/")
}

func buildPasswordMap(passwords []string) map[string]bool {
	passwordMap := make(map[string]bool)
	for _, password := range passwords {
		passwordMap[password] = true
	}
	return passwordMap
}

func getPasswordNames() []string {
	passwordStore := mare.ExpandUser(PasswordStore)
	files := recursivelyListGpgFiles(passwordStore)
	filesWithoutStorePath := mare.Map(files, getRemoveStorePathPrefixFn(passwordStore))
	return mare.Map(filesWithoutStorePath, removeLeadingSlashAndExtension)
}

func getNewItemFilterFn(historyMap history.History) func(string) bool {
	return func(item string) bool {
		return !history.IsInHistory(historyMap, item)
	}
}

func getPasswordNamesNotInHistory(passwordNames []string, historyMap history.History) []string {
	filterFn := getNewItemFilterFn(historyMap)
	return mare.Filter(passwordNames, filterFn)
}

func filterRemovedHistoryItems(historyMap history.History, passwordMap map[string]bool) history.History {
	filteredHistoryItems := make(history.History)
	for historyItem := range historyMap {
		_, exists := passwordMap[historyItem]
		if exists {
			filteredHistoryItems[historyItem] = historyMap[historyItem]
		}
	}
	return filteredHistoryItems
}

func getOrderedPasswords() []string {
	passwordNames := getPasswordNames()
	passwordMap := buildPasswordMap(passwordNames)
	historyMap := getHistoryMap()
	existingHistoryItems := filterRemovedHistoryItems(historyMap, passwordMap)
	passwordsNotInHistory := getPasswordNamesNotInHistory(passwordNames, existingHistoryItems)
	orderedHistory := history.GetOrderedHistoryByCount(historyMap)
	return append(orderedHistory, passwordsNotInHistory...)
}

func printPasswords() {
	orderedPasswords := getOrderedPasswords()
	for _, file := range orderedPasswords {
		fmt.Println(file)
	}
}

func getHistoryMap() history.History {
	historyFile, err := mare.ExpandUserAndOpen(HistoryFile)
	if os.IsNotExist(err) {
		return make(history.History)
	}
	mare.PanicIfErr(err)
	defer historyFile.Close()
	reader := bufio.NewReader(historyFile)
	return history.GetHistoryFromFile(reader)
}

func prepareAndWriteHistoryToFile(historyMap history.History) {
	historyFile := mare.ExpandUser(HistoryFile)
	historyDirectory := path.Dir(historyFile)
	os.MkdirAll(historyDirectory, 0700|os.ModeDir)
	file, err := os.OpenFile(historyFile, os.O_RDWR|os.O_CREATE, 0600)
	mare.PanicIfErr(err)
	defer file.Close()
	history.WriteHistoryToFile(historyMap, file)
	file.Sync()
}

func appendPasswordToHistory(passwordName string) {
	historyMap := getHistoryMap()

	history.AddToHistory(historyMap, passwordName)

	prepareAndWriteHistoryToFile(historyMap)
}

func printAndAddToHistory(passwordName string) {
	appendPasswordToHistory(passwordName)
	fmt.Fprintln(os.Stderr, passwordName)
}

func main() {
	args := os.Args[1:]
	if len(args) > 0 {
		passwordName := args[0]
		printAndAddToHistory(passwordName)
		os.Exit(1)
	} else {
		printPasswords()
	}
}
