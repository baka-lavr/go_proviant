package main

import (
	"github.com/julienschmidt/httprouter"
)


func (app *Application) Route() *httprouter.Router {
	
	router := httprouter.New()
	//
	router.GET("/category", app.CategoryHandler)
	router.GET("/category/:categoryid/product", app.ProductHandler)
	router.POST("/order", app.OrderingHandler)
	//
	return router
}