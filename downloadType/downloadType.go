package downloadtype
import (
    "fmt"
    "log"
    "net/http"
	"net/url"
	"path"
	"strings"
)


func getFilenameFromURL(uri string) string {
    u, err := url.Parse(uri)
    if err != nil {
        return "unknown_filename"
    }
    name := path.Base(u.Path)
    if name == "" || strings.HasSuffix(name, "/") {
        return "index.html"
    }
    return name
}
func Check(URI string) (bool, int64,string,error) {
	resp, err := http.Head(URI)
	filnename:=getFilenameFromURL(URI)
	fmt.Println("http(Head response)=",resp)
	if err != nil {
		return false, 0,filnename,fmt.Errorf("HEAD request failed: %w", err)
	}
	defer resp.Body.Close()

	// Extract Content-Length
	size := resp.ContentLength
	if size < 0 {
		log.Println("⚠️  Server did not provide Content-Length; unknown file size.")
	} else {
		fmt.Printf("File size: %d bytes\n", size)
	}

	// Check Accept-Ranges header
	ar := resp.Header.Get("Accept-Ranges")
	if ar == "bytes" {
		fmt.Println("✅ Server supports byte-range requests: parallel download is possible.")
		return true, size,filnename,nil
	}
	fmt.Printf("❌ Server does not support byte ranges (Accept-Ranges: %q); fallback to single-threaded download.\n", ar)
	return false, size,filnename,nil
}

func SupportsRanges(url string) (bool, error) {
    req, _ := http.NewRequest("GET", url, nil)
    req.Header.Set("Range", "bytes=0-0")
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return false, err
    }
    defer resp.Body.Close()
	fmt.Print(resp.StatusCode)
    return resp.StatusCode == http.StatusPartialContent, nil
}
