package main

import (
	"fmt"
	"net/http"
	"encoding/json"    
	"io/ioutil"
	"os"
	"strings"

	"sync/atomic"

	/*"sync"		
	"os/exec"*/

	"bufio"	
	"io"

	"flag"	
	"log"	
	"time"

	"github.com/gorilla/websocket"
	"github.com/gorilla/mux"
)

type Engine struct {
    Name    string `json:"name"`
    Path    string `json:"path"`
    Config  string `json:"config"`
}

type EngineMessage struct {
    Action     string   `json:"action"`
    Name       string   `json:"name"`
    Command    string   `json:"command"`
    Buffer     string   `json:"buffer"`
    Available  []string `json:"available"`
}

var (
	engines []Engine

	addr               = flag.String("addr", "127.0.0.1:9000", "http service address")

	upgrader           = websocket.Upgrader{
	    ReadBufferSize:  1024,
	    WriteBufferSize: 1024,
	    CheckOrigin: func(r *http.Request) bool {
	        return true
	    },    
	}

	connid  uint64      = 0

	enginews *websocket.Conn
    engineconnid uint64

	process *os.Process = nil

	processw io.Writer  = nil
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Maximum message size allowed from peer.
	maxMessageSize = 8192

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Time to wait before force close on connection.
	closeGracePeriod = 10 * time.Second
)

func getEngines() []Engine {
    raw, err := ioutil.ReadFile("./engines.json")
    if err != nil {
        fmt.Println(err.Error())
        os.Exit(1)
    }

    var c []Engine
    json.Unmarshal(raw, &c)
    return c
}

func createIndex(iname string,ipath string,iconfig string) string {
	var body string

	body+=mainMenu()

	body+="<form method=post action='/change'><table>"
	body+="<tr><td>Name</td><td><input id=name name=name type=text value='"+iname+"'></td>"
	body+="<tr><td>Path</td><td><input id=path name=path type=text value='"+ipath+"'></td>"
	body+="<tr><td>Config</td><td><textarea id=config name=config cols=80 rows=5>"+iconfig+"</textarea></td>"
	body+="<tr><td></td><td><input type=submit value=Submit></td>"
	body+="</table></form>"

	body+="<hr>\n"

	body+="<table border=1 cellpadding=3 cellspacing=3>"
	body+="<tr><td>No.</td><td>Name</td>"
	body+="<td>Path</td><td>Config</td>"
	body+="<td>Edit</td><td>Delete</td></tr>"

	for i, e := range engines {
        body+="<tr>\n"
        body+="<td>"+fmt.Sprintf("%d.",i+1)+"</td>\n"
        body+="<td>"+string(e.Name)+"</td>\n"
        body+="<td>"+string(e.Path)+"</td>\n"
        body+="<td>"+string(e.Config)+"</td>\n"
        body+="<td><form method=post action='/edit'><input type=hidden id=name name=name value='"+e.Name+"''><input type=submit value=Edit></form></td>\n"
        body+="<td><form method=post action='/delete'><input type=hidden id=name name=name  value='"+e.Name+"''><input type=submit value=Delete></form></td>\n"
        body+="</tr>\n"
    }

    body+="</table>"

    jsonbytes, err := json.Marshal(engines)

    if err == nil {
        ioutil.WriteFile("./engines.json", jsonbytes, 0777)
    }

	return "<html><head></head><body>\n" + body + "</body></html>"
}

func indexHandler(w http.ResponseWriter, r *http.Request) {	

	fmt.Fprintf(w, createIndex("","",""))

}

func findbyname(name string) (index int, found bool) {	
    var f bool = false
    var fi int
    for i, e := range engines {
    	if(e.Name==name){
    		f=true
    		fi=i
    		break
    	}
    }
    return fi,f
}

func changeHandler(w http.ResponseWriter, r *http.Request) {
    r.ParseForm()

    name:=strings.Join(r.Form["name"],"")
    path:=strings.Join(r.Form["path"],"")
    config:=strings.Join(r.Form["config"],"")

    index,found:=findbyname(name)

    if(!found){
    	var e Engine
    	engines=append(engines,e)
    	index=len(engines)-1
    }

    engines[index].Name=name
    engines[index].Path=path
    engines[index].Config=config

    fmt.Fprintf(w, createIndex("","",""))
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	name:=strings.Join(r.Form["name"],"")
    
    index,_:=findbyname(name)

  	path:=engines[index].Path
  	config:=engines[index].Config

	fmt.Fprintf(w, createIndex(name,path,config))	
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	name:=strings.Join(r.Form["name"],"")

	index,_:=findbyname(name)

	engines = append(engines[:index], engines[index+1:]...)

	fmt.Fprintf(w, createIndex("","",""))
}

func logem(myconnid uint64, em EngineMessage){
	log.Printf("# %v %v %v %.40v %v\n",myconnid,em.Action,em.Command,em.Buffer,em.Available)
}

func sendem(em EngineMessage,ws *websocket.Conn,myconnid uint64,dolog bool){
	if dolog {
		logem(myconnid, em)
	}

	jsonbytes, err := json.Marshal(em)

    if err == nil {
		if err := ws.WriteMessage(websocket.TextMessage, jsonbytes); err != nil {
			log.Println("error: could not send engine message")
		}
    } else {
    	log.Println("error: marshaling engine message failed")
    }
}

func internalError(ws *websocket.Conn, msg string, err error) {
	log.Println(msg, err)
	ws.WriteMessage(websocket.TextMessage, []byte("Internal server error."))
}

