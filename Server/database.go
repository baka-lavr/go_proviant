package main

import (
	"os"
	"fmt"
	"log"
	"database/sql"
	"time"
	"github.com/lib/pq"
	"context"
)

type Product struct {
	Id int `json: "id"`
	Category_id int `json: "category_id"`
	Name string `json: "name"`
	Name_en string `json: "name_en"`
	Amount int `json: "amount"`
	Image_url string `json: "product_url"`
	Meas_unit string `json: "measure_unit"`
	Price int `json: "price"`
}

type Category struct {
	Id int `json: "id"`
	Name string `json: "name"`
	Name_en string `json: "nameEnglish"`
	Image_url string `json: "imagePath"`
}

type Order struct {
	Client struct {
		Surname string `json: "surname"`
		Name string `json: "name"`
		PhoneNumber int `json: "phoneNumber"`
		Email string `json: "email"`
	} `json: "client"`
	OrderItems []struct {
		Weight int `json: "weight"`
		ProductId int `json: "productId"`
	} `json: "orderItems"`
	Total float32 `json: "total"`
}

func OpenDB() (*sql.DB, error) {
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASS")
	name := os.Getenv("DB_NAME")
	line := fmt.Sprintf("host=localhost port=5432 user=%s password=%s dbname=%s sslmode=disable",user,pass,name)
	db, err := sql.Open("postgres", line)
	if err != nil {
		return nil, err
	}
	return db, err
}


func (app *Application) GetCategory() ([]Category, error) {
	line := "select * from category"
	rows, err := app.db.Query(line)
	if err != nil {
		return nil, err
	}
	var p_arr []Category
	defer rows.Close()
	for rows.Next() {
		var value Category
		var url_null, name_en_null sql.NullString
		err := rows.Scan(&value.Id, &value.Name, &name_en_null, &url_null)
		if err != nil {
			log.Println("Error row")
			return nil, err
		}
		value.Image_url, value.Name_en = url_null.String, name_en_null.String
		p_arr = append(p_arr, value)
	}
	return p_arr, nil
}

func (app *Application) GetProducts(category string, order string) ([]Product, error) {
	line := "select * from product"
	if order != "" {
		line = line+" order by "+order
	}
	line = line+" where category_id = "+category
	rows, err := app.db.Query(line)
	if err != nil {
		return nil, err
	}
	var p_arr []Product
	defer rows.Close()
	for rows.Next() {
		var value Product
		var url_null, name_en_null sql.NullString
		err := rows.Scan(&value.Id, &value.Name, &name_en_null, &value.Category_id, &value.Amount, &url_null, &value.Meas_unit, &value.Price)
		if err != nil {
			log.Println("Error row")
			return nil, err
		}
		value.Image_url, value.Name_en = url_null.String, name_en_null.String
		p_arr = append(p_arr, value)
	}
	return p_arr, nil
}

func (app *Application) CreateOrder(ctx context.Context, ch chan int) {
	ctx, cancel := context.WithCancel(ctx)

	var arr []int
	ord, _ := ctx.Value("data").(Order)
	for _, s := range ord.OrderItems {
		arr = append(arr, s.ProductId)
		arr = append(arr, s.Weight)
	}
	prod, total, err := app.CalculateOrder(arr)
	if err != nil {
		log.Println(err)
		cancel()
	}

	tx, err := app.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		log.Println(err)
		cancel()
	}
	stmt, _ := tx.Prepare("select * from add_order($1,$2,$3,$4,$5,$6)")
	defer stmt.Close()
	//log.Printf("select * from add_order($1,$2,$3,808398908,email,'10.20.1999 10:00:00')",pq.Array(arr), ord.Name, ord.Name)
	_, err = stmt.ExecContext(ctx,pq.Array(arr),ord.Client.Name,ord.Client.Surname,ord.Client.PhoneNumber,ord.Client.Email,time.Now())
	if err != nil {
		log.Println(err)
		cancel()
	}

	res := make(chan int)
	arg := context.WithValue(ctx, "prod", prod)
	arg = context.WithValue(arg, "total", total)
	go app.Mail.SendMail(arg,res)
	
	select {
	case <- ctx.Done():
		tx.Rollback()
		return
	case <- res:
		err = tx.Commit()
		ch <- 1
		return
	}
}

func (app *Application) CalculateOrder(prod []int) ([]Product, int, error) {
	var ids, wgh []int
	for i,s := range prod {
		if i%2 != 0 {
			wgh = append(wgh, s)
		} else {
			ids = append(ids, s)
		}
	}

	stmt, _ := app.db.Prepare("select * from product where product_id = ANY($1)")
	rows, err := stmt.Query(pq.Array(ids))

	if err != nil {
		return nil, 0, err
	}
	var p_arr []Product
	total := 0
	defer rows.Close()
	for i := 0; rows.Next(); i++ {
		var value Product
		var url_null, name_en_null sql.NullString
		err := rows.Scan(&value.Id, &value.Name, &name_en_null, &value.Category_id, &value.Amount, &url_null, &value.Meas_unit, &value.Price)
		if err != nil {
			log.Println("Error row")
			return nil, 0, err
		}
		value.Image_url, value.Name_en = url_null.String, name_en_null.String
		value.Amount = wgh[i]
		p_arr = append(p_arr, value)
		total += value.Price*wgh[i]
	}
	return p_arr, total, nil

}