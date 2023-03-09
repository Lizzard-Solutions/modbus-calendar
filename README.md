# lizzard-modbus-calendar

```
?config={"hvac": ["HVAC 1","HVAC 2"], "mode": [{"mode": 1, "name": "Off"}, {"mode": 2, "name":
"On"}] }
```

function sendRequest_demo( method="GET", data) {
    let APIURL = "http://localhost:4000/datacalc"
    let xmlhttp = new XMLHttpRequest()
    // new HttpRequest instance
    xmlhttp.open(method, APIURL)
    xmlhttp.onreadystatechange = function () {
        // Call a function when the state changes.
        if (xmlhttp.readyState === XMLHttpRequest.DONE) {
            let response = xmlhttp.responseText
            if (response !== "") {
                console.log("Response",response)
            } else {
                console.log("Server connection error, Please try again.");
            }
        }
    }
    xmlhttp.send(JSON.stringify(data))
}
```
* PUT record - insert 
```
sendRequest_demo("PUT",{"FromTime":"999999","ToTime":"888888","Date":"777777","Temperature":"666666","Mode":"555555","Hvac":"444444"});
```
* POST record - update 
```
sendRequest_demo("POST",{"FromTime":"111","ToTime":"222","Date":"333","Temperature":"444","Mode":"555","Hvac":"666", "Id":"6400b7379b4742dc1713371f"});
```
* DELETE record - delete 
```
sendRequest_demo("DELETE",{"Id":"6400b7379b4742dc1713371f"});
```
* GET record - get records 
```
sendRequest_demo("GET",{}); // not required data 
```

## show db list 
```
 $show dbs
```

## create db
```
use remainders
```


## install packages in golang of mqtt
```go get github.com/eclipse/paho.mqtt.golang
```

### mosquitto Install 
```
    sudo systemctl status mosquitto #mqtt
    sudo apt install -y mosquitto-clients #clients of mqtt like rooms 
```

### subscribe to a topic
To subscribe to a topic, execute the `mosquitto_sub -t` command
```
    mosquitto_sub -t "/home/user/git/lizzard-modbus-calendar/mqttclients"
```

Open a second terminal window and don't close the first one. This time around, publish an "ON" message to the topic home/lights/sitting_room topic using the mosquitto_pub -m command
```
    mosquitto_pub -m "Message" -t "/home/user/git/lizzard-modbus-calendar/mqttclients"
```
