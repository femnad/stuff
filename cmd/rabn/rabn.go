package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/femnad/mare"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const (
	separator = " = "
)

func defaultHistoryFile() string {
	home := os.Getenv("HOME")
	return fmt.Sprintf("%s/%s", home, ".local/share/rabn/rabn_history")
}

func listPathContents(path string) []string {
	file, err := os.Open(path)
	mare.PanicIfErr(err)
	names, err := file.Readdirnames(0)
	mare.PanicIfErr(err)
	return mare.Map(names, func(baseName string) string {
		return filepath.Join(path, baseName)
	})
}

func listPathSpecContents(pathSpec string) []string {
	paths := strings.Split(pathSpec, ",")
	paths = mare.Map(paths, mare.ExpandUser)
	output := make([]string, 0)
	for _, path := range paths {
		output = append(output, listPathContents(path)...)
	}
	return output
}

type history map[string]int
type occurrences map[int][]string

func getOrderedItems(selectionHistory history) []string {
	occurrenceMap := make(occurrences)
	occurrences := make([]int, 0)
	for item, occurrence := range selectionHistory {
		items, alreadyExists := occurrenceMap[occurrence]
		if alreadyExists {
			occurrenceMap[occurrence] = append(items, item)
		} else {
			occurrenceMap[occurrence] = []string{item}
			occurrences = append(occurrences, occurrence)
		}
	}
	sort.Sort(sort.Reverse(sort.IntSlice(occurrences)))
	orderedItems := make([]string, 0)
	for _, occurrence := range occurrences {
		items := occurrenceMap[occurrence]
		orderedItems = append(orderedItems, items...)
	}
	return orderedItems
}

func parseDecimal(number string) int {
	parsed, err := strconv.ParseInt(number, 10, 64)
	mare.PanicIfErr(err)
	return int(parsed)
}

func getItemAndOccurrence(historyLine string) (string, int) {
	tokens := strings.Split(historyLine, separator)
	numTokens := len(tokens)
	if numTokens != 2 {
		message := fmt.Sprintf("Unexpected number of tokens %d for line %s", numTokens, historyLine)
		panic(message)
	}
	items := tokens[0]
	occurrence := parseDecimal(tokens[1])
	return items, occurrence
}

func historyFromFile(historyFile string) history {
	file, err := os.Open(historyFile)
	if os.IsNotExist(err) {
		return make(history)
	}
	mare.PanicIfErr(err)
	scanner := bufio.NewScanner(file)
	historyMap := make(history)
	for scanner.Scan() {
		entry := scanner.Text()
		item, occurrence := getItemAndOccurrence(entry)
		historyMap[item] = occurrence
	}
	return historyMap
}

func getOrderedFromHistory(historyFile string) []string {
	historyMap := historyFromFile(historyFile)
	return getOrderedItems(historyMap)
}

func getHistoryLine(item string, occurrence int) string {
	return fmt.Sprintf("%s%s%d\n", item, separator, occurrence)
}

func writeHistory(historyMap history, historyFile string) {
	dir := filepath.Dir(historyFile)
	os.MkdirAll(dir, 0755)
	file, err := os.OpenFile(historyFile, os.O_CREATE|os.O_RDWR, 0644)
	mare.PanicIfErr(err)
	defer file.Close()

	for item, occurrence := range historyMap {
		line := getHistoryLine(item, occurrence)
		_, err := file.WriteString(line)
		mare.PanicIfErr(err)
	}
}

func addToHistory(selection, historyFile string) {
	selection = mare.ExpandUser(selection)
	historyMap := historyFromFile(historyFile)
	historyMap[selection]++
	writeHistory(historyMap, historyFile)
}

func getNonOccurring(subList, superList []string) []string {
	return mare.Filter(superList, func(item string) bool {
		return !mare.Contains(subList, item)
	})
}

func mergeOutputWithHistory(pathSpec, historyFile string) []string {
	orderedItems := getOrderedFromHistory(historyFile)
	output := listPathSpecContents(pathSpec)
	itemsNotInHistory := getNonOccurring(orderedItems, output)
	return append(orderedItems, itemsNotInHistory...)
}

func listPathContentsWithHistory(pathSpec, historyFile string) {
	items := mergeOutputWithHistory(pathSpec, historyFile)
	for _, item := range items {
		fmt.Println(item)
	}
}

func main() {
	historyFile := flag.String("history-file", defaultHistoryFile(), "history file")
	pathSpec := flag.String("path-spec", ".", "list contents of path(s) [comma separated if multiple]")
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		listPathContentsWithHistory(*pathSpec, *historyFile)
	} else {
		addToHistory(args[0], *historyFile)
	}
}
