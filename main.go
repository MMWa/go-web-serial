package main

import (
	"fmt"
	"net/http"
	"log"
	"strings"
)

var dev = true

var mainServer http.Server

//streams data to page using server sent events protocol
//stays open until closed
func streamHandler(w http.ResponseWriter, r *http.Request) {
	//connection alive variable
	var alive = true
	//split the path
	path := strings.Split(r.URL.Path, "/")

	//set the header
	//must be text/event-stream , no-cache, and keep-alive
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("server", "now")
	w.Header().Set("status", "200")

	//conduit for when http connection is closed
	notify := w.(http.CloseNotifier).CloseNotify()

	if len(path) == 3 {

		//if the webpage request path is correct length,
		//then open port with the name of last section
		port, err := OpenSerial(path[2])
		//check for any error that occcurs
		if err != nil {
			fmt.Fprintf(w, "Serial Failed to Open Error")
			return
		} else {
			//go routine triggered by returned conduit notify
			go func() {
				<-notify

				fmt.Println("HTTP connection closed.")
				//alive is the connvection status
				alive = false
				return
			}()
			//infinite loop broken if connection closed
			for ; ; {
				//indicates Connection on or off
				if alive {
					//standard states
					//data must begin with "data:"
					//data ends with a \n or 0x00A in hex

					byteData := []byte("data:")
					byteNewLine := []byte{0x00A}

					w.Write(byteData)
					StreamSerialWeb(port, w)
					w.Write(byteNewLine)
					//then we flush each packet
					w.(http.Flusher).Flush()
					fmt.Println("Flushed")

				} else {
					//close port if connection closed
					port.Close()
					break
				}
			}
			fmt.Println("End of Stream!")
			fmt.Println("Port Closed!")
		}
	}
	fmt.Println("leaving function")
}

//returns the stream reciving page
func streamPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "serv/streamPage.html")
}

//list all open serial ports
func listHandler(w http.ResponseWriter, r *http.Request) {

	data, _ := ScanPorts()
	if len(data) != 0 {
		for _, v := range (data) {
			fmt.Fprintf(w, "<p><a href='/streamPage/")
			fmt.Fprintf(w, v)
			fmt.Fprintf(w, "'>"+v+"</a></p>")
		}
	} else {
		fmt.Fprintf(w, "<p>No devices connected</p></br>")
	}
	fmt.Fprintf(w, "<p><a href='/shutdown'>Shutdown</a></p>")
}

//return any page
func pageHandler(w http.ResponseWriter, r *http.Request) {
	if dev {
		fmt.Println(r.URL.Path)
	}
	if (r.URL.Path == "/") {
		//http.ServeFile(w, r, "serv/index.html")
		//if not page list all devices
		listHandler(w, r)
	} else {
		http.ServeFile(w, r, "/serv"+r.URL.Path)
	}
	fmt.Fprintf(w, r.URL.Path[1:])
}

func main() {
	//shutdown server
	http.HandleFunc("/shutdown", func(w http.ResponseWriter, r *http.Request) {
		mainServer.Shutdown(nil)
	})

	http.HandleFunc("/list", listHandler)
	http.HandleFunc("/stream/", streamHandler)
	http.HandleFunc("/streamPage/", streamPage)

	http.HandleFunc("/", pageHandler)

	names, err := ScanPorts()

	if err == nil {
		fmt.Println(names)

	} else {
		fmt.Println("No Port")
	}

	mainServer = http.Server{Addr: ":8081"}

	log.Fatal(mainServer.ListenAndServe())

}
