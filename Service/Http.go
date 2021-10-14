package service

import (
	"encoding/json"
	"fmt"
	Config "godnslog/Config"
	Db "godnslog/Database"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	mux "github.com/gorilla/mux"
)

type ResponseProps struct {
	StatusCode int
	Message    string
	Total      int
	Records    []interface{}
}
type spaHandler struct {
	staticPath string
	indexPath  string
}

func jsonRes(resp ResponseProps) string {
	if res, err := json.Marshal(resp); err == nil {
		return string(res)
	} else {
		res, _ := json.Marshal(&ResponseProps{
			StatusCode: 500,
			Message:    "Error",
		})
		return string(res)
	}
}
func recordGet(writer http.ResponseWriter, req *http.Request) {
	token := req.Header.Get("authtoken")
	if token != Config.Config.Token {
		fmt.Fprint(writer, jsonRes(ResponseProps{
			StatusCode: 401,
			Message:    "Token verify failed",
		}))
	} else {
		args := req.URL.Query()
		table := args.Get("type")
		offset := args.Get("page")
		search := args.Get("search")
		page, err := strconv.Atoi(offset)
		where := ""
		if err != nil || page < 0 {
			page = 1
		}
		if table != "dns" && table != "http" {
			table = "dns"
		}
		if len(search) > 0 {
			switch table {
			case "dns":
				where = "where record like '%" + search + "%'"
			case "http":
				where = "where url like '%" + search + "%' or body like '%" + search + "%'"
			default:
				fmt.Fprint(writer, jsonRes(ResponseProps{
					StatusCode: 500,
					Message:    "Search type error",
				}))
			}
		}
		result, total := Db.Get(table, page, where)
		fmt.Fprint(writer, jsonRes(ResponseProps{
			StatusCode: 200,
			Total:      total,
			Records:    result,
		}))
	}
}
func recordClean(writer http.ResponseWriter, req *http.Request) {
	args := req.URL.Query()
	table := args.Get("type")
	token := req.Header.Get("authtoken")
	if token == Config.Config.Token {
		if table == "dns" || table == "http" {
			Db.Clean(table)
		}
		fmt.Fprint(writer, jsonRes(ResponseProps{
			StatusCode: 200,
			Message:    table + " records clean",
		}))
	} else {
		fmt.Fprint(writer, jsonRes(ResponseProps{
			StatusCode: 401,
			Message:    "Token verify failed",
		}))
	}
}
func tokenVerify(writer http.ResponseWriter, req *http.Request) {
	data := make(map[string]string)
	body, _ := ioutil.ReadAll(req.Body)
	json.Unmarshal(body, &data)
	if data["token"] == Config.Config.Token {
		fmt.Fprint(writer, jsonRes(ResponseProps{
			StatusCode: 200,
			Message:    "",
		}))
	} else {
		fmt.Fprint(writer, jsonRes(ResponseProps{
			StatusCode: 401,
			Message:    "Token verify failed",
		}))
	}
}
func httpRecord(resp http.ResponseWriter, req *http.Request) {
	body, _ := ioutil.ReadAll(req.Body)
	headers := []string{""}
	for k, v := range req.Header {
		for _, item := range v {
			headers = append(headers, k+":"+item)
		}
	}
	record := Db.HttpRecord{
		Method:   req.Method,
		Url:      req.RequestURI,
		Head:     strings.TrimPrefix(strings.Join(headers, "\n"), "\n"),
		Body:     string(body),
		Ip:       req.RemoteAddr,
		Datetime: time.Now().Format("2006-01-02"),
	}
	_ = record.Insert()
	log.Printf("HTTP Request %s %s %s", req.Method, req.RemoteAddr, req.RequestURI)
	io.WriteString(resp, "<html><body><h1>It works!</h1></body></html>")
}
func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path, err := filepath.Abs(r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	path = filepath.Join(h.staticPath, strings.TrimPrefix(path, "/admin"))
	_, err = os.Stat(path)
	if !os.IsNotExist(err) || path == "Frontend" {
		renderFile := path
		if path == "Frontend" {
			renderFile = h.indexPath
		} else {
			renderFile = strings.TrimPrefix(renderFile, "Frontend/")
		}
		http.ServeFile(w, r, filepath.Join(h.staticPath, renderFile))
		return
	} else if err != nil {
		log.Println("err " + err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
func HttpServe() {
	handler := mux.NewRouter()
	if Config.Config.HttpView {
		spa := spaHandler{
			staticPath: "Frontend",
			indexPath:  "index.html",
		}
		handler.PathPrefix("/admin").Handler(spa)
	}
	handler.HandleFunc("/api/get", recordGet)
	handler.HandleFunc("/api/clean", recordClean)
	handler.HandleFunc("/api/auth", tokenVerify)
	handler.HandleFunc(`/{value:.*}`, httpRecord)
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", Config.Config.Listen, Config.Config.HttpPort),
		Handler: handler,
	}
	log.Printf("HTTP service running at: http://%s:%d", Config.Config.Domain, Config.Config.HttpPort)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal("HTTP service start error !")
	}
}
