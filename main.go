package main

import (
	"encoding/json"
	"log"
	"github.com/gorilla/mux"
	"net/http"
	"github.com/couchbase/gocb"
	"time"
	//"os"

	hashids "github.com/speps/go-hashids"

)

var bucket *gocb.Bucket
var bucketName string

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func ExpandEndpoint(w http.ResponseWriter, req *http.Request) {
	var n1qlParams []interface{}
	query := gocb.NewN1qlQuery("SELECT `" + bucketName + "`.* FROM `" + bucketName + "` WHERE ShortUrl = $1")
	params := req.URL.Query()
	n1qlParams = append(n1qlParams, params.Get("shortUrl"))
	rows, queryErr := bucket.ExecuteN1qlQuery(query, n1qlParams)
	checkError(queryErr)
	var row MyUrl
	rows.One(&row)
	json.NewEncoder(w).Encode(row)
}

func CreateEndpoint(w http.ResponseWriter, req *http.Request) {
	var url MyUrl
	jsonError := json.NewDecoder(req.Body).Decode(&url)
	log.Println(url)
	checkError(jsonError)
	var n1qlParams []interface{} // query parameters in n1ql
	n1qlParams = append(n1qlParams, url.Longurl)
	query := gocb.NewN1qlQuery("SELECT `" + bucketName + "`.* FROM `" + bucketName + "` WHERE Longurl = $1")


	rows, err := bucket.ExecuteN1qlQuery(query, n1qlParams)

	if err != nil {
		w.WriteHeader(401)
		w.Write([]byte(err.Error()))
		return
	}
	var row MyUrl
	rows.One(&row)
	if row == (MyUrl{}) {
		hd := hashids.NewData()
		h, err := hashids.NewWithData(hd)
		checkError(err)
		now := time.Now()
		url.ID, _ = h.Encode([]int{int(now.Unix())})
		url.ShortUrl = "http://localhost:12345/" + url.ID
		bucket.Insert(url.ID, url, 0)
	} else {
		url = row
	}
	json.NewEncoder(w).Encode(url)
}

func RootEndpoint(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	var url MyUrl
	bucket.Get(params["id"], &url)
	http.Redirect(w, req,url.Longurl, 301)
}

type MyUrl struct {
	ID string `json:"id,omitempty"`
	Longurl string `json:"Longurl, omitempty"`
	ShortUrl string `json:"ShortUrl, omitempty"`
}



func main() {
	router := mux.NewRouter()
	auth := gocb.PasswordAuthenticator{"khaledthekhaled", "khaledtinubu"}
	cluster, err1 := gocb.Connect("couchbase://127.0.0.1")
	checkError(err1)
	err2 := cluster.Authenticate(auth)
	checkError(err2)
	bucketName = "example"
	myBucket, bucketError := cluster.OpenBucket(bucketName,"")
	checkError(bucketError)
	bucket = myBucket
	router.HandleFunc("/create", CreateEndpoint).Methods("PUT")
	router.HandleFunc("/expand", ExpandEndpoint).Methods("GET")
	router.HandleFunc("/{id}", RootEndpoint).Methods("GET")
	log.Fatal(http.ListenAndServe(":12345", router))
}