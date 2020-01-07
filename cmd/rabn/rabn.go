package main

import (
	"errors"
	"fmt"
	"github.com/alexflint/go-arg"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/femnad/mare"
	"gopkg.in/yaml.v2"
)

const (
	directoryPermissions = 0700
	filePermissions = 0600
)

var args struct{
	NumPathComponentsShown int `arg:"-P"`
	HistoryFile string `arg:"-H,required"`
	PathSpec string `arg:"-p"`
	Selection string `arg:"positional" default:""`
}

type history map[string]int

func ensureParent(file string) (err error) {
	dir := path.Dir(file)
	_, err = os.Stat(dir)
	if errors.Is(err, os.ErrNotExist) {
		err = os.MkdirAll(dir, directoryPermissions)
		if err != nil {
			return fmt.Errorf("error creating directory %s: %s", dir, err)
		}
	}
	return
}

func (h history) serialize(historyFile string) (err error) {
	if len(h) == 0 {
		return
	}
	out, err := yaml.Marshal(h)
	if err != nil {
		return err
	}
	err = ensureParent(historyFile)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(historyFile, out, filePermissions)
	if err != nil {
		return err
	}
	return
}

func (h *history) deserialize(historyFile string) error {
	contents, err := ioutil.ReadFile(historyFile)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(contents, h)
	if err != nil {
		return err
	}
	return nil
}

func (h *history) addToHistory(selection string) {
	selection = mare.ExpandUser(selection)
	(*h)[selection]++
}

func (h *history) eliminateStaleItems(listOutput []string) {
	for itemKey := range *h {
		if !mare.Contains(listOutput, itemKey) {
			delete(*h, itemKey)
		}
	}
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
	for _, p := range paths {
		_, err := os.Stat(p)
		if err != nil {
			continue
		}
		output = append(output, listPathContents(p)...)
	}
	return output
}

func getOrderedItems(h history) (orderedItems []string) {
	orderedMap := make(map[int][]string)
	for item, count := range h {
		items := orderedMap[count]
		orderedMap[count] = append(items, item)
	}

	counts := make([]int, len(orderedMap))
	for count, _ := range orderedMap {
		counts = append(counts, count)
	}

	sort.Sort(sort.Reverse(sort.IntSlice(counts)))
	sorted := make([]string, 0)
	for _, count := range counts {
		occurrences, _ := orderedMap[count]
		sorted = append(sorted, occurrences...)
	}
	return sorted
}

func initHistory(historyFile string) (h history, err error) {
	file, err := os.OpenFile(historyFile, os.O_CREATE|os.O_WRONLY, filePermissions)
	if err != nil {
		return h, fmt.Errorf("error creating history file: %s", err)
	}
	err = file.Close()
	if err != nil {
		return h, fmt.Errorf("error closing history file: %s", err)
	}
	h = make(history)
	return

}

func historyFromFile(historyFile string) (history, error) {
	_, err := os.Stat(historyFile)
	if os.IsNotExist(err) {
		return initHistory(historyFile)
	} else if err != nil && !os.IsNotExist(err) {
		return history{}, err
	}
	h := history{}
	err = h.deserialize(historyFile)
	return h, err
}

func addToHistory(historyFile, selection string) {
	historyMap, err := historyFromFile(historyFile)
	mare.PanicIfErr(err)
	historyMap.addToHistory(selection)
	err = historyMap.serialize(historyFile)
	mare.PanicIfErr(err)
}

func getNonOccurring(h history, allItems []string) []string {
	nonOccurring := make([]string, 0)
	for _, item := range allItems {
		_, alreadyExist := h[item]
		if !alreadyExist {
			nonOccurring = append(nonOccurring, item)
		}
	}
	return nonOccurring
}

func mergeOutputWithHistory(pathSpec, historyFile string) ([]string, error) {
	output := listPathSpecContents(pathSpec)
	historyMap, err := historyFromFile(historyFile)
	if err != nil {
		return make([]string, 0), fmt.Errorf("can't build history from history file %s: %s", historyFile, err)
	}

	historyMap.eliminateStaleItems(output)

	orderedItems := getOrderedItems(historyMap)
	itemsNotInHistory := getNonOccurring(historyMap, output)
	return append(orderedItems, itemsNotInHistory...), nil
}

func stripOutput(item string, componentsToShow int) string {
	if componentsToShow == 0 {
		return item
	}
	components := strings.Split(item, "/")
	components = mare.Filter(components, func(s string) bool {
		return s != ""
	})
	numComponents := len(components)
	if numComponents <= componentsToShow {
		return item
	}
	stripFrom := numComponents - componentsToShow
	return path.Join(components[stripFrom:]...)
}

func listPathContentsWithHistory(pathSpec, historyFile string, numComponentsShown int) {
	items, err := mergeOutputWithHistory(pathSpec, historyFile)
	mare.PanicIfErr(err)
	for _, item := range items {
		stripped := stripOutput(item, numComponentsShown)
		fmt.Println(stripped)
	}
}

func main() {
	arg.MustParse(&args)
	if args.Selection == "" {
		listPathContentsWithHistory(args.PathSpec, args.HistoryFile, args.NumPathComponentsShown)
	} else {
		addToHistory(args.HistoryFile, args.Selection)
	}
}
