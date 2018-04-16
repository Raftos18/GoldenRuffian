// Package collectlinks does extraordinarily simple operation of parsing a given piece of html
// and providing you with all the hyperlinks hrefs it finds.
package collectlinks

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

// Record is the data structure that contains the data extracted from pages
// TODO: Move this out of here
type Record struct {
	Name       string
	Telephones []string
}

func (r Record) String() string {
	name := r.Name
	phones := ""

	for i := 0; i < len(r.Telephones); i++ {
		phones += r.Telephones[i] + ", "
	}

	return fmt.Sprintf("Name: %s, Telephones: [%s]", name, phones)
}

// All takes a reader object (like the one returned from http.Get())
// It returns a slice of strings representing the "href" attributes from
// anchor links found in the provided html.
// It does not close the reader passed to it.
func All(httpBody io.Reader) []string {
	links := []string{}
	col := []string{}
	page := html.NewTokenizer(httpBody)

	for {
		tokenType := page.Next()
		if tokenType == html.ErrorToken {
			return links
		}
		token := page.Token()
		if tokenType == html.StartTagToken && token.DataAtom.String() == "a" {
			for _, attr := range token.Attr {
				if attr.Key == "href" {
					tl := trimHash(attr.Val)
					col = append(col, tl)
					resolv(&links, col)
				}
			}
		}
	}
}

func Names(httpBody io.Reader) []string {
	names := []string{}
	col := []string{}
	page := html.NewTokenizer(httpBody)

	for {
		tokenType := page.Next()
		if tokenType == html.ErrorToken {
			return names
		}
		token := page.Token()
		if tokenType == html.StartTagToken && token.DataAtom.String() == "a" {
			for _, attr := range token.Attr {
				if attr.Key == "data-event" && attr.Val == "list.profile.name" {
					page.Next()
					page.Next()

					token = page.Token()
					tl := token.String()
					col = append(col, tl)

					//fmt.Printf("Name: %v\r\n", tl)
					resolv(&names, col)
				}
			}
		}
	}
}

func NamesAndPhones(httpBody io.Reader) []Record {
	index := 0
	records := []Record{}
	page := html.NewTokenizer(httpBody)

	var phones []string

	curRecord := Record{}
	for {

		tokenType := page.Next()
		if tokenType == html.ErrorToken {

			// We add the telephones to the last record here because it was not possible in the nomarl flow of the program
			for i := 1; i < len(phones); i++ {
				records[index-1].Telephones = append(records[index-1].Telephones, phones[i])
			}

			return records
		}

		token := page.Token()
		if tokenType == html.StartTagToken && token.DataAtom.String() == "a" {
			for _, attr := range token.Attr {
				if attr.Key == "data-event" && attr.Val == "list.profile.name" {

					// Advance two token to get to the one we want
					page.Next()
					page.Next()

					token = page.Token()
					tl := token.String()

					// We get the current record name and save it
					curRecord.Name = tl

					// If index is bigger than 0 we add the phones we gather in the last iteration (see below) to the previous record
					if index > 0 {
						// BUG: This should be 0 but it return two times the same phone if it is, so we leave it at 1!
						for i := 1; i < len(phones); i++ {
							records[index-1].Telephones = append(records[index-1].Telephones, phones[i])
						}
					}

					// Append current record to records
					records = append(records, curRecord)

					// Clear phones to prepare to for the next Record
					phones = []string{}
					// Append the index to indicate a new name was found thus a new record begins
					index++
				}
			}
		}

		if tokenType == html.StartTagToken && token.DataAtom.String() == "a" {
			for _, attr := range token.Attr {
				if attr.Key == "href" {
					tl := trimHash(attr.Val)
					if strings.HasPrefix(tl, "tel:") {
						tel := strings.Split(tl, ":")
						phones = append(phones, tel[1])
					}
				}
			}
		}

	}
}

// trimHash slices a hash # from the link
func trimHash(l string) string {
	if strings.Contains(l, "#") {
		var index int
		for n, str := range l {
			if strconv.QuoteRune(str) == "'#'" {
				index = n
				break
			}
		}
		return l[:index]
	}
	return l
}

// check looks to see if a url exits in the slice.
func check(sl []string, s string) bool {
	var check bool
	for _, str := range sl {
		if str == s {
			check = true
			break
		}
	}
	return check
}

// resolv adds links to the link slice and insures that there is no repetition
// in our collection.
func resolv(sl *[]string, ml []string) {
	for _, str := range ml {
		if check(*sl, str) == false {
			*sl = append(*sl, str)
		}
	}
}
