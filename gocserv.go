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
)

var(	
	ASSETS_PATH string = os.Getenv("GOCSERV_ASSETS_PATH")	
	MONGODB_URI        = "mongodb://localhost:27017"
	FIND_MAX           = 500
	GAMES_COLLECTION   = "gamesgolang"

	TRANSLATIONS map[string]string = map[string]string{		
		"presentation.engine":"Engine",
		"presentation.hybernate":"Hybernate",
		"presentation.nothybernated":"Not hybernated",
		"presentation.hybernated":"Hybernated",
		"presentation.changehybernated":"Change",
		"presentation.archive":"Archive",
		"presentation.notarchived":"Not archived",
		"presentation.archived":"Archived",
		"presentation.changearchived":"Change",
		"presentation.access.denied":"Access denied!",
		"presentation.deleted":"Presentation deleted.",
		"presentation.doesnotexist":"Presentation does not exist!",
		"ccfg.change":"Change",
		"ccfg.essay":"Essay",
		"ccfg.saveessay":"Save essay",
		"ccfg.savemovenote":"Save",
		"ccfg.editmovenote":"Edit",
		"ccfg.nodelink":"Node link",
		"ccfg.nodeurl":"Node url",
		"ccfg.version":"Version",
		"ccfg.white":"White",
		"ccfg.black":"Black",
		"ccfg.yellow":"Yellow",
		"ccfg.red":"Red",
		"ccfg.variant":"Variant",
		"ccfg.timecontrol":"Time control",
		"ccfg.gamesanalysis":"Game analysis",
		"ccfg.whiteresigned":"White resigned",
		"ccfg.blackresigned":"Black resigned",
		"ccfg.whiteflagged":"White flagged",
		"ccfg.blackflagged":"Black flagged",
		"ccfg.whitemated":"White mated",
		"ccfg.blackmated":"Black mated",
		"ccfg.open":"Open",
		"ccfg.inprogress":"In progress",
		"ccfg.terminated":"Terminated",
		"ccfg.result":"Result",
		"ccfg.created":"Created",
		"ccfg.status":"Status",
		"ccfg.refresh":"Refresh",
		"ccfg.load":"Load",
		"ccfg.loadingboard":"Loading board ...",
		"ccfg.importpgn":"Import PGN : ",
		"ccfg.savenotes":"Save notes",
		"ccfg.upload":"Upload",
		"ccfg.title":"Title",
		"ccfg.owner":"Owner",
		"ccfg.id":"Id",
		"ccfg.gennew":"Generate new",
		"ccfg.resign":"Resign",
		"ccfg.stand":"Stand",
		"ccfg.play":"Play",
		"ccfg.playai":"Play AI",
		"ccfg.flip":"Flip",
		"ccfg.yes":"Yes",
		"ccfg.no":"No",
		"ccfg.current":"Current",
		"ccfg.book":"Book",
		"ccfg.notes":"Notes",
		"ccfg.presentation":"Presentation",
		"ccfg.connecting":"Connecting ...",
		"ccfg.createtable":"Create table",
		"ccfg.creatingtable":"Creating table",
		"ccfg.create":"Create",
	}
)

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
	ext:=filepath.Ext(path)
	mimetype:=mime.TypeByExtension(ext)	
	content, err := os.Open(path)
    if err == nil {
        defer content.Close()
	    w.Header().Set("Content-Type", mimetype)
	    io.Copy(w, content)
    } else {
    	fmt.Fprintf(w, "Not Found - path %v ext %v mine %v",path,ext,mimetype)
    }
}

type User struct {
	Handle string   `json:"handle"`
	Rating float64  `json:"rating"`
	Rd float64      `json:"rd"`
	Email string    `json:"email"`
}

type Presentation struct {
	Id string                     `json:"id"`
	Title string                  `json:"title"`
	Owner string                  `json:"owner"`
	Pgn string                    `json:"pgn"`
	Currentlinealgeb string       `json:"currentlinealgeb"`
	Version int                   `json:"version"`
	Book Book                     `json:"book"`
	Flip bool                     `json:"flip"`
}

func (pr *Presentation) checksanity() {
	pr.Book.checksanity()
}

type Game struct {
	Presentationid string     `json:"presentationid"`
	Presentationtitle string  `json:"presentationtitle"`
}

type GameWithPresentation struct {
	Presentationid string       `json:"presentationid"`
	Presentationtitle string    `json:"presentationtitle"`
	Presentation Presentation   `json:"presentation"`
}

