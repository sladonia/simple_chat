package controllers

import (
	"log"
	"net/http"
)

var (
	FileController FileControllerInterface = &fileController{}
)

type FileControllerInterface interface {
	ServeChatHTML(w http.ResponseWriter, r *http.Request)
	ServeLogInHTML(w http.ResponseWriter, r *http.Request)
}

type fileController struct{}

func (f *fileController) ServeChatHTML(w http.ResponseWriter, r *http.Request) {
	log.Printf("serving chat.html")
	http.ServeFile(w, r, "html/chat.html")
}

func (f *fileController) ServeLogInHTML(w http.ResponseWriter, r *http.Request) {
	log.Printf("serving login.html")
	http.ServeFile(w, r, "html/login.html")
}
