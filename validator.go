package main

import (
	"errors"
	"html"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
)

type URLInfo struct {
	title string
	err   error
}

func ValidateURL(url string) URLInfo {
	regex := regexp.MustCompile("<title>(.*)</title>")
	log.Printf("Validating %s", url)
	res, err := http.Get(url)
	if err != nil {
		log.Printf("Error %s", err.Error())
		return URLInfo{title: "", err: err}
	}

	defer res.Body.Close()

	log.Printf("Returned status code: %d", res.StatusCode)
	if (res.StatusCode < 200) || (res.StatusCode > 299) {
		return URLInfo{title: "", err: errors.New("URL Not found")}
	}

	limitedReader := io.LimitReader(res.Body, 8*1024)
	page, err := ioutil.ReadAll(limitedReader)
	if err != nil {
		return URLInfo{title: "", err: err}
	}

	matches := regex.FindStringSubmatch(string(page))
	var title string

	title = url
	if len(matches) > 1 {
		title = matches[1]
	}

	title = html.UnescapeString(title)
	return URLInfo{title: title, err: nil}
}
