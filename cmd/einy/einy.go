package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-ini/ini"
)

const (
	name          = "einy"
	InventoryFile = "~/.ansible-inventory/hosts"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func addHostToAnsibleInventory(group string, host string) {
	ini.PrettyFormat = false

	home := os.Getenv("HOME")
	inventoryFile := strings.Replace(InventoryFile, "~", home, 1)

	cfg, err := ini.LoadSources(ini.LoadOptions{AllowBooleanKeys: true}, inventoryFile)

	groupSection, err := cfg.GetSection(group)
	if err != nil {
		groupSection, err = cfg.NewSection(group)
		check(err)
	}

	_, err = groupSection.NewBooleanKey(host)
	check(err)

	err = cfg.SaveTo(inventoryFile)
	check(err)
}

func printUsage() {
	fmt.Printf("usage: %s <group> <hostname>\n", name)
}

func main() {
	argv := os.Args
	if len(argv) != 3 {
		printUsage()
		os.Exit(1)
	}

	args := argv[1:]
	group := args[0]
	host := args[1]

	addHostToAnsibleInventory(group, host)
}
