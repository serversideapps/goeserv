package main

import(
	"fmt"
	"net/http"
	"github.com/gorilla/mux"
	"os"
	"path/filepath"
	"mime"
	"io"
	"strings"
)

var(
	ASSETS_PATH string = os.Getenv("GOCSERV_ASSETS_PATH")
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

func assetsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	servAssets(w,r,assetpath(vars["assettype"],vars["assetname"]))
}

func subassetsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	servAssets(w,r,subassetpath(vars["assettype"],vars["assetsubtype"],vars["assetname"]))
}

