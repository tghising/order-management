package main

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
	"strconv"
)

const (
	// configuration for postgres database
	DATABASE_USER     = "postgres"
	DATABASE_PASSWORD = "root"
	DATABASE_NAME     = "order-system"
	DB_HOST           = "localhost"
	DB_PORT           = 5432

	create_customer_companies = "" +
		"CREATE TABLE IF NOT EXISTS customer_companies (" +
		"company_id INT PRIMARY KEY NOT NULL," +
		"company_name VARCHAR(60) NOT NULL" +
		")"

	create_customers string = "" +
		"CREATE TABLE IF NOT EXISTS customers ( " +
		"user_id VARCHAR(30) PRIMARY KEY NOT NULL," +
		"login VARCHAR(50) UNIQUE NOT NULL," +
		"password VARCHAR(100) NOT NULL," +
		"name VARCHAR(100) NOT NULL," +
		"company_id INT," +
		"credit_cards TEXT," +
		"FOREIGN KEY (company_id) REFERENCES customer_companies(company_id)" +
		")"

	create_orders string = "" +
		"CREATE TABLE IF NOT EXISTS orders (" +
		"id INT PRIMARY KEY NOT NULL," +
		"created_at timestamp with time zone NOT NULL," +
		"order_name VARCHAR(100) NOT NULL," +
		"customer_id VARCHAR(30) NOT NULL," +
		"FOREIGN KEY (customer_id) REFERENCES customers(user_id)" +
		")"

	create_order_items string = "" +
		"CREATE TABLE IF NOT EXISTS order_items (" +
		"id INT PRIMARY KEY NOT NULL," +
		"order_id INT NOT NULL," +
		"price_per_unit DECIMAL DEFAULT NULL," +
		"quantity INT NOT NULL," +
		"product VARCHAR(50) NOT NULL," +
		"FOREIGN KEY (order_id) REFERENCES orders(id)" +
		")"

	create_deliveries string = "" +
		"CREATE TABLE IF NOT EXISTS deliveries (" +
		"id INT PRIMARY KEY NOT NULL," +
		"order_item_id INT NOT NULL," +
		"delivered_quantity INT NOT NULL," +
		"FOREIGN KEY (order_item_id) REFERENCES order_items(id)" +
		")"

	insert_customer_companies string = "INSERT INTO customer_companies(company_id, company_name) VALUES($1,$2)"

	insert_customers string = "" +
		"INSERT INTO customers(user_id, login, password, name, company_id, credit_cards)" +
		"VALUES($1, $2, $3, $4, $5, $6)"

	insert_orders string = "" +
		"INSERT INTO orders(id, created_at, order_name, customer_id)" +
		"VALUES($1, $2, $3, $4)"

	insert_order_items string = "" +
		"INSERT INTO order_items(id, order_id, price_per_unit, quantity, product)" +
		"VALUES($1, $2, $3, $4, $5)"

	insert_deliveries string = "" +
		"INSERT INTO deliveries(id, order_item_id, delivered_quantity)" +
		"VALUES($1, $2, $3)"

	searchByOrderOrProduct string = "" +
		"SELECT o.order_name, cc.company_name as customer_company,cc.name as customer_name, o.created_at as order_date," +
		"od.delivered_quantity*od.price_per_unit as delivered_amount, od.product " +
		"from " +
		"(select comp.company_name, c.name, c.user_id  from customer_companies comp, customers c where comp.company_id = c.company_id) cc," +
		"Orders o, " +
		"(Select oi.order_id, oi.price_per_unit,oi.product, d.delivered_quantity from order_items oi join deliveries d on oi.id = d.order_item_id) od " +
		"where cc.user_id = o.customer_id AND o.id = od.order_id AND (od.product like '%$1%' OR o.order_name like '%$2%')"
)

type Customer_Orders struct {
	OrderName       string  `json:"order_name"`
	CustomerCompany string  `json:"customer_company"`
	CustomerName    string  `json:"customer_name"`
	OrderDate       string  `json:"order_date"`
	DeliveredAmount float64 `json:"delivered_amount"`
	TotalAmount     float64 `json:"total_amount"`
	Product         string  `json:"product_name"`
}

type JsonOrdersResponse struct {
	Status           string            `json:"status"`
	Data             []Customer_Orders `json:"data"`
	Page             int               `json:"page"`
	PageSize         int               `json:"pageSize"`
	TotalElements    int               `json:"totalElements"`
	TotalPages       int               `json:"totalPages"`
	GrandTotalAmount float64           `json:"grand_total_amount"`
}

type DefaultResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func main() {
	// initialize the db tables
	initializeDB()

	// read csv data and write into POSTGRESQL
	readCSVData()

	//handling REST API
	handleRequests()
}

func homePage(rw http.ResponseWriter, req *http.Request) {
	printMessage("Endpoint Hit: getOrders == > " + req.Host + req.URL.String())
	json.NewEncoder(rw).Encode(DefaultResponse{Status: "Success", Message: "Welcome to the home page"})
}

