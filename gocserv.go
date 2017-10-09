package main

import(
	"fmt"
	"net/http"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"os"
	"path/filepath"
	"mime"
	"io"
	"strings"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"encoding/json"
	"log"
	"strconv"
)

///////////////////////////////////////////////////////////////////////////

var(	
	VERSION            = 105

	ASSETS_PATH string = os.Getenv("GOCSERV_ASSETS_PATH")	
	MONGODB_URI        = "mongodb://localhost:27017"
	FIND_MAX           = 500
	USERS_DATABASE     = "usersgolang"
	GAMES_COLLECTION   = "gamesgolang"	
)

///////////////////////////////////////////////////////////////////////////

func pathfromparts(parts []string) string {
	return strings.Join(parts,string(filepath.Separator))
}

func assetpathfromparts(parts []string) string {
	return ASSETS_PATH+strings.Join(parts,string(filepath.Separator))
}

func assetpath(assettype string,assetname string) string {
	return assetpathfromparts([]string{assettype,assetname})
}

func subassetpath(assettype string,assetsubtype string,assetname string) string {
	return assetpathfromparts([]string{assettype,assetsubtype,assetname})
}

func servAssets(w http.ResponseWriter, r *http.Request, path string) {
	ext := filepath.Ext(path)
	mimetype := mime.TypeByExtension(ext)	

	content, err := os.Open(path)

    if err == nil {
        defer content.Close()
	    w.Header().Set("Content-Type", mimetype)
	    io.Copy(w, content)
    } else {
    	fmt.Fprintf(w, "Not Found 404 [ info : path %v ext %v mine %v ]",path,ext,mimetype)
    }
}

///////////////////////////////////////////////////////////////////////////

func dbError(w http.ResponseWriter,comment string) {
	fmt.Fprintf(w, "db error "+comment)
}

func dbStoreError() {
	log.Println("db store error")
}

func mainMenu() string {
	body:=""

	body+="<a href=/>Home</a> | \n"
	body+="<a href=/db>Presentations</a> | \n"
	body+="<a href=/newpres>New presentation</a>\n"

	body+="<hr>"

	return body
}

func serveDb(w http.ResponseWriter, r *http.Request) {
	session, err := mgo.Dial(MONGODB_URI)
    if err == nil {
    	defer session.Close()        

        c := session.DB(USERS_DATABASE).C(GAMES_COLLECTION)

        result := []GameWithPresentation{}

        c.Find(nil).Limit(FIND_MAX).Iter().All(&result)

        html:="<html><head></head><body><table cellpadding=3 cellspacing=3>"

        html+=mainMenu()

        for _ , g:=range result {
        	pres := g.Presentation
        	presid := pres.Id

        	html+="<tr>"
        	html+=fmt.Sprintf("<td><a href=/presentation/%v>%v</a></td>", presid, g.Presentationtitle)
        	html+=fmt.Sprintf("<td><a href=/presentation/raw/%v>Raw</a></td>",presid)
        	if pres.Canedit=="no" {
        		html+="<td>Hybernated</td>"
        	} else if pres.Candelete=="no" {
        		html+="<td>Archived</td>"
    		} else {
    			html+=fmt.Sprintf("<td><a href=/presentation/delete/%v>Delete</a></td>",presid)
    		}    		
			html+="</tr>"
		}

		html+="</table></body></html>"

		fmt.Fprint(w,html)
    } else {
    	dbError(w,"serving presentation list")
    }        
}

func newPres(w http.ResponseWriter, r *http.Request) {
	gwp := GameWithPresentation{}

	gwp.Presentation.Title   = "Chessapp Presentation"
	gwp.Presentation.Version = 0
	gwp.Presentation.Book    = Book{Positions:map[string]BookPosition{}}

	fmt.Fprint(w,presHtml(gwp,-1))
}

