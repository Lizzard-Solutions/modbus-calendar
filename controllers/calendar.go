package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"reakgo/models"
	"reakgo/utility"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
)

func AjaxJsonDecoder(r *http.Request) (models.Data, error) {
	var decodeJson models.Data
	log.Println(r.Body)

	err := json.NewDecoder(r.Body).Decode(&decodeJson)
	log.Println(err)
	return decodeJson, err
}

func GetReminder(hvac string) ([]bson.M, error) {

	// get Client, Context, CancelFunc and err from connect method.
	client, ctx, cancel, err := utility.Connect()
	if err != nil {
		return []bson.M{}, err
	}
	// Release resource when main function is returned.
	defer utility.Close(client, ctx, cancel)
	var filter, option interface{}

	filter = bson.D{{"Hvac", hvac}}

	option = bson.D{}
	// call the query method with client, context,
	// database name, collection  name, filter and option
	// This method returns momngo.cursor and error if any.
	results, err := Db.calendar.Get(client, ctx, "reminder", filter, option) // handle the errors.
	/*
		// printing the result of query.
		log.Println("Query Result")
		for _, doc := range results {
			log.Println(doc)
		}
	*/
	return results, err
}

func PutReminder(r *http.Request) (models.Data, error) {
	dataAjax, err := AjaxJsonDecoder(r)
	log.Println(dataAjax)
	if err != nil {
		return dataAjax, err
	}
	// get Client, Context, CancelFunc and err from connect method.
	client, ctx, cancel, err := utility.Connect()
	if err != nil {
		return dataAjax, err
	}
	// Release resource when main function is returned.
	defer utility.Close(client, ctx, cancel)

	var document interface{}
	document = bson.D{{"FromTime", dataAjax.FromTime}, {"ToTime", dataAjax.ToTime}, {"Date", dataAjax.Date}, {"Temperature", dataAjax.Temperature}, {"Mode", dataAjax.Mode}, {"Hvac", dataAjax.Hvac}, {"Day", dataAjax.Day}}

	_, err = Db.calendar.Put(client, ctx, "reminder", document) // handle the errors.
	if err != nil {
		return dataAjax, err
	} else {
		utility.SentMQttMessageToClient("add", dataAjax)
		return dataAjax, nil
	}
}

func PostReminder(r *http.Request) (models.Data, error, int64) {
	dataAjax, err := AjaxJsonDecoder(r)
	if err != nil {
		return dataAjax, err, int64(0)
	}
	// get Client, Context, CancelFunc and err from connect method.
	client, ctx, cancel, err := utility.Connect()
	if err != nil {
		return dataAjax, err, int64(0)
	}
	// Release resource when main function is returned.
	defer utility.Close(client, ctx, cancel)

	log.Println(dataAjax.Id.Hex())
	log.Println(dataAjax.Id)
	filterid := bson.D{{"_id", dataAjax.Id}}
	var document interface{}
	document = bson.D{{"FromTime", dataAjax.FromTime}, {"ToTime", dataAjax.ToTime}, {"Date", dataAjax.Date}, {"Temperature", dataAjax.Temperature}, {"Mode", dataAjax.Mode}, {"Hvac", dataAjax.Hvac}, {"Day", dataAjax.Day}}
	update := bson.D{{"$set", document}}

	result, err := Db.calendar.PostOne(client, ctx, "reminder", filterid, update)
	log.Println(result)
	// handle the errors.
	/* if we needed to check print affected data result
	log.Println(result.MatchedCount)
	log.Println(result.ModifiedCount)
	*/
	if err != nil && !(result.MatchedCount > 0) && !(result.ModifiedCount > 0) {
		return dataAjax, err, int64(0)
	} else {
		utility.SentMQttMessageToClient("update", dataAjax)
		return dataAjax, nil, result.ModifiedCount
	}
}

func DeleteReminder(r *http.Request) (models.Data, error, int64) {
	dataAjax, err := AjaxJsonDecoder(r)
	if err != nil {
		return dataAjax, err, int64(0)
	}
	// get Client, Context, CancelFunc and err from connect method.
	client, ctx, cancel, err := utility.Connect()
	if err != nil {
		return dataAjax, err, int64(0)
	}
	// Release resource when main function is returned.
	defer utility.Close(client, ctx, cancel)

	filterid := bson.D{{"_id", dataAjax.Id}}

	result, err := Db.calendar.DeleteOne(client, ctx, "reminder", filterid) // handle the errors.
	/* if we needed to check print affected data result
	log.Println(result.MatchedCount)
	log.Println(result.ModifiedCount)
	*/
	if err == nil {
		utility.SentMQttMessageToClient("delete", dataAjax)
	}
	return dataAjax, err, result.DeletedCount
}

func GetData(w http.ResponseWriter, r *http.Request) {
	hvac := r.URL.Query().Get("hvac")
	response := utility.AjaxResponse{Status: "failure", Message: "Sorry! Unable to get records from server"}
	dataAjax, err := GetReminder(hvac)
	if err != nil {
		response.Status = "failure"
		response.Message = "Getting empty records from server"
		response.Payload = []bson.D{}
	} else {
		response.Status = "success"
		response.Message = "Successfull getting records"
		response.Payload = dataAjax
	}
	json, _ := json.Marshal(response)
	w.Write([]byte(json))
}

func PutData(w http.ResponseWriter, r *http.Request) {
	response := utility.AjaxResponse{Status: "failure", Message: "Sorry! Unable to get records from server"}
	// log.Println(r.Body)
	dataAjax, err := PutReminder(r)
	if err != nil {
		log.Println("[error] - put records ", err)
		response.Status = "failure"
		response.Message = "Failed to put records on db"
		response.Payload = models.Data{}
	} else {
		response.Status = "success"
		response.Message = "Successfully records inserted on db"
		response.Payload = dataAjax
	}

	// log.Println(err)
	json, _ := json.Marshal(response)
	w.Write([]byte(json))
}

func PostData(w http.ResponseWriter, r *http.Request) {
	response := utility.AjaxResponse{Status: "failure", Message: "Sorry! Unable to get records from server"}
	// log.Println(r.Body)
	dataAjax, err, modifiedCount := PostReminder(r)
	if err != nil {
		log.Println("[error] - post records ", err)
		response.Status = "warning"
		response.Message = "Records not to be update on db"
	} else if modifiedCount > 0 {
		response.Status = "success"
		response.Message = "Successfully " + strconv.Itoa(int(modifiedCount)) + " records update on db"
	} else {
		response.Status = "failure"
		response.Message = "Failed to update records on db"
	}
	response.Payload = dataAjax

	json, _ := json.Marshal(response)
	w.Write([]byte(json))
}

func DeleteData(w http.ResponseWriter, r *http.Request) {
	response := utility.AjaxResponse{Status: "failure", Message: "Sorry! Unable to get records from server"}
	// log.Println(r.Body)
	dataAjax, err, deletedCount := DeleteReminder(r)
	if err != nil {
		log.Println("[error] - delete records ", err)
		response.Status = "warning"
		response.Message = "Records not to be delete on db"
	} else if deletedCount > 0 {
		response.Status = "success"
		response.Message = "Successfully records deleted on db"

	} else {
		response.Status = "failure"
		response.Message = "Failed to delete records on db"
	}
	response.Payload = dataAjax
	json, _ := json.Marshal(response)
	w.Write([]byte(json))
}

func Calendar(w http.ResponseWriter, r *http.Request) {
	utility.RenderTemplate(w, r, "calendar", nil)
}
