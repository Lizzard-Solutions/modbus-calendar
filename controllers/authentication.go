package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"reakgo/utility"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

/*get username and password with env file
@output env username and password
*/

func getEnvUserNameAndPassword() (string, string) {
	return "user@reak.in", "$2a$10$XP5NfeBXgL36rXpsv5eDIeY7zrJzxu6MDeIbQpRns3DjwVG/dmU9K"
}

func Login(w http.ResponseWriter, r *http.Request) {
	if (r.Method) == "POST" {
		err := r.ParseForm()
		if err != nil {
			log.Println(err)
		} else {
			username := r.FormValue("UserName")
			password := r.FormValue("Password")
			if username == "" || password == "" {
				log.Println("Please Fill all required fields")
			} else {
				/* get username and password from env file and Compare */
				env_username, env_password := getEnvUserNameAndPassword()
				err := bcrypt.CompareHashAndPassword([]byte(env_password), []byte(password))
				checkusername := strings.Compare(username, env_username)
				// 0 if eqal
				if err == nil && checkusername == int(0) {
					sessionData := []utility.Session{
						{Key: "Id", Value: strconv.Itoa(int(time.Now().Unix()))},
						{Key: "UserName", Value: username},
						{Key: "Type", Value: "user"},
					}
					utility.SessionSet(w, r, sessionData)
					// added data to database when user login first time
					utility.RedirectTo(w, r, `calendar?config={"hvac"%3A 1 %2C "mode"%3A [{"mode"%3A 1%2C "name"%3A "Off"}%2C {"mode"%3A 2%2C "name"%3A "On"}] }`)
				} else {
					log.Println(err)
					utility.AddFlash("Failure", "The Entered credentials are incorrect.", w, r)
					log.Println("Failure", "The Entered credentials are incorrect.", w, r)
				}
			}
		}
	}
	utility.RenderTemplate(w, r, "login", nil)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	res := utility.SessionDestroy(w, r)
	data := make(map[string]string)
	if res {
		data["Status"] = "success"
	} else {
		data["Status"] = "failure"
	}
	json, _ := json.Marshal(data)
	w.Write([]byte(json))
}

func CheckLoginStatus(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("loggedin"))
}

func Error404(w http.ResponseWriter, r *http.Request) {
	utility.RenderTemplate(w, r, "error404", nil)
}
