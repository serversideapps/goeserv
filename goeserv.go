package main

import (
	"fmt"
	"net/http"
	"encoding/json"    
    "io/ioutil"
    "os"
    "strings"
)

type Engine struct {
    Name    string `json:"name"`
    Path    string `json:"path"`
    Config  string `json:"config"`
}

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

var (
	engines []Engine
)

func createIndex(iname string,ipath string,iconfig string) string {
	var body string

	body+="<form method=post action='/change'><table>"
	body+="<tr><td>Name</td><td><input id=name name=name type=text value='"+iname+"'></td>"
	body+="<tr><td>Path</td><td><input id=path name=path type=text value='"+ipath+"'></td>"
	body+="<tr><td>Config</td><td><textarea id=config name=config cols=80 rows=5>"+iconfig+"</textarea></td>"
	body+="<tr><td></td><td><input type=submit value=Submit></td>"
	body+="</table></form>"

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


func main() {
	
	fmt.Print("goeserv - the golang engine server\n")

	http.HandleFunc("/", indexHandler)

	http.HandleFunc("/change", changeHandler)
	http.HandleFunc("/edit", editHandler)
	http.HandleFunc("/delete", deleteHandler)

	engines = getEngines()

	http.ListenAndServe(":9000", nil)

}