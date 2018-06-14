package main

import (
	"bufio"
	"fmt"
	"strings"

	"flag"

	"github.com/femnad/mare"
)

const (
	linePrefix        = "    "
	SshConfigFilePath = "~/.ssh/config"
)

type hostConfig struct {
	Key   string
	Value string
}

func (c hostConfig) String() string {
	return fmt.Sprintf("%s%s %s", linePrefix, c.Key, c.Value)
}

type directive struct {
	Identifier string
	Value      string
	Config     []hostConfig
}

func (d directive) String() string {
	var containees string
	var formatted []string
	for _, c := range d.Config {
		formatted = append(formatted, c.String())
		containees = strings.Join(formatted, "\n")
	}
	return fmt.Sprintf("%s %s\n%s", d.Identifier, d.Value, containees)
}

func (d *directive) addConfig(key, value string) {
	config := hostConfig{Key: key, Value: value}
	d.Config = append(d.Config, config)
}

func splitTokens(line string) (string, string) {
	trimmedLine := strings.TrimSpace(line)
	if trimmedLine == "" {
		return "", ""
	}
	tokens := strings.SplitN(trimmedLine, " ", 2)
	k := tokens[0]
	v := strings.Join(tokens[1:], " ")
	return k, v
}

func parseDirective(line string) directive {
	k, v := splitTokens(line)
	return directive{Identifier: k, Value: v}
}

func processLineGroup(identifierLine string, configLines ...string) directive {
	h := parseDirective(identifierLine)
	for _, line := range configLines {
		k, v := splitTokens(line)
		h.addConfig(k, v)
	}
	return h
}

func isContainer(line string) bool {
	return strings.HasPrefix(line, "Host")
}

func maybeConsumeGroup(group []string, directives []directive) []directive {
	if len(group) == 0 {
		return directives
	}
	identifier := group[0]
	config := group[1:]
	newDirective := processLineGroup(identifier, config...)

	return append(directives, newDirective)
}

func processConfig(lines []string) []directive {
	directives := make([]directive, 0)
	pDirectives := &directives
	group := make([]string, 0)
	for _, line := range lines {
		if isContainer(line) {
			*pDirectives = maybeConsumeGroup(group, *pDirectives)
			group = make([]string, 0)
		}
		group = append(group, line)
	}
	*pDirectives = maybeConsumeGroup(group, *pDirectives)
	return *pDirectives
}

func readSSHConfig() []string {
	sshConfigFile, err := mare.ExpandUserAndOpen(SshConfigFilePath)
	mare.PanicIfErr(err)
	defer sshConfigFile.Close()

	reader := bufio.NewReader(sshConfigFile)
	lines := make([]string, 0)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		lines = append(lines, line)
	}
	return lines
}

func main() {
	host := flag.String("host", "", "host definition")
	hostName := flag.String("hostname", "", "host name")
	flag.Parse()
	lines := readSSHConfig()
	config := processConfig(lines)
	newHost := true
	for index, l := range config {
		if l.Identifier == "Host" && l.Value == *host {
			fmt.Println("found one")
			pL := &l
			newConfig := hostConfig{Key: "HostName", Value: *hostName}
			separator := hostConfig{"", ""}
			*pL = directive{Identifier: "Host", Value: *host, Config: []hostConfig{newConfig, separator}}
			config[index] = *pL
			newHost = false
		}
	}
	if newHost {
		newConfig := hostConfig{Key: "HostName", Value: *hostName}
		newDirective := directive{Identifier: "Host", Value: *host, Config: []hostConfig{newConfig}}
		config = append(config, newDirective)
	}
	for _, c := range config {
		fmt.Println(c)
	}
}
