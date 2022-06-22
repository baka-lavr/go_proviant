package main

import (
	"time"
	"context"
	"log"
	"encoding/json"
	"net/http"
	"github.com/julienschmidt/httprouter"
)

func (app *Application) CategoryHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	data, err := app.GetCategory()
	if err != nil {
		log.Println(err)
		http.Error(w, "Error", 500)
	}
	js, _ := json.Marshal(data)
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func (app *Application) ProductHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	//files := []string{"./Front/html/index.html"}
	ord := r.URL.Query().Get("order")
	cat := params.ByName("categoryid")
	data, err := app.GetProducts(cat, ord)
	if err != nil {
		log.Println(err)
		http.Error(w, "Error", 500)
	}
	js, _ := json.Marshal(data)
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func (app *Application) OrderingHandler(w http.ResponseWriter, r *http.Request, _params httprouter.Params) {
	js := r.Body
	//js := []byte(`
	//{"orderItems": [{"productId": 1, "weight": 3},{"productId": 2, "weight": 1}], "client": {"surname": "Shcherbinin", "name": "Fedor", "phoneNumber": 88005553535, "email": "v88555@mail.ru"}}
	//`)
	var ord Order
	_ = json.NewDecoder(js).Decode(&ord)

	ctx, cancel := context.WithTimeout(context.WithValue(context.Background(), "data", ord), time.Second*5)
	defer cancel()
	res := make(chan int)

	go app.CreateOrder(ctx, res)
	select {
	case <- ctx.Done():
		http.Error(w, "Error", 500)
		log.Println(ctx.Err())
	case <- res:
		log.Println("Order done")
	}
}