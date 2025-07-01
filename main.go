package main

import (
	"fmt"
	"log"

	downloadtype "github.com/mirdha8846/IDM.git/downloadType"
	"github.com/mirdha8846/IDM.git/manager"
	// "log"
	// "net/http"
)

func main() {
	var URI string
	fmt.Print("Enter your URI: ")
	fmt.Scan(&URI)
	speedUp, err := downloadtype.SupportsRanges(URI)
	if err != nil {
		log.Fatal("error")
	}
	if !speedUp {
		fmt.Print("cant be speed up")
	}

	// var urls []string
	// urlQueue:=make(chan Task,5)

	// for _,uri:=range(urls){

	parallelDownloadExist, size,filename ,err := downloadtype.Check(URI)
	if err != nil {
		return 
	}
	//     task:=Task{
	// 		URL: uri,
	// 		Size: size,

	// 	}
	// 	urlQueue<-task

	if !parallelDownloadExist {
		fmt.Println("cant be speedup")
		speedUp, err := downloadtype.SupportsRanges(URI)
		if err == nil && speedUp {
            fmt.Println("value of speedUP:%s",speedUp)
			fmt.Println("This URL support parallel downloading")
		}
	}

	if parallelDownloadExist{

	manager.Download(size,URI,filename)
	}
	// }

	// for task:=range(urlQueue){

	// }

}
