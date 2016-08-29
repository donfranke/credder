// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"html/template"
	"io/ioutil"
	"net/http"
	"regexp"
	"gopkg.in/mgo.v2"
	"fmt"
	"log"
	"time"
	"strings"
	"gopkg.in/mgo.v2/bson"
)

const C_MONGODB = "127.0.0.1"

var templates = template.Must(template.ParseFiles("credlist.html"))
var validPath = regexp.MustCompile("^/(save|credlist)/([a-zA-Z0-9]+)$")

func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func credlistHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	logEvent("Request for credlist made by " + ExtractIP(r.RemoteAddr))
	if err != nil {
		//http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "credlist", p)
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}

func main() {
	http.HandleFunc("/credlist/", makeHandler(credlistHandler))
	http.ListenAndServe(":8080", nil)
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

func getCred(id string) string {
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

func delCred(id string) string {
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
