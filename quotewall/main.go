package main

import (
	"html/template"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"net/http"
)

type Quote struct {
	Text   string
	Author string
}

func indexhandler(w http.ResponseWriter, r *http.Request, session *mgo.Session) {
	c := session.DB("quotewall").C("quotes")

	result := []Quote{}
	err := c.Find(bson.M{}).All(&result)

	if err != nil {
		panic(err)
	}

	err = templates.ExecuteTemplate(w, "index.html", result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

func newhandler(w http.ResponseWriter, r *http.Request, session *mgo.Session) {
	c := session.DB("quotewall").C("quotes")
	quote := Quote{}
	quote.Author = r.FormValue("author")
	quote.Text = r.FormValue("text")
	err := c.Insert(quote)

	if err != nil {
		panic(err)
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

/*
 * makeHandler warps around handlers providing them with a session, and automatically closing that session
 * to prevent loosing memory.
 */
func makeHandler(fn func(http.ResponseWriter, *http.Request, *mgo.Session), session *mgo.Session) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s := session.Copy() // Create a copy of the session
		fn(w, r, s)
		s.Close() // Close the copy
	}
}

// Global templates cache
var templates *template.Template

func main() {
	session, err := mgo.Dial("mongodb://localhost") // Create the master session
	if err != nil {
		panic(err)
	}
	templates = template.Must(template.ParseFiles("index.html")) // Parse the template cache
	http.HandleFunc("/", makeHandler(indexhandler, session))
	http.HandleFunc("/new/", makeHandler(newhandler, session))
	http.ListenAndServe(":8080", nil)
}
