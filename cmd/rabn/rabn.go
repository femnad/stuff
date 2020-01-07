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
	HistoryFile string `arg:"-H,required"`
	PathSpec []string `arg:"-p,required,separate"`
	Selection string `arg:"positional" default:""`
}

type historyItems map[string]int

type history struct {
	Items  historyItems
	Prefix string
	historyFile string
}

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
	if len(h.Items) == 0 {
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
	h.Items[selection]++
}

func (h *history) eliminateStaleItems(listOutput []string) {
	for itemKey := range h.Items {
		canonicalItem := h.canonicalizeItem(itemKey)
		if !mare.Contains(listOutput, canonicalItem) {
			delete(h.Items, itemKey)
		}
	}
}

func (h *history) canonicalizeItem(item string) string {
	return mare.ExpandUser(path.Join(h.Prefix, item))
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

func listPathSpecContents(paths []string) []string {
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
	for item, count := range h.Items {
		items := orderedMap[count]
		orderedMap[count] = append(items, item)
	}

	counts := make([]int, len(orderedMap))
	for count := range orderedMap {
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

func initHistory(historyFile, prefix string) (h history, err error) {
	file, err := os.OpenFile(historyFile, os.O_CREATE|os.O_WRONLY, filePermissions)
	if err != nil {
		return h, fmt.Errorf("error creating history file: %s", err)
	}
	err = file.Close()
	if err != nil {
		return h, fmt.Errorf("error closing history file: %s", err)
	}
	h = history{Items:make(historyItems), Prefix: prefix, historyFile:historyFile}
	return
}

func getHistory(historyFile, prefix string) (history, error) {
	h := history{historyFile:historyFile, Prefix:prefix}
	err := h.deserialize(historyFile)
	if err != nil {
		return h, err
	}
	if h.Items == nil {
		h.Items = make(historyItems)
	}
	return h, err
}

func historyFromFile(historyFile, prefix string) (history, error) {
	historyFile = os.ExpandEnv(strings.Replace(historyFile, "~", "$HOME", 1))
	_, err := os.Stat(historyFile)
	if os.IsNotExist(err) {
		return initHistory(historyFile, prefix)
	} else if err != nil && !os.IsNotExist(err) {
		return history{}, err
	}
	return getHistory(historyFile, prefix)
}

func addToHistory(h history, selection string) {
	h.addToHistory(selection)
	err := h.serialize(h.historyFile)
	mare.PanicIfErr(err)
}

func getNonOccurring(h history, allItems []string) []string {
	nonOccurring := make([]string, 0)
	for _, item := range allItems {
		_, alreadyExist := h.Items[item]
		if !alreadyExist {
			nonOccurring = append(nonOccurring, item)
		}
	}
	return nonOccurring
}

func mergeOutputWithHistory(h history, paths []string) ([]string, error) {
	output := listPathSpecContents(paths)

	h.eliminateStaleItems(output)

	orderedItems := getOrderedItems(h)
	itemsNotInHistory := getNonOccurring(h, output)
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

func listPathContentsWithHistory(h history, paths []string, prefix string) {
	items, err := mergeOutputWithHistory(h, paths)
	mare.PanicIfErr(err)

	for _, item := range items {
		stripped := strings.TrimPrefix(item, prefix)
		fmt.Println(stripped)
	}
}

func main() {
	arg.MustParse(&args)
	prefix := findLongestCommonPrefix(args.PathSpec)
	h, err := historyFromFile(args.HistoryFile, prefix)
	mare.PanicIfErr(err)
	if args.Selection == "" {
		listPathContentsWithHistory(h, args.PathSpec, prefix)
	} else {
		addToHistory(h, args.Selection)
	}
}