func presHtml(gwp GameWithPresentation,currentnodeid int) string {
	presgbytes , err := json.Marshal(gwp)

	if err != nil {
		log.Printf("json marshal error in %v\n",gwp)
	}

	chessconfig:="{\"kind\":\"analysis\",\"currentnodeid\":"+strconv.Itoa(currentnodeid)+",\"translations\":"+translationsjson()+",\"presgame\":"+string(presgbytes)+"}"

    user:=""
    title:=""
    piecetype:=""
    scriptversionstr:=strconv.Itoa(VERSION)

    html:="<!DOCTYPE html>\n"
    html+="<html lang='en'>\n"

    html+="<head>\n"
    html+="<meta charset='utf-8'>\n"
	html+="<meta http-equiv='X-UA-Compatible' content='IE=edge'>\n"
	html+="<meta name='viewport' content='width=device-width, initial-scale=1'>\n"
	html+="<link rel='shortcut icon' type='image/png' href=/assets/images/favicon.png>\n"	
	html+="<title>"+title+"</title>\n"
	html+="<link rel='stylesheet' href=/assets/stylesheets/reset.css>\n"
	html+="<link rel='stylesheet' href=/assets/stylesheets/main.css>\n"
	html+="<link rel='stylesheet' href=/assets/stylesheets/piece/alpha.css>\n"
	html+="<link rel='stylesheet' href=/assets/stylesheets/piece/merida.css>\n"
    html+="</head>\n"

    html+="<body>\n"
    html+="<div id='chessconfig' style='visibility: hidden; width: 0px; height: 0px;''>"+chessconfig+"</div>\n"        
    html+="<div id='root' admin='false' user='"+user+"'' viewid='chess' viewserialized='' piecetype='"+piecetype+"'></div>\n"
    html+="<div id='info' style='position: relative;'>Loading board...</div>\n"
    html+="<script src=/assets/javascripts/client-jsdeps.min.js?v"+scriptversionstr+"></script>\n"
	html+="<script src=/assets/javascripts/client-opt.js?v"+scriptversionstr+"></script>\n"
    html+="</body>\n"

    html+="</html>\n"

    return html
}

func servePresentation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	presid := vars["presid"]

	currentnodeid := -1

	parts := strings.Split(presid,"--")

	if len(parts)>1 {		
		currentnodeid , _ = strconv.Atoi(parts[1])
		presid = parts[0]
	}

	session , err := mgo.Dial(MONGODB_URI)
    if err == nil {
    	defer session.Close()        

        c := session.DB(USERS_DATABASE).C(GAMES_COLLECTION)

        gwp := GameWithPresentation{}

        err := c.Find(bson.M{"presentationid":presid}).One(&gwp)

        if err != nil {
        	dbError(w,fmt.Sprintf("finding presentation presid %v currentnodeid %v",presid,currentnodeid))
        	return
        }        

        gwp.Presentation.checksanity()        

        html:=presHtml(gwp,currentnodeid)

		fmt.Fprintf(w, "%v", html)
    } else {
    	dbError(w,"in serving presentation session")
    }        
}

func deletePresentation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	presid := vars["presid"]

	session, err := mgo.Dial(MONGODB_URI)

    if err == nil {
    	defer session.Close()        

        c := session.DB(USERS_DATABASE).C(GAMES_COLLECTION)

        c.Remove(bson.M{"_id":presid})
        
        serveDb(w,r)
    } else {
    	dbError(w," deleting presentation")
    }        
}

func rawPresentation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	presid := vars["presid"]

	session, err := mgo.Dial(MONGODB_URI)

    if err == nil {
    	defer session.Close()        

        c := session.DB(USERS_DATABASE).C(GAMES_COLLECTION)

		gwp := GameWithPresentation{}

		err := c.Find(bson.M{"presentationid":presid}).One(&gwp)

		if err != nil {
        	dbError(w,fmt.Sprintf("finding presentation presid %v",presid))
        	return
        }        

        gwp.Presentation.checksanity()        

        presgbytes , err := json.Marshal(gwp)

		if err != nil {
			log.Printf("json marshal error in %v\n",gwp)
		}

        html:=string(presgbytes)

        w.Header().Set("Content-Type", "text/plain")

		fmt.Fprintf(w, "%v", html)
    } else {
    	dbError(w," serving raw presentation")
    }        
}

func assetsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	servAssets(w,r,assetpath(vars["assettype"],vars["assetname"]))
}

func subassetsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	servAssets(w,r,subassetpath(vars["assettype"],vars["assetsubtype"],vars["assetname"]))
}

func storePresentation(ws *websocket.Conn,message string){
	var spm StorePresentationMessage

    err := json.Unmarshal([]byte(message), &spm)

    if err != nil {
		fmt.Printf("store pres unmarshal error in : %v", message)    	
		return
    }

    presid := spm.Presid
    presg := spm.Presg

    pres := presg.Presentation

    prestitle := pres.Title
    
	fmt.Printf("store pres : %v\nmessage : %v\n", presid, message)

	session, err := mgo.Dial(MONGODB_URI)

    if err == nil {
    	defer session.Close()        

        c := session.DB(USERS_DATABASE).C(GAMES_COLLECTION)

        doc:=bson.M{
        	"_id":presid,
        	"presentationid":presid,
        	"presentationtitle":prestitle,
        	"presentation":pres,
        }

        c.Upsert(bson.M{"_id":presid},doc)

        ws.WriteMessage(websocket.TextMessage, []byte("StorePresentationResultMessage {\"success\":true}"))
    } else {
    	dbStoreError()
    }        
}