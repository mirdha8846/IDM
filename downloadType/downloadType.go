package downloadtype
import (
    "fmt"
    "log"
    "net/http"
)

func Check(URI string) (bool, int64,error) {
	resp, err := http.Head(URI)
	if err != nil {
		return false, 0,fmt.Errorf("HEAD request failed: %w", err)
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
		return true, size,nil
	}
	fmt.Printf("❌ Server does not support byte ranges (Accept-Ranges: %q); fallback to single-threaded download.\n", ar)
	return false, size,nil
}
