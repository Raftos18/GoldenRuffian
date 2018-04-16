package collectlinks

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	SleepDuration = 1

	//XO ΧρυσοςΟδηγος
	XO = iota
	//VR Vrisko.gr
	VR
)

// SearchSite The site you want goldenRuffian to search
var SearchSite int

// SetSearchSite set the website you want to search
func SetSearchSite(searchSite string) error {
	// Assign in what site to search
	switch searchSite {
	case "XO":
		SearchSite = XO
		return nil
	case "VR":
		SearchSite = VR
		return nil
	default:
		return errors.New("Error getArguments: Wrong search site provided. Choose between [XO/VR]")
	}
}

// FetchPagesAsync fetches html page based on what, where and page number and puts them in a channel
func FetchPagesAsync(searchWord string, location string, pages string) chan http.Response {
	var wg sync.WaitGroup

	out := make(chan http.Response)
	l, _ := strconv.Atoi(pages)

	URLs := make([]string, l)
	URLs = gatherURLs(searchWord, location, l, &URLs)

	for i := 0; i < len(URLs); i++ {

		// Put thread to sleep for the duration specified
		time.Sleep(SleepDuration)

		wg.Add(1)
		go func(URL string) {
			out <- fetchAsync(URL)
			wg.Done()
		}(URLs[i])
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

// CollectNamesAsync searches the pages provided in the channel for names
func CollectNamesAsync(in chan http.Response) chan string {
	var wg sync.WaitGroup
	out := make(chan string)
	for html := range in {
		wg.Add(1)
		go func(h http.Response) {
			defer h.Body.Close()

			names := Names(h.Body)
			for i := 0; i < len(names); i++ {
				out <- names[i]
			}
			wg.Done()
		}(html)
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

// CollectNamesAsync searches the pages provided in the channel for names
func CollectNamesAndPhonesAsync(in chan http.Response) chan string {
	var wg sync.WaitGroup
	out := make(chan string)
	for html := range in {
		wg.Add(1)
		go func(h http.Response) {
			defer h.Body.Close()

			namesAndPhones := NamesAndPhones(h.Body)
			for i := 0; i < len(namesAndPhones); i++ {
				out <- namesAndPhones[i].String()
			}
			wg.Done()
		}(html)
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

// CollectMailsAsync searches the pages provided in the channel for telephones
func CollectPhonesAsync(in chan http.Response) chan string {
	var wg sync.WaitGroup
	out := make(chan string)

	for html := range in {
		wg.Add(1)
		go func(h http.Response) {
			defer h.Body.Close()

			links := All(h.Body)
			for i := 0; i < len(links); i++ {
				if strings.HasPrefix(links[i], "tel:") {
					mail := strings.Split(links[i], ":")
					out <- mail[1]
				}
			}
			wg.Done()
		}(html)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

// CollectMailsAsync searches the pages provided in the channel for mails
func CollectMailsAsync(in chan http.Response) chan string {
	var wg sync.WaitGroup
	out := make(chan string)

	for html := range in {
		wg.Add(1)
		go func(h http.Response) {
			defer h.Body.Close()

			links := All(h.Body)
			for i := 0; i < len(links); i++ {
				if strings.HasPrefix(links[i], "mailto:") {
					mail := strings.Split(links[i], ":")
					out <- mail[1]
				}
			}
			wg.Done()
		}(html)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

func formatWord(word string) string {
	if strings.Contains(word, "-") {
		wordsplited := strings.Split(word, "-")
		word = wordsplited[0] + "+" + wordsplited[1]
	}
	return word
}

func gatherURLs(searchWord string, location string, pages int, urls *[]string) []string {
	pagesRemaining := pages
	urlSlice := make([]string, pagesRemaining)
	i := 0
	for pagesRemaining > 0 {
		page := strconv.Itoa(pagesRemaining)
		urlSlice[i] = formatURL(searchWord, location, page)
		pagesRemaining--
		i++
	}
	return urlSlice
}

func formatURL(searchWord string, location string, pages string) string {
	var URL string

	switch SearchSite {
	case XO:
		searchWord = formatWord(searchWord)
		URL = fmt.Sprintf("https://www.xo.gr/search/?what=%s&where=%s&page=%s", searchWord, location, pages)
	case VR:
		URL = fmt.Sprintf("http://www.vrisko.gr/search/%s/%s/?page=%s", searchWord, location, pages)
	}

	//fmt.Printf("SearchSite: %d\n", SearchSite)

	return URL
}

// fetch data
func fetch(url string, ch chan http.Response) {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetch: %v. There are not so many pages try less\n", err)
		os.Exit(1)
	}
	fetchInfoMsg(url)
	ch <- *resp
}

func fetchAsync(url string) http.Response {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetch: %v. There are not so many pages try less\n", err)
		os.Exit(1)
	}
	fetchInfoMsg(url)
	return *resp
}

func fetchInfoMsg(url string) {
	fmt.Printf("[--- Fetching data from: %s ---]\r\n\r\n", url)
}
