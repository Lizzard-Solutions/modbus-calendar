package controllers

import (
	"encoding/json"
	"net/http"
	"reakgo/utility"
)

func Dashboard(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]interface{})
	data["Status"] = "success"
	data["Data"] = "aa"
	// data["UserName"] = fmt.Sprintf("%v", utility.SessionGet(r, "UserName"))
	utility.SentMQttMessageToClient("insert", data)
	json, _ := json.Marshal(data)
	w.Write([]byte(json))
}
