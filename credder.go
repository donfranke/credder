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
    "encoding/json"
)

type Page struct {
	Title string
	Body  []byte
}

type Cred struct {
	ID         string
	SecretInfo string
	KeyID      string
}

type Cred2 struct {
	Appname   string
	Credid    string
	Secret    string
	Keyid      string
}

type CredRequest struct {
	Credid        string
	Appname string
} 

type LogEvent struct {
	Timestamp        time.Time
	EventDescription string
}

type Key struct {
	ID  string
	Key string
}

type CredList struct {
    Collection []Cred
}


var encKey string

const C_MONGODB = "127.0.0.1"
const C_ERROR_MESSAGE = "NOTFOUND"
const C_WEBSERVERPORT = "8889"

// use this web service call to get a specific encryption key
func KeyServer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
    w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
    w.Header().Set("Access-Control-Allow-Origin", "*")
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
	w.Header().Set("Content-Type", "application/json")
    w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
    w.Header().Set("Access-Control-Allow-Origin", "*")
	
	var result Cred
	r.ParseForm()

	// fetch values from request
	credID := r.FormValue("credid")
	appName := r.FormValue("appname")
	remoteIP := ExtractIP(r.RemoteAddr)
	userAgent := r.Header.Get("User-Agent")

	fmt.Printf("App Name: %s\n",appName)

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
		resultString = "[{\"credid\":\"" + credID + "\",\"secretinfo\":\"" + result.SecretInfo + "\",\"keyid\":\"" + result.KeyID + "\"}]"
		//resultString += ",{\"credid\":\"" + credID + "\",\"secretinfo\":\"" + result.SecretInfo + "\",\"keyid\":\"" + result.KeyID + "\"}"
	} else {
		resultString = "{\"secretinfo\":\"" + C_ERROR_MESSAGE + "\"}"
	}
	fmt.Printf("Sent to client: %s\n", resultString)

	io.WriteString(w, resultString)

	// decrypt  <-- DEBUG
	//plaintext := decryptValue(encKey,result)
	//fmt.Printf("Plaintext: %s\n",plaintext)
}

// use this web service call to retrieve a specific credential document
func CredListServer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
    w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
    w.Header().Set("Access-Control-Allow-Origin", "*")

	// get post variables
	decoder := json.NewDecoder(r.Body)
	var c CredRequest
	err := decoder.Decode(&c)
	if err!=nil {
		fmt.Println(err)
	}
	
	appName := c.Appname

	fmt.Println("APPNAME: " + c.Appname + " CREDID: " + c.Credid)

	remoteIP := ExtractIP(r.RemoteAddr)
	userAgent := r.Header.Get("User-Agent")

	// log event
	eventdesc := "\n******************************************\nRequest received for cred list (" + remoteIP + " " + userAgent + ") from " + appName + " at " + time.Now().Format(time.RFC850)
	logEvent(eventdesc)
	resultString := "["

	// 1. validate request 
	validateresult := validateCredRequest("0", appName, remoteIP, userAgent)
	fmt.Printf("Cred request validation result: %d\n", validateresult)

	if validateresult>0 {
		// display info
		fmt.Printf("=== CRED LIST ===\n")
		fmt.Printf("\tApp name: %s\n", appName)
		fmt.Printf("\tRemote IP: %s\n", remoteIP)
		fmt.Printf("\tUser Agent: %s\n", userAgent)
		result := getCredList()

		for _,element := range result {
			resultString += "{\"credid\":\"" + element.ID + "\",\"secretinfo\":\"" + element.SecretInfo + "\",\"keyid\":\"" + element.KeyID + "\"},"
		}
		resultString = resultString[:len(resultString)-1]
		resultString += "]"
	} 
	io.WriteString(w, resultString)
}


