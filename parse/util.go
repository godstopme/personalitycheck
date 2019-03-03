package parse

import (
	"io/ioutil"
	"net/http"
	"time"
)

func robustHTTPGet(uri string) (string, error) {
	for {
		time.Sleep(time.Millisecond * 500) // sleeping here because there is a timegap between push & github page availability
		response, err := http.Get(uri)
		if err != nil {
			return "", err
		}

		defer response.Body.Close()

		if response.StatusCode == 404 {
			continue
		}

		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return "", nil
		}

		return string(body), nil
	}
}