func getOrders(rw http.ResponseWriter, req *http.Request) {
	printMessage("Endpoint Hit: getOrders == > " + req.Host + req.URL.String())
	rw.Header().Set("Access-Control-Allow-Origin", "*")

	totalElements := 0
	grandTotalAmount := 0.0
	totalPages := 0
	var order_list []Customer_Orders

	page, _ := strconv.Atoi(req.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(req.URL.Query().Get("pageSize"))
	orderNameOrProduct := req.URL.Query().Get("orderNameOrProduct")

	start := (page - 1) * pageSize
	if page < 1 || pageSize < 1 {
		json.NewEncoder(rw).Encode(DefaultResponse{Status: "Failed", Message: "Request without page and pageSize is invalid."})
		return
	}

	db := dbConnect() // database connection for query orders query
	if orderNameOrProduct != "" || len(orderNameOrProduct) > 0 {
		searchTotalOrders, err := db.Query("SELECT COUNT(*) as total_count, SUM(total_amount) as grand_total_amount FROM (SELECT o.order_name, cc.company_name as customer_company,cc.name as customer_name, o.created_at as order_date,od.delivered_quantity*od.price_per_unit as delivered_amount, od.quantity*od.price_per_unit as total_amount, od.product from (select comp.company_name, c.name, c.user_id  from customer_companies comp, customers c where comp.company_id = c.company_id) cc, Orders o, (Select oi.order_id, oi.quantity, oi.price_per_unit,oi.product, d.delivered_quantity from order_items oi join deliveries d on oi.id = d.order_item_id) od where cc.user_id = o.customer_id AND o.id = od.order_id AND (od.product LIKE '%' || $1 || '%' OR o.order_name LIKE '%' || $2 || '%')) RS ", orderNameOrProduct, orderNameOrProduct)
		for searchTotalOrders.Next() {
			var total_count int
			var grand_total_amount float64
			err = searchTotalOrders.Scan(&total_count, &grand_total_amount)
			checkErr(err)
			totalElements = total_count
			grandTotalAmount = grand_total_amount
		}

		orders, err := db.Query("SELECT o.order_name, cc.company_name as customer_company,cc.name as customer_name, o.created_at as order_date,od.delivered_quantity*od.price_per_unit as delivered_amount, od.quantity*od.price_per_unit as total_amount, od.product from (select comp.company_name, c.name, c.user_id  from customer_companies comp, customers c where comp.company_id = c.company_id) cc, Orders o, (Select oi.order_id, oi.quantity, oi.price_per_unit,oi.product, d.delivered_quantity from order_items oi join deliveries d on oi.id = d.order_item_id) od where cc.user_id = o.customer_id AND o.id = od.order_id AND (od.product LIKE '%' || $1 || '%' OR o.order_name LIKE '%' || $2 || '%') LIMIT $3 OFFSET $4", orderNameOrProduct, orderNameOrProduct, pageSize, start)
		checkErr(err)
		sum := 0.0

		for orders.Next() {
			var order_name string
			var customer_company string
			var customer_name string
			var order_date string
			var delivered_amount float64
			var total_amount float64
			var product string

			err = orders.Scan(&order_name, &customer_company, &customer_name, &order_date, &delivered_amount, &total_amount, &product)
			checkErr(err)
			sum += total_amount
			order_list = append(order_list, Customer_Orders{OrderName: order_name + "\n" + product, CustomerCompany: customer_company,
				CustomerName: customer_name, OrderDate: order_date, DeliveredAmount: delivered_amount, TotalAmount: total_amount, Product: product})
		}

	} else {
		ordersTotal, err := db.Query("SELECT COUNT(*) as total_count, SUM(total_amount) as grand_total_amount FROM (SELECT oid.order_name, cc.company_name as customer_company,cc.name as customer_name, oid.created_at as order_date,oid.delivered_quantity*oid.price_per_unit as delivered_amount,oid.quantity*oid.price_per_unit as total_amount, oid.product FROM (SELECT o.*, od.* from orders o join (select oi.order_id, oi.price_per_unit, oi.quantity,oi.product, d.delivered_quantity from order_items oi join deliveries d on oi.id = d.order_item_id) od on od.order_id = o.id) oid join (select comp.company_name, c.name, c.user_id  from customer_companies comp, customers c where comp.company_id = c.company_id) cc on oid.customer_id = cc.user_id) RS")
		checkErr(err)
		for ordersTotal.Next() {
			var total_count int
			var grand_total_amount float64
			err = ordersTotal.Scan(&total_count, &grand_total_amount)
			checkErr(err)
			totalElements = total_count
			grandTotalAmount = grand_total_amount
		}

		orders, err := db.Query("SELECT oid.order_name, cc.company_name as customer_company,cc.name as customer_name, oid.created_at as order_date,oid.delivered_quantity*oid.price_per_unit as delivered_amount,oid.quantity*oid.price_per_unit as total, oid.product FROM (SELECT o.*, od.* from orders o join (select oi.order_id, oi.price_per_unit, oi.quantity,oi.product, d.delivered_quantity from order_items oi join deliveries d on oi.id = d.order_item_id) od on od.order_id = o.id) oid join (select comp.company_name, c.name, c.user_id  from customer_companies comp, customers c where comp.company_id = c.company_id) cc on oid.customer_id = cc.user_id LIMIT $1 OFFSET $2", pageSize, start)
		checkErr(err)
		for orders.Next() {
			var order_name string
			var customer_company string
			var customer_name string
			var order_date string
			var delivered_amount float64
			var total_amount float64
			var product string

			err = orders.Scan(&order_name, &customer_company, &customer_name, &order_date, &delivered_amount, &total_amount, &product)
			checkErr(err)

			order_list = append(order_list, Customer_Orders{OrderName: order_name + "\n" + product, CustomerCompany: customer_company,
				CustomerName: customer_name, OrderDate: order_date, DeliveredAmount: delivered_amount, TotalAmount: total_amount, Product: product})
		}
	}
	db.Close()

	if totalElements%pageSize == 0 {
		totalPages = totalElements / pageSize
	} else {
		totalPages = totalElements/pageSize + 1
	}

	if len(order_list) > 0 {
		var response = JsonOrdersResponse{Status: "Success", Data: order_list,
			Page: page, PageSize: pageSize, TotalElements: totalElements, TotalPages: totalPages, GrandTotalAmount: grandTotalAmount}
		json.NewEncoder(rw).Encode(response)
	} else {
		var response = JsonOrdersResponse{Status: "Failed", Data: order_list,
			Page: page, PageSize: pageSize, TotalElements: 0, TotalPages: 0, GrandTotalAmount: 0.0}
		json.NewEncoder(rw).Encode(response)
	}
}

func handleRequests() {
	// Init the mux router
	router := mux.NewRouter()

	// Route handles & endpoints

	// get to home page records
	router.HandleFunc("/", homePage).Methods("GET")

	//get all orders records
	router.HandleFunc("/api/orders", getOrders).Methods("GET")

	// serve http requests
	log.Fatal(http.ListenAndServe(":8090", router))
	printMessage("The server has been started at port : 8090")
}

func initializeDB() {
	db := dbConnect()                  // db connection
	db.Exec(create_customer_companies) // customer_company table creation
	db.Exec(create_customers)          // customers table creation
	db.Exec(create_orders)             // orders table creation
	db.Exec(create_order_items)        // order_items table creation
	db.Exec(create_deliveries)         // deliveries table creation
	printMessage("The tables has been created in database.")
	db.Close()
}

func readCSVData() {
	db := dbConnect() // db connection

	// read & write to db customer_companies.csv records
	records, err := readData("data/Test task - Mongo - customer_companies.csv")
	checkErr(err)
	for _, record := range records {
		db.Exec(insert_customer_companies, record[0], record[1])
	}

	// read & write to db customer.csv records
	customers, err := readData("data/Test task - Mongo - customers.csv")
	checkErr(err)
	for _, customer := range customers {
		db.Exec(insert_customers, customer[0], customer[1], customer[2], customer[3],
			customer[4], customer[5])
	}
	printMessage("Customers records has been read from csv and inserted into customers table in database.")

	// read & write to db orders.csv records
	orders, err := readData("data/Test task - Postgres - orders.csv")
	checkErr(err)
	for _, order := range orders {
		db.Exec(insert_orders, order[0], order[1], order[2], order[3])
	}
	printMessage("Orders records has been read from csv and inserted into orders table in database.")

	// read & write to db order_items.csv records
	order_items, err := readData("data/Test task - Postgres - order_items.csv")
	checkErr(err)
	for _, o_item := range order_items {
		db.Exec(insert_order_items, o_item[0], o_item[1], o_item[2],
			o_item[3], o_item[4])
	}
	printMessage("Order_Items records has been read from csv and inserted into order_items table in database.")

	// read & write to db deliveries.csv records
	deliveries, err := readData("data/Test task - Postgres - deliveries.csv")
	checkErr(err)
	for _, delivery := range deliveries {
		db.Exec(insert_deliveries, delivery[0], delivery[1], delivery[2])
	}

	db.Close()
	printMessage("Deliveries records has been read from csv and inserted into deliveries table in database.")
}

func readData(fileName string) ([][]string, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return [][]string{}, err
	}
	defer f.Close()
	req := csv.NewReader(f)
	// skip first line
	if _, err := req.Read(); err != nil {
		return [][]string{}, err
	}
	records, err := req.ReadAll()
	if err != nil {
		return [][]string{}, err
	}
	return records, nil
}

// DB connect
func dbConnect() *sql.DB {
	dbConInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		DB_HOST, DB_PORT, DATABASE_USER, DATABASE_PASSWORD, DATABASE_NAME)
	dbConnect, err := sql.Open("postgres", dbConInfo) // connect with given database
	checkErr(err)
	return dbConnect
}

// Function for handling errors
func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

// Function for handling messages
func printMessage(message string) {
	fmt.Println("")
	fmt.Println(message)
}
