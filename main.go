package main

import (
	"encoding/json"
	"log"
	"github.com/gorilla/mux"
	"net/http"
	"github.com/couchbase/gocb"
	"fmt"
	"time"

	hashids "github.com/speps/go-hashids"

)

var bucket *gocb.Bucket
var bucketName string

func ExpandEndpoint(w http.ResponseWriter, req *http.Request) {

}

func CreateEndpoint(w http.ResponseWriter, req *http.Request) {
	var url MyURL
	_ = json.NewDecoder(req.Body).Decode(&url)
	var n1qlParams []interface{}
	n1qlParams = append(n1qlParams, url.Longurl)
	query := gocb.NewN1qlQuery("SELECT `" + bucketName + "`.* FROM `" + bucketName + "` WHERE Longurl = $1")
	fmt.Println(bucket.ExecuteN1qlQuery)
	fmt.Println(bucketName)

	rows, err := bucket.ExecuteN1qlQuery(query, n1qlParams)
	if err != nil {
		w.WriteHeader(401)
		w.Write([]byte(err.Error()))
		return
	}
	var row MyURL
	rows.One(&row)
	if row == (MyURL{}) {
		hd := hashids.NewData()
		h, err := hashids.NewWithData(hd)
		if err != nil {
			log.Fatal(err)
		}
		now := time.Now()
		url.ID, _ = h.Encode([]int{int(now.Unix())})
		url.ShortUrl = "http//localhost:12345/" + url.ID
		bucket.Insert(url.ID, url, 0)
	} else {
		url = row
	}
	json.NewEncoder(w).Encode(url)

}

func RootEndpoint(w http.ResponseWriter, req *http.Request) {

}

type MyURL struct {
	ID string `json:"id,omitempty"`
	Longurl string `json:"Longurl, omitempty"`
	ShortUrl string `json:"ShortUrl, omitempty"`
}



func main() {
	router := mux.NewRouter()
	auth :=  gocb.PasswordAuthenticator{"khaledthekhaled", "khaledtinubu"}
	cluster, err1 := gocb.Connect("couchbase://127.0.0.1")
	if err1 != nil {
			log.Fatal(err1)
		}
	err2 := cluster.Authenticate(auth)
	if err2 != nil {
			fmt.Println("FAD")
			log.Fatal(err2)
		}
	bucketName = "example"
	myBucket, err := cluster.OpenBucket(bucketName,"")
	if err != nil {
			log.Fatal(err)
		}
	bucket = myBucket
	fmt.Println(bucket)
	router.HandleFunc("/create", CreateEndpoint).Methods("PUT")
	router.HandleFunc("/expand", ExpandEndpoint).Methods("GET")
	router.HandleFunc("/{id}", RootEndpoint).Methods("GET")
	log.Fatal(http.ListenAndServe(":12345", router))
}