func AddCredServer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
    w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
    w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Printf("AddCredServer()\n")

	// get post variables
	decoder := json.NewDecoder(r.Body)
	var c Cred2
	err := decoder.Decode(&c)
	if err!=nil {
		fmt.Println(err)
	}

	fmt.Printf("Appname: %s\n",c.Appname)
	fmt.Printf("Credid: %s\n",c.Credid)
	fmt.Printf("Secret: %s\n",c.Secret)
	fmt.Printf("Keyid: %s\n",c.Keyid)

	// add to database
	session, err := mgo.Dial(C_MONGODB)
	if err != nil {
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	m := session.DB("credder").C("creds")

	err = m.Insert(&Cred{ID: c.Credid, SecretInfo: c.Secret, KeyID: c.Keyid})

	if err != nil {
		panic(err)
	}
}

func DelCredServer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
    w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
    w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Printf("AddCredServer()\n")

	// get post variables
	decoder := json.NewDecoder(r.Body)
	var c Cred2
	err := decoder.Decode(&c)
	if err!=nil {
		fmt.Println(err)
	}

	fmt.Printf("Appname: %s\n",c.Appname)
	fmt.Printf("Credid: %s\n",c.Credid)
	fmt.Printf("Secret: %s\n",c.Secret)
	fmt.Printf("Keyid: %s\n",c.Keyid)

	// add to database
	session, err := mgo.Dial(C_MONGODB)
	if err != nil {
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	m := session.DB("credder").C("creds")

	err = m.Remove(&Cred{ID: c.Credid, SecretInfo: c.Secret, KeyID: c.Keyid})

	if err != nil {
		panic(err)
	}
}

// use this web service call to retrieve a specific credential document
func KeyListServer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
    w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
    w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Println("Key List Server")

	// get post variables
	decoder := json.NewDecoder(r.Body)
	var c CredRequest
	err := decoder.Decode(&c)
	if err!=nil {
		fmt.Println(err)
	}
	
	appName := c.Appname

	fmt.Println("APPNAME: " + c.Appname + " CREDID: " + c.Credid)
	remoteIP := ExtractIP(r.RemoteAddr)
	userAgent := r.Header.Get("User-Agent")

	// log event
	eventdesc := "Request received for key list (" + remoteIP + " " + userAgent + ") from " + appName
	logEvent(eventdesc)
	resultString := "["

	// 1. validate request 
	validateresult := validateKeyRequest("0", appName, remoteIP, userAgent)
	fmt.Printf("Key request validation result: %d\n", validateresult)

	if validateresult>0 {
		// display info
		fmt.Printf("=== KEY LIST ===\n")
		fmt.Printf("\tApp name: %s\n", appName)
		fmt.Printf("\tRemote IP: %s\n", remoteIP)
		fmt.Printf("\tUser Agent: %s\n", userAgent)
		result := getKeyList()

		for _,element := range result {
			resultString += "{\"keyid\":\"" + element.ID + "\",\"key\":\"" + element.Key + "\"},"
		}
		resultString = resultString[:len(resultString)-1]
		resultString += "]"
	} 
	io.WriteString(w, resultString)
}

func main() {
	// init handlers
	http.HandleFunc("/cred", CredServer)
	http.HandleFunc("/key", KeyServer)
	http.HandleFunc("/credlist", CredListServer)
	http.HandleFunc("/keylist", KeyListServer)
	http.HandleFunc("/addcred", AddCredServer)
	http.HandleFunc("/delcred", DelCredServer)
	fmt.Printf("Web service started on port %s\n", C_WEBSERVERPORT)

	err := http.ListenAndServe(":" + C_WEBSERVERPORT, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
	fmt.Println("Web server running on port %d\n", C_WEBSERVERPORT)
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

	result := Key{}
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

func getCredList() []Cred {
	session, err := mgo.Dial(C_MONGODB)
	if err != nil {
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	c := session.DB("credder").C("creds")

	var result []Cred
	err = c.Find(nil).All(&result)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
	}
	//fmt.Println(result)
	return result
}

func getKeyList() []Key {
	fmt.Println("getKeyList")
	session, err := mgo.Dial(C_MONGODB)
	if err != nil {
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	c := session.DB("credder").C("keys")

	var result []Key
	err = c.Find(nil).All(&result)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
	}
	//fmt.Println(result)
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
	//fmt.Printf("Event logged: %s\n\n",event)
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
