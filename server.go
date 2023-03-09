package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"crypto/sha256"
    "crypto/subtle"

	// "html/template" replace with "text/template" reason tmpl security error solov
	"text/template"
	"time"

	"reakgo/router"
	"reakgo/utility"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
)

func init() {
	var err error
	err = godotenv.Load()
	if err != nil {
		log.Println(".env file wasn't found, looking at env variables")
	}
	motd()
	// Read Config
	utility.Db, err = sqlx.Open("mysql", os.Getenv("DB_USER")+":"+os.Getenv("DB_PASSWORD")+"@/"+os.Getenv("DB_NAME"))
	if err != nil {
		log.Println("Wowza !, We didn't find the DB or you forgot to setup the env variables")
		//panic(err)
	}

	utility.Store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))
	utility.View = cacheTemplates()
	// See "Important settings" section.
	utility.Db.SetConnMaxLifetime(time.Minute * 3)
	utility.Db.SetMaxOpenConns(10)
	utility.Db.SetMaxIdleConns(10)
	utility.Client = utility.StartMqtt()

}

func main() {
	http.HandleFunc("/", basicAuth(handler))
	// Serve static assets
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets"))))
	http.Handle("/upload/", http.StripPrefix("/upload/", http.FileServer(http.Dir("./upload"))))
	log.Fatal(http.ListenAndServe(":"+os.Getenv("WEB_PORT"), nil))
}

func cacheTemplates() *template.Template {
	templ := template.New("")
	err := filepath.Walk("./templates", func(path string, info os.FileInfo, err error) error {
		if strings.Contains(path, ".html") {
			_, err = templ.ParseFiles(path)
			if err != nil {
				log.Println(err)
			}
		}

		return err
	})

	if err != nil {
		panic(err)
	}

	return templ
}

func handler(w http.ResponseWriter, r *http.Request) {
	router.Routes(w, r)
}

func motd() {
	logo := `
______ _____  ___   _   __
| ___ \  ___|/ _ \ | | / /
| |_/ / |__ / /_\ \| |/ / 
|    /|  __||  _  ||    \ 
| |\ \| |___| | | || |\  \
\_| \_\____/\_| |_/\_| \_/
                          
----------------------------
Application should now be accessible on port ` + os.Getenv("WEB_PORT") + `

`
	log.Println(logo)
}

func basicAuth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if ok {
			usernameHash := sha256.Sum256([]byte(username))
			passwordHash := sha256.Sum256([]byte(password))
			expectedUsernameHash := sha256.Sum256([]byte(os.Getenv("FROM_EMAIL")))
			expectedPasswordHash := sha256.Sum256([]byte(os.Getenv("PASSWORD")))

			usernameMatch := (subtle.ConstantTimeCompare(usernameHash[:], expectedUsernameHash[:]) == 1)
			passwordMatch := (subtle.ConstantTimeCompare(passwordHash[:], expectedPasswordHash[:]) == 1)

			if usernameMatch && passwordMatch {
				next.ServeHTTP(w, r)
				return
			}
		}

		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}