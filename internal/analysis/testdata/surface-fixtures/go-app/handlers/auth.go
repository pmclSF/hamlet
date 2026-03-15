package handlers

import "net/http"

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}

type UserController struct{}

func (c *UserController) GetUser(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("user"))
}

func (c *UserController) DeleteUser(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("deleted"))
}

func internalHelper() string {
	return "private"
}
