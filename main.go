package ucd

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"strings"
	"time"
)

var (
	helpFlag        bool
	clearFlag       bool
	repeatFlag      int
	listFlag        bool
	listStashFlag   bool
	historyPathFlag int
	stashFlag       bool
	cachePath       string
	cacheFile       *os.File

	invalidPath bool = false
)

func main() {
	log.Printf("ucd-v0.1\n")

	// flags
	flag.BoolVar(&helpFlag, "h", false, "display help list")
	flag.BoolVar(&clearFlag, "c", false, "clear history list")
	flag.BoolVar(&listFlag, "l", false, "MRU list for recently used cd commands")
	flag.BoolVar(&listStashFlag, "ls", false, "list stashed cd commands")
	flag.BoolVar(&stashFlag, "s", false, "stash cd path into a separate list")
	flag.IntVar(&repeatFlag, "r", 1, "repeat dynamic cd path (for ..)")
	flag.IntVar(&historyPathFlag, "p", 0, "execute the # path listed from MRU list")
	flag.Parse()

	args := flag.Args()
	// log.Printf("args: %v\n", args)

	homeDir, _ := os.UserHomeDir()
	// curDir, _ := os.Getwd()

	// log.Printf("cwd: %v\n", curDir)

	if helpFlag {
		fmt.Print(".")
		os.Exit(1)
	}

	cachePath = homeDir + "/.ucd-cache"
	cacheFile, _ := os.Open(cachePath)
	byteValue, _ := ioutil.ReadAll(cacheFile)

	var r Records
	err := json.Unmarshal(byteValue, &r)
	if err != nil {
		r = Records{
			PathRecords:  map[string]PathRecord{},
			StashRecords: map[string]StashRecord{},
		}
	}

	if clearFlag {
		r = Records{
			PathRecords:  map[string]PathRecord{},
			StashRecords: map[string]StashRecord{},
		}
		output, _ := json.Marshal(r)
		ioutil.WriteFile(cachePath, output, 0644)
		fmt.Print(".")
		os.Exit(1)
	}

	// exit earlier depending on flag passed in
	if listFlag {
		r.listRecords("path")
		fmt.Print(".")

		os.Exit(1)
	}

	if listStashFlag {
		r.listRecords("stash")
		fmt.Print(".")

		os.Exit(1)
	}

	if len(args) > 1 {
		log.Fatalln("Only < 1 arguments can be passed to ucd")
	}

	// fmt.Print sends output to stdout, this will be consumed by builtin `cd` command

	var targetPath string
	if historyPathFlag > 0 {
		mruRecords := sortRecords(r.PathRecords)
		targetPath = mruRecords[historyPathFlag-1]
	} else {
		if len(args) > 0 {
			targetPath = repeat(args[0], repeatFlag)
		} else {
			targetPath = homeDir
		}
	}
	// log.Printf("targetPath: %v\n", targetPath)

	// attempt to chdir into target path
	err = os.Chdir(targetPath)
	if err != nil {
		invalidPath = true
	}

	if invalidPath {
		fmt.Print(targetPath)
		os.Exit(1)
	}

	targetPath, _ = os.Getwd()

	rec, ok := r.PathRecords[targetPath]
	if ok {
		rec.Count++
		rec.Timestamp = timeNow()
		r.PathRecords[targetPath] = rec
	} else {
		r.PathRecords[targetPath] = PathRecord{Count: 1, Timestamp: timeNow()}
	}

	fmt.Print(targetPath)
	if stashFlag {
		r.StashRecords[targetPath] = StashRecord{Timestamp: timeNow()}
	}

	output, _ := json.Marshal(r)
	ioutil.WriteFile(cachePath, output, 0644)
	cacheFile.Close()
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

func getType(v interface{}) string {
	return reflect.TypeOf(v).String()
}