type StorePresentationMessage struct {
	Presid string                `json:"presid"`
	Presg GameWithPresentation   `json:"presg"`
}

type BookMove struct {
	Fen string                  `json:"fen"`
	San string                  `json:"san"`
	Annot string                `json:"annot"`
	Comment string              `json:"comment"`
	Open bool                   `json:"open"`
	Hasscore bool               `json:"hasscore"`
	Scorecp bool                `json:"scorecp"`
	Scoremate bool              `json:"scoremate"`
	Score int                   `json:"score"`
	Depth int                   `json:"depth"`
}

func (bm *BookMove) checksanity() {
	if (!bm.Scorecp) && (!bm.Scoremate){
		bm.Scorecp=true
	}
}

type BookPosition struct {
	Fen string                  `json:"fen"`
	Moves map[string]BookMove   `json:"moves"`
	Notes string                `json:"notes"`
	Arrowalgebs []string        `json:"arrowalgebs"`
}

func (bp *BookPosition) checksanity() {
	newmoves:=map[string]BookMove{}
	for san , bm := range bp.Moves {
		bm.checksanity()
		newmoves[san]=bm
	}
	bp.Moves=newmoves
}

type Book struct {
	Positions map[string]BookPosition    `json:"positions"`
	Essay string                         `json:"essay"`
}

func (bk *Book) checksanity() {
	newpositions:=map[string]BookPosition{}
	for fen , bp := range bk.Positions {
		bp.checksanity()
		newpositions[fen]=bp
	}
	bk.Positions=newpositions
}

func dbError(w http.ResponseWriter) {
	fmt.Fprintf(w, "db error")
}

func dbStoreError() {
	log.Println("db store error")
}

func serveDb(w http.ResponseWriter, r *http.Request) {
	session, err := mgo.Dial(MONGODB_URI)
    if err == nil {
    	defer session.Close()        
        c := session.DB("users").C(GAMES_COLLECTION)
        result := []Game{}
        c.Find(nil).Limit(FIND_MAX).Iter().All(&result)
        for _ , g:=range result {
			fmt.Fprintf(w, "<a href=/presentation/%v>%v</a><br>", g.Presentationid, g.Presentationtitle)
		}
    } else {
    	dbError(w)
    }        
}

func newPres(w http.ResponseWriter, r *http.Request) {
	gwp:=GameWithPresentation{}
	gwp.Presentation.Title="Chessapp Presentation"
	gwp.Presentation.Version=0
	gwp.Presentation.Book=Book{Positions:map[string]BookPosition{}}
	fmt.Fprint(w,presHtml(gwp))
}

func translationsjson() string {
	first:=true
	body:="{"
	for k , v := range TRANSLATIONS {
		if first {
			first=false
		} else {
			body+=","
		}
		body+="\""+k+"\":\""+v+"\""
	}
	body+="}"
	return body
}

func presHtml(gwp GameWithPresentation) string {
	presgbytes, err := json.Marshal(gwp)
	if err!=nil {
		log.Printf("json marshal error in %v\n",gwp)
	}
	chessconfig:="{\"kind\":\"analysis\",\"translations\":"+translationsjson()+",\"presgame\":"+string(presgbytes)+"}"
    user:=""
    title:=""
    piecetype:=""
    scriptversionstr:="103"

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
	presid:=vars["presid"]

	session, err := mgo.Dial(MONGODB_URI)
    if err == nil {
    	defer session.Close()        
        c := session.DB("users").C(GAMES_COLLECTION)
        gwp := GameWithPresentation{}
        err := c.Find(bson.M{"presentationid":presid}).One(&gwp)

        if err!=nil {
        	dbError(w)
        	return
        }        

        gwp.Presentation.checksanity()        

        html:=presHtml(gwp)

		fmt.Fprintf(w, "%v", html)
    } else {
    	dbError(w)
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
		fmt.Printf("store pres unmarshal error in : %v",message)    	
		return
    }

    presid:=spm.Presid
    presg:=spm.Presg

    pres:=presg.Presentation

    prestitle:=pres.Title
    
	fmt.Printf("store pres : %v\nmessage : %v\n",presid,message)

	session, err := mgo.Dial(MONGODB_URI)
    if err == nil {
    	defer session.Close()        
        c := session.DB("users").C(GAMES_COLLECTION)
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