package main

import (
	"fmt"
	"flag"
	"io"
	"net"
	"net/http"
  	"github.com/wsxiaoys/terminal/color"
  	"strings"
  	"strconv"
  	"code.google.com/p/go.net/websocket"
)

//// file check handler ::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::

type fileCheckHandler struct {
	root http.FileSystem
	regular http.Handler
	index http.Handler
}

func FileCheckServer(root http.FileSystem, regular http.Handler, index http.Handler) http.Handler{
	return &fileCheckHandler{root, regular, index}
}

func (f *fileCheckHandler) ServeHTTP(w http.ResponseWriter, r *http.Request){
	// sanitization : http://golang.org/src/pkg/net/http/fs.go?s=12008:12048#L401
	upath := r.URL.Path
	if !strings.HasPrefix(upath, "/") {
		upath = "/" + upath
		r.URL.Path = upath
	}
	
	// index check
	if r.URL.Path == "/" {
		index, err := f.root.Open("index.html")
		if err != nil {
			if f.index != nil {
				f.index.ServeHTTP(w, r)
			} else {
				fmt.Fprint(w, "This is your Website! No index file or index handler configured.")
			}
		} else {
			d, _ := index.Stat()
			http.ServeContent(w, r, d.Name(), d.ModTime(), index)
		}
		
	// file check
	} else {
		file, err := f.root.Open(upath)
		if err != nil {
			if f.regular != nil {
				f.regular.ServeHTTP(w, r)
			} else {
				http.NotFound(w, r)
			}
		} else {
			d, _ := file.Stat()
			http.ServeContent(w, r, d.Name(), d.ModTime(), file)
		}
	}
}

//// pipe ::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::

var lucyAddress string
var mqttAddress string

func LucyServer(ws *websocket.Conn) {
	tcp, cErr := net.Dial("tcp", lucyAddress)
	if cErr == nil {
		go io.Copy(ws, tcp)
		io.Copy(tcp, ws)
	} else {
		color.Printf( "@r%s\r\n", cErr)
		fmt.Fprintf(ws, "Connection Error: '%s'\r\n",cErr)
	}
}

func MqttBroker(ws *websocket.Conn) {
	color.Printf( "@bHelloMQTT\r\n")
	tcp, cErr := net.Dial("tcp", mqttAddress)
	if cErr == nil {
		go io.Copy(ws, tcp)
		io.Copy(tcp, ws)
	} else {
		color.Printf( "@r%s\r\n", cErr)
		fmt.Fprintf(ws, "Connection Error: '%s'\r\n",cErr)
	}
}

func main(){
	var webroot = flag.String("webroot", "/tmp/", "The folder being served to the web.")
	var port = flag.Int("port", 8080, "The port the server is listening to.")
	var lucy = flag.String("lucy", "localhost:7654", "The tcp address of lucy that should be exposed on websocket.")
	var mqtt = flag.String("mqtt","localhost:1883","The tcp address of the mqtt broker that should be exposed on websocket.")
	flag.Parse()
	lucyAddress = *lucy
	mqttAddress = *mqtt
	color.Printf("@yBRIDGED TCP ADDRESS: %s\r\nWEBROOT: %s\r\nPORT: %d\r\nMQTT Broker: %s\r\n", *lucy, *webroot, *port, *mqtt)
	http.Handle("/mqtt",websocket.Handler(MqttBroker))
	http.Handle("/",FileCheckServer(http.Dir(*webroot), nil, websocket.Handler(LucyServer)))
	http.ListenAndServe(":"+strconv.Itoa(*port), nil)
}