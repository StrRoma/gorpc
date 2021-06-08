package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/api/Get/Price/{pair}/{decimals}/{swap}", GetPrice).Methods("GET")
	// router.HandleFunc("/api/Get/Price/{pair}/{decimals}", GetPrice).Methods("GET") // old method

	srv := &http.Server{
		Addr:         ":9000",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		// IdleTimeout: 120 * time.Second,
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf(err.Error())
		}
	}()

	// Ожидание сигнала
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	// Попытка корректного завершения
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)

	log.Printf("closes\n")
}

var GetPrice = func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)

	fmt.Println(" request", "/api/Get/Price/{pair}/{decimals}/{swap}", params["pair"], params["decimals"], params["swap"])

	d, _ := strconv.ParseInt(params["decimals"], 10, 64)

	var price float64
	var err error

	if params["swap"] == "uni" || params["swap"] == "" {
		price, err = getPriceUni(params["pair"], d)
	} else {
		price, err = getPriceCake(params["pair"], d)
	}

	if err != nil {
		Respond(w, Message(true, "Error in getPrice" + "   Err: "+err.Error()))
		return
	}
	// fmt.Println(data)

	resp := Message(false, "success")
	resp["data"] = price
	Respond(w, resp)
}

func Message(status bool, message string) map[string]interface{} {
	return map[string]interface{}{"isError": status, "message": message}
}

func Respond(w http.ResponseWriter, data map[string]interface{}) {
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
