package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/Raftos18/goldenruffian/collectlinks"
	"net/http"
	"time"

	"os"
	"strconv"
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: goldenRuffian searchWord location(optional) pages(optional)\n"+
		"searchWord = The word from which you want to get data\n"+
		"location(optional) = The location in which you want to search (defualt Αθηνα)\n"+
		"pages(optional) = The number of pages you want to get results from (default All)")

	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	flag.Usage = usage
	flag.Parse()

	searchWord, location, pages, err := getArguments()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting arguments: %v\n", err)
		os.Exit(1)
	}
		
	phoneHtmlChan := make(chan http.Response)
	phonesOutChan := make(chan string)

	displayMessage(searchWord, location, pages)
	start := time.Now()
	
	
	phoneHtmlChan = collectlinks.FetchPagesAsync(searchWord, location, pages)
	

	// Searches and saves the phones found
	phonesOutChan = collectlinks.CollectNamesAndPhonesAsync(phoneHtmlChan)
	saveToFile(searchWord + " phones", location, phonesOutChan)


	secs := time.Since(start).Seconds()
	fmt.Printf("Execution time: %v", secs)
}

func getArguments() (string, string, string, error) {
	searchWord := ""
	location := ""
	pages := ""
	args := flag.Args()

	var err error

	if len(args) < 1 {
		return "", "", "", errors.New("Error getArguments: Please provide a search site: XO/VR")
	}
	err = collectlinks.SetSearchSite(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting arguments: %v\n", err)
		os.Exit(1)
	}

	// Ugly code but what can you do here
	if len(args) < 2 {
		return "", "", "", errors.New("Please provide a search word")
	}
	searchWord = args[1]
	if len(args) < 3 {
		location = "Αθήνα"
	} else {
		location = args[2]
	}
	if len(args) < 4 {
		pages = strconv.Itoa(10)
	} else {
		pages = args[3]
	}
	return searchWord, location, pages, err
}

func displayMessage(searchWord string, location string, pages string) {
	ipages, err := strconv.Atoi(pages)
	if err != nil {
		fmt.Println("User must insert valid number of pages")
	}
	if ipages == 0 {
		fmt.Printf("Requesting all pages for %s from %s\n", searchWord, location)
	} else {
		fmt.Printf("Requesting %d pages for %s from %s\n", ipages, searchWord, location)
	}
}

func saveToFile(searchWord string, location string, in chan string) {
	var path string

	switch collectlinks.SearchSite {
	case collectlinks.XO:
		path = "./" + "XO_" + searchWord + "_" + location + ".txt"
	case collectlinks.VR:
		path = "./" + "Vrisko_" + searchWord + "_" + location + ".txt"
	}

	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	fmt.Printf("[-Data-]\n")
	for val := range in {		
		_, err = f.WriteString(val + "\r\n")
		if err != nil {
			panic(err)
		}
		fmt.Println("Record: " + val)
	}
}
