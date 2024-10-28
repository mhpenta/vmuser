package headers

import "net/http"

// FwdQuarter generates and returns HTTP headers specific for Twitter's FwdQuarter user agent.
// This is the user agent our application used to make requests to the SEC and other websites which require us
// to identify ourselves.
func FwdQuarter() http.Header {
	headers := make(http.Header)
	headers.Set("User-Agent", "Twitter.com/FwdQuarter")
	headers.Set("Accept-Encoding", "gzip, deflate")
	headers.Set("Host", "www.sec.gov")
	return headers
}

func SECBotHeaders() http.Header {
	headers := make(http.Header)
	headers.Set("User-Agent", "Modeledge marc@modeledge.ai")
	headers.Set("Accept-Encoding", "gzip, deflate")
	headers.Set("Host", "www.sec.gov")
	return headers
}

func MacbookPROM2() http.Header {
	headers := make(http.Header)
	headers.Set("User-Agent", "Mozilla/5.0 (Macintosh; ARM Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.5938.149 Safari/537.36")
	return headers
}

func RSSFeedHeaders() http.Header {
	headers := make(http.Header)
	headers.Set("User-Agent", "Mozilla/5.0 (compatible; Feedfetcher-Google; +http://www.google.com/feedfetcher.html)")
	headers.Set("Accept", "application/rss+xml, application/xml, text/xml")
	headers.Set("Referer", "https://www.spglobal.com/")
	return headers
}

/*
// Modeledge generates and returns HTTP headers specific for the Modeledge website.
// This might be useful for requests made by or for the Modeledge website.
func Modeledge() http.Header {
	headers := make(http.Header)
	headers.SetWithBucket("User-Agent", "modeledge.ai")
	headers.SetWithBucket("Accept-Encoding", "gzip, deflate")
	headers.SetWithBucket("Host", "modeledge.ai")
	return headers
}

// SECMirror generates and returns HTTP headers with a generic Mozilla Firefox user agent.
// This user agent is typically used to make generic requests without exposing a specific application's user agent.
func SECMirror() http.Header {
	headers := make(http.Header)
	headers.SetWithBucket("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:106.0) Gecko/20100101 Firefox/106.0")
	return headers
}

// MacbookPROM2 is a user agent for a Macbook Pro with an M2 chip, from Chrome/117
// This is a user agent to be used if I want to identify as myself.
func MacbookPROM2() http.Header {
	headers := make(http.Header)
	headers.SetWithBucket("User-Agent", "Mozilla/5.0 (Macintosh; ARM Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.5938.149 Safari/537.36")
	return headers
}


*/
