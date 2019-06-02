package main

import (
	"encoding/json"
	"log"

	"golang.org/x/net/websocket"
)

type subscriptionOHLC struct {
	Name     string `json:"name"`
	Interval string `json:"interval"`
	Depth    string `json:"depth"`
}
type subscriptionTicker struct {
	Name string `json:"name"`
}
type subscriptionPayload struct {
	Event        string             `json:"event"`
	Pair         []string           `json:"pair"`
	Subscription subscriptionTicker `json:"subscription"`
}

type response struct {
	Heartbeat      map[string]interface{}
	TickerUpdate   []interface{}
	UnmarshaledMap map[string][]float64
}

const (
	krakenURL = "wss://ws.kraken.com"
	origin    = "http://localhost"
)

var channels map[float64]string

func send(conn *websocket.Conn, msg []byte) (bool, error) {
	if n, e := conn.Write(msg); e != nil {
		log.Println(e.Error())
		return false, e
	} else if n == 0 {
		return false, nil
	}
	return true, nil
}

// The response fetched from streams is NOT consistent
// hence this loose parsing is established as step one.
// If you expect json, reflect will asert a map
// from which you can further check the types of the map values,
// an example of that is krakenTickerSorter().
func responsePayload(b []byte) (interface{}, error) {
	if len(b) == 0 {
		log.Println("Nothing to unmarshal.")
	}
	var payload interface{}
	if e := json.Unmarshal(b, &payload); e != nil {
	}
	return payload, nil
}

func populateResponseObject(data interface{}) {
	switch returnType(data) {
	case "map[string]interface {}":
		m := data.(map[string]interface{})
		switch m["event"] {
		case "systemStatus":
			log.Println(m["status"], m["connectionID"])
		case "subscriptionStatus":
			log.Println(m["pair"], m["status"], m["channelID"])
			channels[m["channelID"].(float64)] = m["pair"].(string)
		}
	case "[]interface {}":
		log.Println("Received update.")
		s := data.([]interface{})
		updateMap := s[1].(map[string]interface{})
		//log.Println(updateMap)
		ParsedKrakenTicker := krakenTickerSorter(updateMap, channels[s[0].(float64)])
		log.Println(ParsedKrakenTicker) // to be stored ...
	}
}

func listenForUpdates(ws *websocket.Conn, data []byte) {
	n, e := ws.Read(data)
	if e != nil {
		log.Println(e.Error())
	}
	response := data[:n]
	payload, e := responsePayload(response)
	if e != nil {
		log.Println("Unmarshaler:", e.Error())
		return
	}
	populateResponseObject(payload)
}

func clientBasic(url, origin string) *websocket.Conn {
	conn, e := websocket.Dial(url, "", origin)
	if e != nil {
		log.Println(e.Error())
	}
	if conn.IsClientConn() {
		log.Println("Connected.")
		return conn
	}
	return nil
}

// According to https://www.kraken.com/en-us/features/websocket-api#message-ticker
type krakenTicker struct {
	Ask                        []string `json:"a"`
	Bid                        []string `json:"b"`
	Close                      []string `json:"c"`
	Volume                     []string `json:"v"`
	VolumeWeightedAveragePrice []string `json:"p"`
	NumberOfTrades             []int64  `json:"t"`
	LowPrice                   []string `json:"l"`
	HighPrice                  []string `json:"h"`
	OpenPrice                  []string `json:"o"`
}

func krakenTickerSorter(m map[string]interface{}, pair string) krakenTicker {
	var ticker krakenTicker
	b, e := json.Marshal(&m)
	if e != nil {
		log.Println(e.Error())
	}
	if e = json.Unmarshal(b, &ticker); e != nil {
		log.Println(e.Error())
	}
	return ticker
}

func main() {
	channels = make(map[float64]string)
	subdef := subscriptionTicker{Name: "ticker"}
	sub := subscriptionPayload{Event: "subscribe", Pair: []string{"XBT/EUR", "XBT/USD"}, Subscription: subdef}
	ws := clientBasic(krakenURL, origin)
	b, e := json.Marshal(&sub)
	if e != nil {
		log.Println(e.Error())
		return
	}
	if success, e := send(ws, b); e == nil && success {
		data := make([]byte, 2048)
		for ws.IsClientConn() {
			listenForUpdates(ws, data)
		}
	} else {
		log.Println(e.Error())
		return
	}
}
