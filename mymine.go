package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"runtime"

	"os/exec"

	"github.com/jessevdk/go-flags"
	"menteslibres.net/gosexy/rest"
)

const (
	version = "1.0"
)

func lookupEnv(key string) string {
	for _, v := range os.Environ() {
		kv := strings.Split(v, "=")
		if kv[0] == key {
			return kv[1]
		}
	}
	return ""
}

type openCmd struct {
	cmd  string
	args []string
}

var m = map[string]openCmd{
	"linux":   {cmd: "xdg-open"},
	"windows": {cmd: "rundll32", args: "url.dll,FileProtocolHandler"},
	"darwin":  {cmd: "open"},
}

func openURLByBrowser(url string) error {
	cmd, ok := m[runtime.GOOS]
	if !ok {
		return errors.New("could not determine how to open URL by browser in this platform")
	}
	args := append(cmd.args, url)
	return exec.Command(cmd.cmd, args...).Start()
}

func showVersion() {
	fmt.Println("mymine version", version)
}

type Options struct {
	Open    []int  `short:"o" long:"open"    description:"Open specified ticket on a web browser"`
	Version []bool `short:"v" long:"version" description:"Show version"`
}

func main() {
	var opts Options
	parser := flags.NewParser(&opts, flags.Default)

	parser.Name = "mymine"
	parser.Usage = "[OPTIONS]"

	_, err := parser.Parse()
	if err != nil {
		fmt.Println("failed to parse cmd line argument. exit.")
		os.Exit(1)
	}

	if opts.Version != nil {
		showVersion()
		os.Exit(1)
	}

	redmineURL := lookupEnv("REDMINE_URL")
	if redmineURL == "" {
		fmt.Println("REDMINE_URL is not specified. exit.")
		os.Exit(1)
	}

	if opts.Open != nil {
		url := redmineURL + "issues/" + strconv.Itoa(opts.Open[0])
		if err := openURLByBrowser(url); err != nil {
			fmt.Println("failed to open URL by browser: %s", err.Error())
			os.Exit(1)
		}
	}

	redmineAPIKey := lookupEnv("REDMINE_API_KEY")
	if redmineAPIKey == "" {
		fmt.Println("REDMINE_API_KEY is not specified. exit.")
		os.Exit(1)
	}

	request := redmineURL + "issues.json?key=" + redmineAPIKey + "&status_id=open&assigned_to_id=me&limit=100"
	fmt.Println("request =", request)
	fmt.Println("fetching information...")

	var buf map[string]interface{}
	rest.Get(&buf, request, nil)

	issues := buf["issues"].([]interface{})
	for _, v := range issues {
		issue := v.(map[string]interface{})

		id := int(issue["id"].(float64))
		status := issue["status"].(map[string]interface{})
		fmt.Printf("[#%d] %11s %s\n", id, status["name"], issue["subject"])
	}
}
