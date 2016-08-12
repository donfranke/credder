package main

import (
    "io"
    "net/http"
    "log"
    "gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
    "fmt"
)

func HelloServer(w http.ResponseWriter, r *http.Request) {
	entryID := r.URL.Query().Get("id")
	fmt.Println("Remote addr: ",r.RemoteAddr)
	fmt.Println("User Agent: ",r.Header.Get("User-Agent"))

	var resultString string
	if (entryID!="" ) {
	    result := getCreds(entryID)
	    resultString = "{\"entryid\":\"" + entryID + "\",\"secretinfo\":\"" + result + "\"}"
	} else {
		resultString = "{\"entryid\":\"" + entryID + "\",\"secretinfo\":\"not found\"}"

	}
	io.WriteString(w, resultString)
}

func main() {
    http.HandleFunc("/find", HelloServer)
    err := http.ListenAndServeTLS(":443", "server.crt", "server.key", nil)
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}


type Cred struct {
        EntryID string
        SecretInfo string
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
        fmt.Println("Request for ",id)
        return result.SecretInfo
        //fmt.Println("SecretInfo:", result.SecretInfo)
}