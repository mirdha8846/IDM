package manager

import (
	"fmt"
    "io"
    "log"
    "net/http"
   
    "os"
   
    "sync"
    "time"
)


func StartDownloading(URI string, start, end int64, file *os.File, wg *sync.WaitGroup) {
    defer wg.Done()
    const maxRetries = 3
    var attempt int

    

    for attempt = 1; attempt <= maxRetries; attempt++ {
        req, err := http.NewRequest("GET", URI, nil)
        if err != nil {
            log.Printf("[%s: %d-%d] Request creation error: %v", start, end, err)
            return
        }

        req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))
        client := &http.Client{}
        resp, err := client.Do(req)
        if err != nil {
            log.Printf("[%s: %d-%d] HTTP error (attempt %d/%d): %v",  start, end, attempt, maxRetries, err)
            time.Sleep(time.Second * 2)
            continue
        }
        defer resp.Body.Close()

        buf := make([]byte, 32*1024)
        offset := start
        for {
            n, err := resp.Body.Read(buf)
            if n > 0 {
                if _, werr := file.WriteAt(buf[:n], offset); werr != nil {
                    log.Printf("[%s: %d-%d] Write error: %v",  start, end, werr)
                    return
                }
                offset += int64(n)
            }
            if err != nil {
                if err != io.EOF {
                    log.Printf("[%s: %d-%d] Read error: %v",start, end, err)
                    time.Sleep(time.Second * 2)
                    break
                }
                return // success
            }
        }
    }

    if attempt > maxRetries {
        log.Printf("[%s: %d-%d] Failed after %d attempts",  start, end, maxRetries)
    }
}


func Download(size int64,url string,filename string){
	const parts=8
	partSize:=size/parts
	 outFile := filename
    file, err := os.Create(outFile)
    if err != nil {
        log.Fatal("File create error:", err)
    }
    defer file.Close()

	 var wg sync.WaitGroup
	 for i := 0; i < parts; i++ {
		fmt.Printf("goroutine started :%d",i)
        start := partSize * int64(i)
        end := start + partSize - 1
        if i == parts-1 {
            end = size - 1
        }
        wg.Add(1)
        go StartDownloading(url, start, end, file, &wg)
    }

    wg.Wait()
    fmt.Println("✅ Download finished!")

}
