package utility

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"strings"
	"time"

	// "html/template" replace with "text/template" reason tmpl security error solov
	"text/template"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Template Pool
var View *template.Template

// Session Store
var Store *sessions.CookieStore

// DB Connections
var Db *sqlx.DB

type Session struct {
	Key   string
	Value string
}
type AjaxResponse struct {
	Status  string
	Message string
	Payload interface{}
}

// mqtt client
var Client mqtt.Client

var ConnectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("MQTT Connected successfully")
}

func RedirectTo(w http.ResponseWriter, r *http.Request, url string) {
	http.Redirect(w, r, url, http.StatusFound)
}

func SessionSet(w http.ResponseWriter, r *http.Request, data []Session) {
	session, _ := Store.Get(r, os.Getenv("SESSION_NAME"))
	// Set some session values.
	session.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   6000,
		HttpOnly: true,
	}
	for _, dataSingle := range data {
		session.Values[dataSingle.Key] = dataSingle.Value
	}
	// Save it before we write to the response/return from the handler.
	err := session.Save(r, w)
	if err != nil {
		log.Println(err)
	}
}

func SessionGet(r *http.Request, key string) interface{} {
	session, _ := Store.Get(r, os.Getenv("SESSION_NAME"))
	// Set some session values.
	return session.Values[key]
}

func fetchSession(r *http.Request) map[interface{}]interface{} {
	session, _ := Store.Get(r, os.Getenv("SESSION_NAME"))
	return session.Values
}

func CheckACL(w http.ResponseWriter, r *http.Request, minLevel int) bool {
	userType := SessionGet(r, "type")
	var level int = 0
	switch userType {
	case "user":
		level = 1
	case "admin":
		level = 2
	default:
		level = 0
	}
	if level >= minLevel {
		return true
	} else {
		RedirectTo(w, r, os.Getenv("APPURL"))
		AddFlash("Failure", "Access Denied.", w, r)
		return false
	}
}

func AddFlash(flavour string, message string, w http.ResponseWriter, r *http.Request) {
	session, err := Store.Get(r, "flash-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	flash := make(map[string]string)
	flash["Flavour"] = flavour
	flash["Message"] = message
	session.AddFlash(flash, "message")
	session.Save(r, w)
}

func viewFlash(w http.ResponseWriter, r *http.Request) interface{} {
	session, err := Store.Get(r, "flash-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	fm := session.Flashes("message")
	if fm == nil {
		return nil
	}
	session.Save(r, w)
	return fm
}

func RenderTemplate(w http.ResponseWriter, r *http.Request, template string, data interface{}) {
	tmplData := make(map[string]interface{})
	config := make(map[string]interface{})
	tmplData["data"] = data
	tmplData["flash"] = viewFlash(w, r)
	tmplData["session"] = fetchSession(r)
	//Config option
	config["APPURL"] = os.Getenv("APPURL")
	config["APIURL"] = os.Getenv("APIURL")
	config["ASSETS_DIR"] = os.Getenv("ASSETS_DIR")
	config["LogIn_Status_Check_Interval"] = os.Getenv("LogIn_Status_Check_Interval")
	tmplData["config"] = config
	View.ExecuteTemplate(w, template, tmplData)
}

func HandlePanic(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			r := recover()
			// Marshall r and log
			//http.Error(w, "", http.StatusInternalServerError)

			if r != nil {
				json, _ := json.Marshal(r)
				log.Println("Recoverd in HandlePanic - ", string(json))
				w.WriteHeader(500)
				View.ExecuteTemplate(w, "500_error", nil)
			}
		}()
		h.ServeHTTP(w, r)
	})
}

/* smtp send email*/
func SendEmailSMTP(to []string, subject string, body string) (bool, error) {
	//Sender data.
	from := os.Getenv("FROM_EMAIL")
	// Set up email information.
	header := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n"
	msg := []byte("From: " + from + "\n" + "To: " + strings.Join(to, ",") + "\n" + "Subject: " + subject + "\r\n" + header + body)
	// Sending email.
	// fmt.Println("From: " + from + "\n" + "To: " + strings.Join(to, ",") + "\n" + "Subject: " + subject + "\r\n" + header + "\r\n" + body)
	err := smtp.SendMail(os.Getenv("SMTP_HOST")+":"+os.Getenv("SMTP_PORT"), smtp.PlainAuth("", os.Getenv("FROM_USER"), os.Getenv("PASSWORD"), os.Getenv("SMTP_HOST")), from, to, msg)
	if err != nil {
		return false, err
	}
	return true, nil
}

/*
* main function to call for send mail
#input

	to : string array
	template : tempatePath
	data : associative array of data which is set on template, by-default app_url and app_name is set
*/
func SendEmail(to []string, template string, data map[string]string) (bool, error) {
	buf := new(bytes.Buffer)
	//extra information on email
	data["app_url"] = os.Getenv("APPURL")
	data["app_name"] = os.Getenv("APPNAME")
	// Set up email information.
	err := View.ExecuteTemplate(buf, template, data)
	if err != nil {
		return false, err
	}
	return SendEmailSMTP(to, data["subject"], buf.String())
}

func SessionDestroy(w http.ResponseWriter, r *http.Request) bool {
	session, err := Store.Get(r, os.Getenv("SESSION_NAME"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return false
	}
	session.Options.MaxAge = -1
	err = session.Save(r, w)
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

// This is a user defined method that returns mongo.Client,
// context.Context, context.CancelFunc and error.
// mongo.Client will be used for further database operation.
// context.Context will be used set deadlines for process.
// context.CancelFunc will be used to cancel context and
// resource associated with it.
// to -do have to read this
func Connect() (*mongo.Client, context.Context, context.CancelFunc, error) {
	uri := os.Getenv("DBConnection")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Println(err)
	}
	return client, ctx, cancel, err
}

// This is a user defined method to close resources.
// This method closes mongoDB connection and cancel context.
// to -do have to read this
func Close(client *mongo.Client, ctx context.Context, cancel context.CancelFunc) {
	defer cancel()

	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			log.Println(err)
		}
	}()
}

func SentMQttMessageToClient(object string, messageJson interface{}) {
	// topic/test
	topic := os.Getenv("MQttTopicPath")
	data := map[string]interface{}{"Object": object, "Data": messageJson}
	message := JsonEncode(data)
	token := Client.Publish(topic, 0, false, message)
	token.Wait()
	time.Sleep(time.Second)
}

func JsonEncode(str_msg interface{}) string {
	// response := map[string]interface{}{"status": "success"}
	encodingJson, err := json.Marshal(str_msg)
	if err != nil {
		log.Println(err)
	}
	return string([]byte(encodingJson))
}
func StartMqtt() mqtt.Client {
	var broker = os.Getenv("MQTT_BROKER_URL")
	var port = os.Getenv("MQTT_BROKER_PORT")
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", broker, port))
	opts.SetClientID("go_mqtt_client")
	opts.SetUsername("emqx")
	opts.SetPassword("public")
	opts.OnConnect = ConnectHandler
	Client = mqtt.NewClient(opts)
	if token := Client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	return Client
}

func StartMQttSub() bool {
	topic := os.Getenv("MQttTopicPath")
	token := Client.Subscribe(topic, 1, nil)
	token.Wait()
	fmt.Printf("Subscribed to topic: %s", topic)
	return true
}
