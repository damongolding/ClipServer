package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/gen2brain/beeep"
	"golang.design/x/clipboard"
)

const port = 8090

type JsonResponse struct {
	Timestamp string `json:"timestamp"`
	Code      int    `json:"code"`
	Success   bool   `json:"success"`
	Message   string `json:"message"`
}

func init() {
	err := clipboard.Init()
	if err != nil {
		log.Fatalf(err.Error())
		panic(err)
	}
}

func GetLocalIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	localAddress := conn.LocalAddr().(*net.UDPAddr)

	return localAddress.IP
}

func main() {

	mux := http.NewServeMux()

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		s := fmt.Sprintf("ClipServer is running %s:%d", GetLocalIP(), port)
		_, err := fmt.Fprint(w, s)
		if err != nil {
			return
		}
	})

	mux.HandleFunc("POST /add", func(w http.ResponseWriter, r *http.Request) {

		err := r.ParseForm()
		if err != nil {
			return
		}

		postData := r.Form.Get("clip")
		postData = strings.TrimSpace(postData)

		if postData == "" {
			log.Error("missing clip form data")
			return
		}

		data, err := base64.StdEncoding.DecodeString(postData)
		if err != nil {
			log.Error(err)
			return
		}
		dataString := string(data)

		log.Info("received post request", "message", dataString)

		w.Header().Add("content-type", "application/json")

		outJson := JsonResponse{
			Timestamp: time.Now().String(),
			Code:      200,
			Success:   true,
			Message:   dataString,
		}

		out, err := json.Marshal(outJson)
		if err != nil {
			return
		}

		clipboard.Write(clipboard.FmtText, data)

		confirm := clipboard.Read(clipboard.FmtText)
		if string(confirm) != dataString {
			log.Error("missmatch", "confirm", string(confirm), "data", dataString)
			return
		}

		_, err = w.Write(out)
		if err != nil {
			log.Error(err.Error())
			return
		}

		err = beeep.Notify("Added to clipboard", dataString, "/Users/damon/GIT/clipserver/assets/iconfile.png")
		if err != nil {
			log.Error(err.Error())
			return
		}
	})

	fmt.Printf("Listening on %s:%d\n", GetLocalIP(), port)

	addr := fmt.Sprintf(":%d", port)

	err := http.ListenAndServe(addr, mux)
	if err != nil {
		log.Fatalf(err.Error())
		return
	}
}
