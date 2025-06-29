package manager

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
)

func StartDownloading(URI string, start, end int64, file *os.File,wg *sync.WaitGroup){

	defer wg.Done()
	req,err:=http.NewRequest("GET",URI,nil)
	if err!=nil{

	}

	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))
	client:=&http.Client{}
	resp,err:=client.Do(req)
	  if err != nil {
        log.Printf("[Seg %d-%d] HTTP error: %v", start, end, err)
        return
    }
    defer resp.Body.Close()
	buf:=make([]byte,32*1024)
	var offset=start
    for {
       n,err:=resp.Body.Read(buf)
	   if n>0{
		 if _, werr := file.WriteAt(buf[:n], offset); werr != nil {
                log.Printf("Write error: %v", werr)
                return
            }
            offset += int64(n)
	   }
	    if err != nil {
            if err != io.EOF {
                log.Printf("Read error: %v", err)
            }
            break
        }
    }

}

func Download(size int64,url string){
	const parts=2
	partSize:=size/parts
	 outFile := "output.file"
    file, err := os.Create(outFile)
    if err != nil {
        log.Fatal("File create error:", err)
    }
    defer file.Close()

	 var wg sync.WaitGroup
	 for i := 0; i < parts; i++ {
		fmt.Println("goroutine %d is started",i)
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