func startengine(ws *websocket.Conn,name string,myconnid uint64){
	enginews=ws
	engineconnid=myconnid

	log.Printf("myconnid : %v , starting : %s\n", engineconnid, name)

	index,found:=findbyname(name)

	if !found {
		log.Println("error: engine not found")
		return
	}

	if process != nil {
		log.Println("killing previous process")
		process.Kill()
	}

	path:=engines[index].Path

	outr, outw, err := os.Pipe()

	/*defer outr.Close()
	defer outw.Close()*/
	
	if err != nil {
		internalError(ws, "stdout:", err)
		return
	}	

	inr, inw, err := os.Pipe()

	/*defer inr.Close()
	defer inw.Close()*/

	if err != nil {
		internalError(ws, "stdin:", err)
		return
	}

	/*inr.Close()
	outw.Close()*/

	proc, err := os.StartProcess(path, flag.Args(), &os.ProcAttr{
		Files: []*os.File{inr, outw, outw},
	})

	if err != nil {		
		internalError(ws, "start:", err)
		return
	}

	process = proc

	processw = inw

	log.Println("process started")

	config := engines[index].Config

	message := append([]byte(config),'\n')

	if _ , err := processw.Write(message); err==nil {
		log.Printf("issued config : %s",config)
	}

	stdoutDone := make(chan struct{})
	go pumpStdout(outr, stdoutDone)
	go ping(stdoutDone)
}

func killengine() {
	if process == nil {
		return
	}
	process.Kill()
	process = nil
	log.Println("engine process stopped")
}

func ping(done chan struct{}) {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := enginews.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(writeWait)); err != nil {
				log.Println("ping:", err)
				killengine()
			}
		case <-done:
			return
		}
	}
}

func pumpStdout(r io.Reader, done chan struct{}) {
	defer func() {
	}()

	s := bufio.NewScanner(r)

	for s.Scan() {
		enginews.SetWriteDeadline(time.Now().Add(writeWait))

		var em EngineMessage

		em.Action="thinkingoutput"
		em.Buffer=string(s.Bytes())

		sendem(em,enginews,engineconnid,true)
	}

	if s.Err() != nil {
		log.Println("scan:", s.Err())
	}

	close(done)

	enginews.SetWriteDeadline(time.Now().Add(writeWait))
	enginews.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	time.Sleep(closeGracePeriod)
	enginews.Close()
}

func pumpStdin(ws *websocket.Conn,myconnid uint64) {
	defer ws.Close()
	ws.SetReadLimit(maxMessageSize)
	ws.SetReadDeadline(time.Now().Add(pongWait))
	ws.SetPongHandler(func(string) error { ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := ws.ReadMessage()
		if err != nil {
			break
		}

		parts:=strings.Split(string(message)," ")
		if parts[0]=="StorePresentationMessage" {
			storePresentation(ws,strings.Join(parts[1:]," "))
			break
		}

		if (parts[0]=="SendTableMessage") || (parts[0]=="StoreViewMessage") {
			log.Printf("%v ignored",parts[0])			
		} else {
			var em EngineMessage
	    	json.Unmarshal(message, &em)
			
			logem(myconnid, em)

			if em.Action=="sendavailable" {
				var am EngineMessage
				am.Action="available"
				var available []string
				for _,e := range engines {
					available=append(available,e.Name)
				}
				am.Available=available

				sendem(am,ws,myconnid,true)
			}

			if em.Action=="start" {
				startengine(ws,em.Name,myconnid)
			}

			if (em.Action=="issue") && (process != nil) {
				enginews=ws
				engineconnid=myconnid
				
				message = append([]byte(em.Command),'\n')

				if _, err := processw.Write(message); err != nil {
					log.Printf("myconnid : %v , error writing engine\n",myconnid)
					killengine()
					break
				}
			}
		}
	}
}

func serveWs(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade:", err)
		return
	}

	atomic.AddUint64(&connid, 1)

	defer ws.Close()

	pumpStdin(ws,connid)
}

func main() {
	
	fmt.Print("goeserv - the golang engine server\n")

	r := mux.NewRouter()

	r.HandleFunc("/", indexHandler).Methods("GET")

	r.HandleFunc("/change", changeHandler).Methods("POST")
	r.HandleFunc("/edit", editHandler).Methods("POST")
	r.HandleFunc("/delete", deleteHandler).Methods("POST")

	///////////////////////////////////////////////////////////////////////
	// gocserv

	r.HandleFunc("/assets/{assettype}/{assetname}", assetsHandler).Methods("GET")
	r.HandleFunc("/assets/{assettype}/{assetsubtype}/{assetname}", subassetsHandler).Methods("GET")

	r.HandleFunc("/db", serveDb).Methods("GET")
	r.HandleFunc("/newpres", newPres).Methods("GET")
	r.HandleFunc("/presentation/{presid}", servePresentation).Methods("GET")
	r.HandleFunc("/presentation/delete/{presid}", deletePresentation).Methods("GET")
	r.HandleFunc("/presentation/raw/{presid}", rawPresentation).Methods("GET")
	r.HandleFunc("/analysis/{presid}", servePresentation).Methods("GET")

	///////////////////////////////////////////////////////////////////////

	r.HandleFunc("/ws", serveWs)

	engines = getEngines()

	http.Handle("/",r)

	http.ListenAndServe(":9000", nil)

}