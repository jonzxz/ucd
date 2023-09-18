package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jedib0t/go-pretty/table"
)

var (
	repeatFlag int
	listFlag   bool
	cachePath  string
	cacheFile  *os.File
)

type PathRecords struct {
	Records map[string]PathRecord `json:"paths"`
}

type PathRecord struct {
	Count     int    `json:"count"`
	Timestamp string `json:"ts"`
}

func main() {
	log.Printf("ucd-v0.1\n")

	// flags
	flag.BoolVar(&listFlag, "l", false, "MRU list for recently used cd commands")
	flag.IntVar(&repeatFlag, "r", 1, "repeat dynamic cd path (for ..)")
	flag.Parse()

	args := flag.Args()
	log.Printf("args: %v\n", args)
	log.Printf("listFlag: %v\n", listFlag)

	homeDir, _ := os.UserHomeDir()
	curDir, _ := os.Getwd()

	log.Printf("~: %v\n", homeDir)
	log.Printf("cwd: %v\n", curDir)

	cachePath = homeDir + "/.ucd-cache"

	// cacheFile, _ = os.OpenFile(cachePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	log.Printf("Configuring cachePath to %v\n", cachePath)

	cacheFile, _ := os.Open(cachePath)
	byteValue, _ := ioutil.ReadAll(cacheFile)

	var pr PathRecords
	err := json.Unmarshal(byteValue, &pr)
	if err != nil {
		pr = PathRecords{Records: map[string]PathRecord{}}
	}

	// exit earlier depending on flag passed in
	if listFlag {
		listRecords(pr)
		// dummy data for the time being
		fmt.Print(".")

		os.Exit(1)
	}

	if len(args) > 1 {
		log.Fatalln("Only < 1 arguments can be passed to ucd")
	}

	// fmt.Print sends output to stdout, this will be consumed by builtin `cd` command

	var targetPath string
	if len(args) > 0 {
		targetPath = repeat(args[0], repeatFlag)
	} else {
		targetPath = homeDir
	}

	rec, ok := pr.Records[targetPath]
	if ok {
		rec.Count++
		rec.Timestamp = timeNow()
		pr.Records[targetPath] = rec
	} else {
		pr.Records[targetPath] = PathRecord{Count: 1, Timestamp: timeNow()}
	}

	fmt.Print(targetPath)

	output, _ := json.Marshal(pr)
	ioutil.WriteFile(cachePath, output, 0644)
	cacheFile.Close()
}

func listRecords(pr PathRecords) {
	t := table.NewWriter()
	t.SetOutputMirror(log.Writer())
	t.AppendHeader(table.Row{"#", "path", "count", "timestamp"})

	index := 1
	// TODO: sort via timestamp before displaying
	for path, rec := range pr.Records {
		t.AppendRow([]interface{}{index, path, rec.Count, rec.Timestamp})
		index++
	}

	t.Render()
}

func repeat(str string, times int) string {
	s := make([]string, times)
	for i := range s {
		s[i] = str
	}

	return strings.Join(s, "/")
}

func timeNow() string {
	return time.Now().Format("2006-01-02 15:04:05 MST")
}
