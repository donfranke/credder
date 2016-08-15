package main

import (
	"io"
	"net/http"
	"log"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"fmt"
	"strings"
	"time"
	"regexp"
)

type Cred struct {
	EntryID string
	SecretInfo string
}
type Algo struct {
	ID string
	Seq string
}

type LogEvent struct {
	Timestamp time.Time
	EventDescription string
}

func FindServer(w http.ResponseWriter, r *http.Request) {

	entryID := r.URL.Query().Get("id")	
	encodedRequest := r.URL.Query().Get("req")
	fmt.Printf("Encoded request: %s\n",encodedRequest)
	ra := r.RemoteAddr
	remoteIP := ""
	ipv4regex := "(\\d{1,3}\\.){3}\\d{1,3}\\:\\d{1,5}"
	rx, _ := regexp.Compile(ipv4regex)
	//remoteIP := ""
	if(rx.MatchString(ra)) {
		i := strings.Index(ra, ":")
		remoteIP = ra[0:i]
	} else {
		remoteIP = "0.0.0.0"
	}
	fmt.Printf("  Remote IP: %s\n",remoteIP)
	
	fmt.Println("  User Agent: ",r.Header.Get("User-Agent"))
	eventdesc := "Request received for "+entryID+" ("+r.RemoteAddr+" "+r.Header.Get("User-Agent")+")"
	logEvent(eventdesc)

	var resultString string
	if (entryID!="" ) {
		algo := getAlgo()
		// decode request
		//decodedrequest = decode(algo,encodedrequest)
		fmt.Printf("Encoding sequence to use: %s\n",algo)
		result := getCreds(entryID)
		resultString = "{\"entryid\":\"" + entryID + "\",\"secretinfo\":\"" + result + "\"}"
	} else {
		resultString = "{\"entryid\":\"" + entryID + "\",\"secretinfo\":\"not found\"}"
	}
	io.WriteString(w, resultString)
}

func main() {
	webserverport := "8888"
	http.HandleFunc("/find", FindServer)
	//err := http.ListenAndServeTLS(":443", "server.crt", "server.key", nil)
	err := http.ListenAndServe(":" + webserverport, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
	fmt.Println("Web server running on port %d\n",webserverport)
}



func getAlgo() string {
	session, err := mgo.Dial("127.0.0.1")
	if err != nil {
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	c := session.DB("credder").C("algo")

	result := Algo{}
	err = c.Find(bson.M{"id": 1}).One(&result)
	if err != nil {
		fmt.Println(err)
	}
	return result.Seq
	
}

func getCreds(id string) string {
	session, err := mgo.Dial("127.0.0.1")
	if err != nil {
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	c := session.DB("credder").C("creds")

	result := Cred{}
	err = c.Find(bson.M{"entryid": id}).One(&result)
	if err != nil {
	//        log.Fatal(err)
		fmt.Println(err)
	}
	return result.SecretInfo
	//fmt.Println("SecretInfo:", result.SecretInfo)
}

func logEvent(event string) {
	session, err := mgo.Dial("127.0.0.1")
	if err != nil {
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	c := session.DB("credder").C("eventlog")
	err = c.Insert(&LogEvent{time.Now(),event})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Event has been logged")
}


