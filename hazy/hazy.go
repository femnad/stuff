package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const Home = "HOME"
const HostNameKey = "HostName"
const Name = "hazy"
const SshConfigFilePath = "~/.ssh/config"
const Tilde = "~"

type MapOfMapOfStrings map[string]map[string]string
type MapOfStrings map[string]string
type MapOfStringArrays map[string][]string

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func expandhome(path string) string {
	home := os.Getenv(Home)
	return strings.Replace(path, Tilde, home, 1)
}

func openWithHomeExpand(path string) (*os.File, error) {
	return os.Open(expandhome(path))
}

func createWithHomeExpand(path string) (*os.File, error) {
	return os.Create(expandhome(path))
}

func startsWith(s string, sub string) bool {
	return 0 == strings.Index(s, sub)
}

func getHostFromLine(line string) string {
	trimmedLine := strings.TrimSpace(line)
	split := strings.Split(trimmedLine, " ")
	if len(split) != 2 {
		message := fmt.Sprintf("Unexpected host line %s", trimmedLine)
		panic(message)
	}
	return split[1]
}

func getConfigLineKeyAndValue(trimmedLine string) (string, string) {
	indexOfFirstSpace := strings.Index(trimmedLine, " ")

	if indexOfFirstSpace < 0 {
		panic("No space in line")
	}

	key := trimmedLine[:indexOfFirstSpace]
	rest := trimmedLine[indexOfFirstSpace+1:]
	value := strings.TrimSpace(rest)

	return key, value
}

func isHostLine(line string) bool {
	return startsWith(line, "Host ")
}

func getSshConfigAsMap() (MapOfMapOfStrings, []string, MapOfStringArrays) {
	sshConfigFile, err := openWithHomeExpand(SshConfigFilePath)
	check(err)
	defer sshConfigFile.Close()

	reader := bufio.NewReader(sshConfigFile)

	sshConfigMap := make(MapOfMapOfStrings)
	configOrderMap := make(MapOfStringArrays)

	var hostConfig MapOfStrings
	var currentHost string
	var hosts []string
	var configOrder []string

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			sshConfigMap[currentHost] = hostConfig
			configOrderMap[currentHost] = configOrder
			break
		}
		trimmedLine := strings.TrimSpace(line)
		if isHostLine(line) {
			if len(hostConfig) > 0 {
				sshConfigMap[currentHost] = hostConfig
				configOrderMap[currentHost] = configOrder
				configOrder = make([]string, 0)
			}

			currentHost = getHostFromLine(line)

			hosts = append(hosts, currentHost)
			hostConfig = make(MapOfStrings)
		} else if trimmedLine != "" {
			key, value := getConfigLineKeyAndValue(trimmedLine)
			hostConfig[key] = value
			configOrder = append(configOrder, key)
		}
	}

	return sshConfigMap, hosts, configOrderMap
}

func isNewHost(sshConfigMap MapOfMapOfStrings, host string) bool {
	hostConfig := sshConfigMap[host]
	return len(hostConfig) == 0
}

func sshConfigMapToFile(sshConfigMap MapOfMapOfStrings, hosts []string, configOrderMap MapOfStringArrays) {
	sshConfigFile, err := createWithHomeExpand(SshConfigFilePath)
	check(err)
	defer sshConfigFile.Close()

	writer := bufio.NewWriter(sshConfigFile)

	numberOfHosts := len(hosts)
	for _, host := range hosts {
		hostLine := fmt.Sprintf("Host %s\n", host)
		writer.WriteString(hostLine)
		numberOfHosts--

		config := sshConfigMap[host]
		configKeys, oldHost := configOrderMap[host]

		if !oldHost {
			for key := range config {
				configKeys = append(configKeys, key)
			}
		}

		for _, key := range configKeys {
			value := config[key]
			configLine := fmt.Sprintf("    %s %s\n", key, value)
			writer.WriteString(configLine)
		}
		if numberOfHosts > 0 {
			writer.WriteString("\n")
		}
	}

	writer.Flush()
}

func getHostValue(sshConfigMap MapOfMapOfStrings, host string, key string) (string, error) {
	hostConfig, ok := sshConfigMap[host]
	if !ok {
		return "", fmt.Errorf("Not found: Host %s", host)
	}

	configValue, ok := hostConfig[key]
	if !ok {
		return "", fmt.Errorf("Not found: Key %s for Host %s", key, host)
	}

	return configValue, nil
}

func updateConfigMap(sshConfigMap MapOfMapOfStrings, host string, key string, value string) MapOfMapOfStrings {
	hostConfig, ok := sshConfigMap[host]

	if !ok {
		hostConfig = make(MapOfStrings)
	}

	hostConfig[key] = value
	sshConfigMap[host] = hostConfig

	return sshConfigMap
}

func printUsage() {
	fmt.Printf("usage: %s <host> <hostname>\n", Name)
}

func setHostNameForHost(host string, hostName string) {
	sshConfigMap, hosts, configOrderMap := getSshConfigAsMap()

	if isNewHost(sshConfigMap, host) {
		hosts = append(hosts, host)
	}

	updatedConfigMap := updateConfigMap(sshConfigMap, host, HostNameKey, hostName)

	sshConfigMapToFile(updatedConfigMap, hosts, configOrderMap)
}

func main() {
	args := os.Args[1:]

	if len(args) != 2 {
		printUsage()
		os.Exit(1)
	}

	host := args[0]
	hostName := args[1]

	setHostNameForHost(host, hostName)
}
