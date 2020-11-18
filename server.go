package main

import (
	"go.uber.org/zap"
	"html/template"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// stores all connected clients
var clients = make(map[*websocket.Conn]bool)
// channel that receives updated number of visitors
var broadcast = make(chan int)

var logger *zap.SugaredLogger

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	logging, _ := zap.NewProduction()
	logger = logging.Sugar()

	router := mux.NewRouter()
	router.HandleFunc("/", rootHandler).Methods("GET")
	router.HandleFunc("/ws", wsHandler)
	go handleBroadcast()

	logger.Infof("starting service on port 8808")
	logger.Fatal(http.ListenAndServe(":8808", router))
}

type PageData struct {
	Visitors int
}

func rootHandler(w http.ResponseWriter, _ *http.Request) {
	tmpl, err := template.ParseFiles("static/index.html")
	if err != nil {
		logger.Errorf("failed to open template file %v", err)
	}
	// +1 because current user hasn't connected to websocket at that moment yet
	data := PageData{Visitors: len(clients) + 1}

	err = tmpl.Execute(w, data)
	if err != nil {
		logger.Errorf("failed to render template %v", err)
	}
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Errorf("failed to create ws connection %v", err)
	}

	// returns error if client disconnects
	go func() {
		for {
			_, _, err := ws.ReadMessage()
			if err != nil {
				delete(clients, ws)
				logger.Infof("client disconnected, clients: %d", len(clients))
				broadcast <- len(clients)
				break
			}
		}
	}()

	// register client
	clients[ws] = true
	broadcast <- len(clients)
	logger.Infof("client connected, clients: %d", len(clients))
}

// sends message about visitors number change to all clients
func handleBroadcast() {
	for {
		val := <-broadcast
		for client := range clients {
			err := client.WriteMessage(websocket.TextMessage, []byte(strconv.Itoa(val)))
			if err != nil {
				logger.Errorf("Websocket error: %v", err)
				err = client.Close()
				if err != nil {
					logger.Errorf("failed to close connection: %v", err)
				}
				delete(clients, client)
			}
		}
	}
}
