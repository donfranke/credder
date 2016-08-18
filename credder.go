package main

import (
	"fmt"
	"github.com/donfranke/algo"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var encKey string

const C_MONGODB = "127.0.0.1"
const C_ERROR_MESSAGE = "NOTFOUND"

type Cred struct {
	ID         string
	SecretInfo string
	KeyID      string
}
type Algo struct {
	ID  string
	Key string
}

type LogEvent struct {
	Timestamp        time.Time
	EventDescription string
}

// use this web service call to get a specific encryption key
func KeyServer(w http.ResponseWriter, r *http.Request) {
	var resultString string

	// expecting POST
	r.ParseForm()
	// fetch values from request
	keyID := r.FormValue("keyid")
	appName := r.FormValue("appname")
	remoteIP := ExtractIP(r.RemoteAddr)
	userAgent := r.Header.Get("User-Agent")

	// log event
	eventdesc := "Request received for key " + keyID + " (" + remoteIP + " " + userAgent + ") from " + appName
	logEvent(eventdesc)

	// 1. make sure request is legit
	result := validateKeyRequest(keyID, appName, remoteIP, userAgent)
	fmt.Printf("Key request validation result: %d\n", result)

	// 2. get key (if validation passed)
	if result>0  {
		encKey = getKey(keyID)
	} else {
		encKey = C_ERROR_MESSAGE
	}
	resultString = "{\"key\":\""+encKey+"\"}"
	
	io.WriteString(w, resultString)
	fmt.Printf("Sent to client: %s\n", resultString)

}

// use this web service call to retrieve a specific credential document
func CredServer(w http.ResponseWriter, r *http.Request) {
	var result Cred
	// fetch values from request
	credID := r.FormValue("credid")
	appName := r.FormValue("appname")
	remoteIP := ExtractIP(r.RemoteAddr)
	userAgent := r.Header.Get("User-Agent")

	// log event
	eventdesc := "Request received for cred " + credID + " (" + remoteIP + " " + userAgent + ") from " + appName
	logEvent(eventdesc)

	var resultString string

	// 1. validate request 
	validateresult := validateCredRequest(credID, appName, remoteIP, userAgent)
	fmt.Printf("Cred request validation result: %d\n", validateresult)

	if validateresult>0 {
		// display info
		fmt.Printf("=== CRED and KEY ID ===\n")
		fmt.Printf("\tCred ID: %s\n", credID)
		fmt.Printf("\tApp name: %s\n", appName)
		fmt.Printf("\tRemote IP: %s\n", remoteIP)
		fmt.Printf("\tUser Agent: %s\n", userAgent)

		result = getCreds(credID)
		resultString = "{\"secretinfo\":\"" + result.SecretInfo + "\",\"keyid\":\"" + result.KeyID + "\"}"
	} else {
		resultString = "{\"secretinfo\":\"" + C_ERROR_MESSAGE + "\"}"
	}
	fmt.Printf("Sent to client: %s\n", resultString)
	io.WriteString(w, resultString)

	// decrypt  <-- DEBUG
	//plaintext := decryptValue(encKey,result)
	//fmt.Printf("Plaintext: %s\n",plaintext)
}

func main() {
	webserverport := "8888"

	// init handlers
	http.HandleFunc("/cred", CredServer)
	http.HandleFunc("/key", KeyServer)
	fmt.Printf("Web service started on port %s\n", webserverport)

	err := http.ListenAndServe(":"+webserverport, nil)
	//err := http.ListenAndServeTLS(":443", "server.crt", "server.key", nil)

	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
	fmt.Println("Web server running on port %d\n", webserverport)
}

func validateKeyRequest(keyID string, appName string, remoteIP string, userAgent string) int {
	fmt.Printf("=== VALIDATE KEY REQUEST ===\n")
	fmt.Printf("\tKey ID: %s\n", keyID)
	fmt.Printf("\tApp name: %s\n", appName)
	fmt.Printf("\tRemote IP: %s\n", remoteIP)
	fmt.Printf("\tUser Agent: %s\n", userAgent)
	
	session, err := mgo.Dial(C_MONGODB)
	if err != nil {
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	c := session.DB("credder").C("keyprivs")

	rscount, err := c.Find(bson.M{"keyid": keyID, "appname": appName, "remoteip": remoteIP, "useragent": userAgent}).Count()
	if err != nil {
		fmt.Printf("Key Request Validation ERROR: %s\n", err)
	}
	return rscount
}

func validateCredRequest(credID string, appName string, remoteIP string, userAgent string) int {
	fmt.Printf("=== VALIDATE CRED REQUEST ===\n")
	fmt.Printf("\tCred ID: %s\n", credID)
	fmt.Printf("\tApp name: %s\n", appName)
	fmt.Printf("\tRemote IP: %s\n", remoteIP)
	fmt.Printf("\tUser Agent: %s\n", userAgent)
	
	session, err := mgo.Dial(C_MONGODB)
	if err != nil {
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	c := session.DB("credder").C("credprivs")

	rscount, err := c.Find(bson.M{"credid": credID, "appname": appName, "remoteip": remoteIP, "useragent": userAgent}).Count()
	if err != nil {
		fmt.Printf("Cred Request Validation ERROR: %s\n", err)
	}
	return rscount
}

// retrieve encryption key from database
func getKey(id string) string {
	session, err := mgo.Dial(C_MONGODB)
	if err != nil {
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	c := session.DB("credder").C("keys")

	result := Algo{}
	err = c.Find(bson.M{"id": id}).One(&result)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
	}
	return result.Key
}

//
func getCreds(id string) Cred {
	session, err := mgo.Dial(C_MONGODB)
	if err != nil {
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	c := session.DB("credder").C("creds")

	result := Cred{}
	err = c.Find(bson.M{"id": id}).One(&result)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
	}
	return result
}

func logEvent(event string) {
	session, err := mgo.Dial(C_MONGODB)
	if err != nil {
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	c := session.DB("credder").C("eventlog")
	err = c.Insert(&LogEvent{time.Now(), event})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Event logged: %s\n\n",event)
}

func ExtractIP(ip string) string {
	ipv4regex := "(\\d{1,3}\\.){3}\\d{1,3}\\:\\d{1,5}"
	rx, _ := regexp.Compile(ipv4regex)
	if rx.MatchString(ip) {
		i := strings.Index(ip, ":")
		ip = ip[0:i]
	} else {
		ip = "0.0.0.0"
	}
	return ip
}

// for debug purposes
func encryptValue(enckey string, plaintext string) string {
	fmt.Println("Key: ", enckey)
	v, ok := algo.NewVigenère(enckey)
	if !ok {
		fmt.Println("Invalid key")
		return "unknown"
	}
	fmt.Println("Plain text:", plaintext)
	ct := v.Encipher(plaintext)
	return ct
}

// for debug purposes
func decryptValue(enckey string, ciphertext string) string {
	fmt.Println("Key: ", enckey)
	v, ok := algo.NewVigenère(enckey)
	if !ok {
		fmt.Println("Invalid key")
		return "unknown"
	}
	fmt.Println("Cipher text:", ciphertext)
	plaintext, ok := v.Decipher(ciphertext)
	return plaintext
}
