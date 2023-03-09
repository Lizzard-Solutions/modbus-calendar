package router

import (
	"net/http"
	"reakgo/controllers"
	//"reakgo/utility"
	"strings"
)

func Routes(w http.ResponseWriter, r *http.Request) {
	// Trailing slash is a pain in the ass so we just drop it
	route := strings.Trim(r.URL.Path, "/")
	switch route {
	case "", "index", "login", "singin":
		//utility.CheckACL(w, r, 0)
		//controllers.Login(w, r)
		controllers.Calendar(w, r)
	case "datacalc":
		if r.Method == "GET" {
			controllers.GetData(w, r)
		}
		if r.Method == "PUT" {
			controllers.PutData(w, r)
		}
		if r.Method == "POST" {
			controllers.PostData(w, r)
		}
		if r.Method == "DELETE" {
			controllers.DeleteData(w, r)
		}
	case "calendar":
		controllers.Calendar(w, r)

	default:
		controllers.Error404(w, r)
	}

}
