// Package google provides a function to do Google searches using the Google Web
// Search API. See https://developers.google.com/web-search/docs/
//
// This package is an example to accompany https://blog.golang.org/context.
// It is not intended for use by others.
//
// Google has since disabled its search API,
// and so this package is no longer useful.
package google

import (
	"context"
	"net/http"

	"helloworld/userip"
)

// Results is an ordered list of search results.
type Results []Result

// A Result contains the title and URL of a search result.
type Result struct {
	Title, URL string
}

// Search sends query to Google search and returns the results.
func Search(ctx context.Context, query string) (string, error) {
	// Prepare the Google Search API request.
	req, err := http.NewRequest("GET", "https://developers.google.com/custom-search", nil)
	if err != nil {
		return "", err
	}
	q := req.URL.Query()
	q.Set("q", query)

	// If ctx is carrying the user IP address, forward it to the server.
	// Google APIs use the user IP to distinguish server-initiated requests
	// from end-user requests.
	if userIP, ok := userip.FromContext(ctx); ok {
		q.Set("userip", userIP.String())
	}
	req.URL.RawQuery = q.Encode()

	// Issue the HTTP request and handle the response. The httpDo function
	// cancels the request if ctx.Done is closed.
	res := make([]byte, 10240000)
	err = httpDo(ctx, req, func(resp *http.Response, err error) (string,error) {
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		resp.Body.Read(res)
		return string(res), nil
	})
	// httpDo waits for the closure we provided to return, so it's safe to
	// read results here.
	return string(res), err
}

// httpDo issues the HTTP request and calls f with the response. If ctx.Done is
// closed while the request or f is running, httpDo cancels the request, waits
// for f to exit, and returns ctx.Err. Otherwise, httpDo returns f's error.
func httpDo(ctx context.Context, req *http.Request, f func(*http.Response, error) (string,error)) error {
	// Run the HTTP request in a goroutine and pass the response to f.
	c := make(chan error, 1)
	req = req.WithContext(ctx)
	go func() {
		_, err := f(http.DefaultClient.Do(req))
		c <- err
	}()
	select {
	case <-ctx.Done():
		<-c // Wait for f to return.
		return ctx.Err()
	case err := <-c:
		return err
	}
}
