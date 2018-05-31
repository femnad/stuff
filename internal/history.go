package internal

import (
	"bufio"
	"fmt"
	"github.com/femnad/mare"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
)

type History map[string]int
type reverseHistory map[int][]string

type intSet map[int]bool

func addToSet(set intSet, item int) {
	set[item] = true
}

func getSetAsSlice(set intSet) []int {
	items := make([]int, 0)
	for item, _ := range set {
		items = append(items, item)
	}
	return items
}

func IsInHistory(history History, item string) bool {
	_, ok := history[item]
	return ok
}

func AddToHistory(history History, item string) {
	if IsInHistory(history, item) {
		history[item] += 1
	} else {
		history[item] = 1
	}
}

func WriteHistoryToFile(history History, file *os.File) {
	for historyItem, count := range history {
		line := fmt.Sprintf("%s %d\n", historyItem, count)
		_, err := file.WriteString(line)
		mare.PanicIfErr(err)
	}
}

func getItemAndCountFromLine(historyLine string) (string, int) {
	trimmedHistoryLine := strings.TrimSpace(historyLine)
	splitWords := strings.Split(trimmedHistoryLine, " ")
	if len(splitWords) != 2 {
		errorMessage := fmt.Sprintf("Unexpected line: %s", trimmedHistoryLine)
		panic(errorMessage)
	}
	item := splitWords[0]
	countString := splitWords[1]
	count, err := strconv.ParseInt(countString, 10, 64)
	countAsInt := int(count)
	mare.PanicIfErr(err)
	return item, countAsInt
}

func GetHistoryFromFile(reader *bufio.Reader) History {
	history := make(History)
	for {
		line, err := reader.ReadString('\n')
		mare.PanicIfNotOfType(err, io.EOF)
		if err != nil {
			break
		}
		item, count := getItemAndCountFromLine(line)
		history[item] = count
	}
	return history
}

func appendToCountMap(countToItemMap reverseHistory, count int, item string) {
	itemListForCount, ok := countToItemMap[count]
	if ok {
		itemListForCount = append(itemListForCount, item)
		sort.Strings(itemListForCount)
		countToItemMap[count] = itemListForCount
	} else {
		countToItemMap[count] = []string{item}
	}
}

func buildReverseMap(history History) (reverseHistory, []int) {
	countToItemMap := make(reverseHistory)
	countSet := make(intSet)
	for item, count := range history {
		addToSet(countSet, count)
		appendToCountMap(countToItemMap, count, item)
	}
	counts := getSetAsSlice(countSet)
	sort.Sort(sort.Reverse(sort.IntSlice(counts)))
	return countToItemMap, counts
}

func GetOrderedHistoryByCount(history History) []string {
	reverseHistory, orderedCounts := buildReverseMap(history)
	itemsOrderedByCount := make([]string, 0)
	for _, count := range orderedCounts {
		items := reverseHistory[count]
		itemsOrderedByCount = append(itemsOrderedByCount, items...)
	}
	return itemsOrderedByCount
}
