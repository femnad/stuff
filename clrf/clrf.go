package main

import (
	"flag"
	"fmt"
	"os/exec"
)

const (
	defaultSource      = "primary"
	defaultDestination = "clipboard"
	defaultSelections  = 25
)

var validSelections = map[string]struct{}{"primary": {}, "clipboard": {}}

func validateSelections(selections ...string) {
	for _, selection := range selections {
		_, validSelection := validSelections[selection]
		if !validSelection {
			errorMassage := fmt.Sprintf("%s is not a valid selection\n", selection)
			panic(errorMassage)
		}
	}
}

func getParameters() (string, string, int) {
	source := flag.String("source", defaultSource, "Selection for source")
	destination := flag.String("destination", defaultDestination, "Selection for destination")
	numberOfSelections := flag.Int("selections", defaultSelections, "Number of selections")
	flag.Parse()
	return *source, *destination, *numberOfSelections
}

func determineClipsterSource(source string) string {
	return string(source[0])
}

func showSelectionMenu(source, destination string, selections int) {
	clipsterSource := determineClipsterSource(source)
	commandString := fmt.Sprintf("clipster -o%s -n %d -0 | "+
		"rofi -dmenu -matching fuzzy -sep '\\0' -p '%s to %s: ' | "+
		"xclip -i -selection %s", clipsterSource, selections, source, destination, destination)
	command := exec.Command("bash", "-c", commandString)
	err := command.Run()
	if err != nil {
		panic(err)
	}
}

func main() {
	source, destination, selections := getParameters()
	validateSelections(source, destination)
	showSelectionMenu(source, destination, selections)
}
