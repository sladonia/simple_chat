package controllers

import (
	"log"
	"net/http"
)

var (
	FileController FileControllerInterface = &fileController{}
)

type FileControllerInterface interface {
	ServeHTML(w http.ResponseWriter, r *http.Request)
}

type fileController struct{}

func (f *fileController) ServeHTML(w http.ResponseWriter, r *http.Request) {
	log.Printf("serving chat.html")
	http.ServeFile(w, r, "html/chat.html")
}
