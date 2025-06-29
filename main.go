package main

import (
	"fmt"

	downloadtype "github.com/mirdha8846/IDM.git/downloadType"
	"github.com/mirdha8846/IDM.git/manager"
	// "log"
	// "net/http"
)

func main() {
    var URI string
    fmt.Print("Enter your URI: ")
    fmt.Scan(&URI)

	parallelDownloadExist, size,err := downloadtype.Check(URI)
	if err != nil {
		return
	}
	if !parallelDownloadExist {
		fmt.Println("This URL doesn't support parallel downloading")
	}

	if parallelDownloadExist{
		manager.Download(size,URI)
	}
  

    
}
