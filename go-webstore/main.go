package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/a-h/templ"
	"github.com/go-sql-driver/mysql"
	"github.com/labstack/echo/v4"

	etag "github.com/pablor21/echo-etag/v4"

	"go-store/templates"
	"go-store/types"
	"go-store/db"
	"go-store/middleware"
)

type PurchaseData struct {
	FirstName         string
	LastName          string
	Email             string
	Product           string
	Price             float64
	Quantity          int
	Subtotal          float64
	TaxRate           float64
	TaxAmount         float64
	Total             float64
	Donation          bool
	TotalWithDonation float64
}

const TaxRate = 0.010

var conn *sql.DB

func main() {
	e := echo.New()

	dbCfg := mysql.Config{
		User:   "jnguyen1",
		Passwd: "Jdogjust2002!!",
		DBName: "jnguyen1",
	}

	var err error
	conn, err = sql.Open("mysql", dbCfg.FormatDSN())
	if err != nil {
		e.Logger.Fatal(err)
	}
	defer conn.Close()

	e.Use(etag.Etag())

	e.Static("/assets", "./assets")

	e.GET("/store", func(ctx echo.Context) error {
		products := db.GetProducts(conn)
		return Render(ctx, http.StatusOK, templates.Base(products))
	})

	e.POST("/purchase", func(ctx echo.Context) (err error) {
		FirstName := ctx.FormValue("fname")
		LastName := ctx.FormValue("lname")
		Email := ctx.FormValue("email")
		Product := ctx.FormValue("product")
		Quantity := ctx.FormValue("quantity")
		Donation := ctx.FormValue("donation")

		quantity, err := strconv.Atoi(Quantity)
		if err != nil || quantity < 1 {
			return ctx.String(http.StatusBadRequest, "Invalid quantity")
		}

		productData, err := db.GetProductByName(conn, Product)
		if err != nil {
			return ctx.String(http.StatusBadRequest, "Invalid product hellooooo")
		}

		price := productData.Price
		subtotal := float64(quantity) * price
		tax := subtotal * TaxRate
		total := subtotal + tax

		donation := 0.99

		if Donation == "yes" {
			total = float64(int(total + donation))
		}

		customerID, err := db.AddCustomer(conn, FirstName, LastName, Email)
		if err != nil {
			return ctx.String(http.StatusBadRequest, "Cant add customer")
		}

		db.AddOrder(conn, productData.ID, int(customerID), quantity, price, tax, donation)

		db.SellProduct(conn, productData.ID, quantity)

		confirmationHTML := fmt.Sprintf(`
			<!DOCTYPE html>
			<html>
			<head>
				<link rel="stylesheet" href="/assets/styles/styles.css">
			</head>
			<body>
				<h1>Order Confirmation</h1>
				<p>Thank you for your purchase, %s %s!</p>
				<p>Email: %s</p>
				<p>Product: %s</p>
				<p>Quantity: %d</p>
				<p>Price per item: $%.2f</p>
				<p>Subtotal: $%.2f</p>
				<p>Tax (%.2f%%): $%.2f</p>
				<p>Total: $%.2f</p>
				<a href="/store">Return to store</a>
			</body>
			</html>
		`, FirstName, LastName, Email, Product, quantity, price, subtotal, TaxRate*100, tax, total)

		return RenderWithHeaderFooter(ctx, confirmationHTML)
	})

	e.GET("/dbQueries", func(ctx echo.Context) error {
		var customerResults types.CustomerResults
		var productResults types.ProductResults
		var orderResults types.OrderResults
		customerResults.Customers = db.GetCustomers(conn)
		customer, _ := db.GetCustomerById(conn, 2)
		customerResults.Customer2 = customer
		_, err = db.GetCustomerById(conn, 3)
		if err != nil {
			customerResults.Customer3 = "Customer 3 not found!"
		}
		customer, _ = db.GetCustomerByEmail(conn, "mickeymouse@mines.edu")
		customerResults.Customer4 = customer
		_, err = db.GetCustomerByEmail(conn, "notfound@notanemail.io")
		if err != nil {
			customerResults.Customer5Find = "Customer notfound@notanemail.io not found ... adding ..."
		}
		db.AddCustomer(conn, "Not", "Found", "notfound@notanemail.io")
		customer, _ = db.GetCustomerByEmail(conn, "notfound@notanemail.io")
		customerResults.Customer5 = customer

		productResults.Product = db.GetProducts(conn)
		product, _ := db.SellProduct(conn, 2, 1)
		productResults.Product2 = product
		product2, _ := db.SellProduct(conn, 1, 5)
		productResults.Product3 = product2

		return Render(ctx, http.StatusOK, templates.DbQueries(customerResults, productResults, orderResults))
	})

	e.GET("/Admin", func(ctx echo.Context) error {
		customers := db.GetCustomers(conn)
		orders := db.GetOrders(conn)
		products := db.GetProducts(conn)
		return Render(ctx, http.StatusOK, templates.Admin(customers, orders, products))
	}, middleware.AuthMiddleware(2))

	e.GET("/OrderEntry", func(ctx echo.Context) error {
		products := db.GetProducts(conn)
		return Render(ctx, http.StatusOK, templates.OrderEntry(products))
	}, middleware.AuthMiddleware(1))

	e.GET("/get_product_quantity", func(ctx echo.Context) error {
		productName := ctx.QueryParam("product")
		product, err := db.GetProductByName(conn, productName)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
				"error": "Product not found",
			})
		}
		return ctx.JSON(http.StatusOK, map[string]interface{}{
			"quantity": product.InStock,
		})
	})

	e.GET("/get_customers", func(ctx echo.Context) error {
		search := ctx.QueryParam("search")
		customers, err := db.GetCustomersByNameFragment(conn, search)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch customers"})
		}

		customerTable := `<table id="customerTable"><thead><tr><th>First Name</th><th>Last Name</th><th>Email</th></tr></thead><tbody>`
		for _, customer := range customers {
			customerTable += fmt.Sprintf(
				`<tr data-id="%d"><td>%s</td><td>%s</td><td>%s</td></tr>`,
				customer.ID,
				customer.FirstName,
				customer.LastName,
				customer.Email,
			)
		}
		customerTable += "</tbody></table>"

		return ctx.HTML(http.StatusOK, customerTable)
	})

	e.GET("/process_purchase", func(ctx echo.Context) error {
		FirstName := ctx.QueryParam("fname")
		LastName := ctx.QueryParam("lname")
		Email := ctx.QueryParam("email")
		Product := ctx.QueryParam("product")
		Quantity := ctx.QueryParam("quantity")
		Donation := ctx.QueryParam("donation")

		quantity, err := strconv.Atoi(Quantity)
		if err != nil || quantity < 1 {
			return ctx.String(http.StatusBadRequest, "Invalid quantity")
		}

		productData, err := db.GetProductByName(conn, Product)
		if err != nil {
			return ctx.String(http.StatusBadRequest, "Invalid product process_purchase")
		}

		price := productData.Price
		subtotal := float64(quantity) * price
		tax := subtotal * TaxRate
		total := subtotal + tax

		donation := 0.00
		if Donation == "yes" {
			total += donation
		}

		customerID, err := db.AddCustomer(conn, FirstName, LastName, Email)
		if err != nil {
			return ctx.String(http.StatusBadRequest, "Can't add customer")
		}

		db.AddOrder(conn, productData.ID, int(customerID), quantity, price, tax, donation)
		db.SellProduct(conn, productData.ID, quantity)

		confirmation := map[string]interface{}{
			"firstName": FirstName,
			"lastName":  LastName,
			"quantity":  quantity,
			"product":   Product,
			"total":     total,
		}

		return ctx.JSON(http.StatusOK, confirmation)
	})

	e.GET("/Products", func(ctx echo.Context) error {
		var productResults types.ProductResults
		productResults.Product = db.GetProducts(conn)
		return Render(ctx, http.StatusOK, templates.Products(productResults))
	}, middleware.AuthMiddleware(1))

	e.POST("/addProduct", func(ctx echo.Context) error {
		var product struct {
			ProductName      string  `json:"pname"`
			ProductImageName string  `json:"pimage"`
			Quantity         int     `json:"quantity"`
			Price            float64 `json:"price"`
			Active           int     `json:"status"`
		}

		fmt.Println("Product Name:", product.ProductName)
		fmt.Println("Product Image Name:", product.ProductImageName)
		fmt.Println("Quantity:", product.Quantity)
		fmt.Println("Price:", product.Price)
		fmt.Println("Active Status:", product.Active)

		if err := ctx.Bind(&product); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request data"})
		}

		db.AddProduct(conn, product.ProductName, product.ProductImageName, product.Price, product.Quantity, product.Active)
		return ctx.JSON(http.StatusOK, map[string]string{"message": "Product added successfully"})
	})

	e.POST("/updateProduct", func(ctx echo.Context) error {
		var product struct {
			ProductName      string  `json:"pname"`
			ProductImageName string  `json:"pimage"`
			Quantity         int     `json:"quantity"`
			Price            float64 `json:"price"`
			Active           int     `json:"status"`
		}

		if err := ctx.Bind(&product); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request data"})
		}

		// ProductID, _ := db.GetProductIDByName(conn, product.ProductName)
		fmt.Println("Product Name:", product.ProductName)
		fmt.Println("Product Image Name:", product.ProductImageName)
		fmt.Println("Quantity:", product.Quantity)
		fmt.Println("Price:", product.Price)
		fmt.Println("Active Status:", product.Active)

		db.UpdateProduct(conn, product.ProductName, product.ProductImageName, product.Price, product.Quantity, product.Active)
		return ctx.JSON(http.StatusOK, map[string]string{"message": "Product updated successfully"})
	})

	// e.POST("/updateProduct", func(ctx echo.Context) error {
	// 	ProductName := ctx.FormValue("pname")
	// 	ProductImageName := ctx.FormValue("pimage")
	// 	Quantity := ctx.FormValue("quantity")
	// 	Price := ctx.FormValue("price")
	// 	Active := ctx.FormValue("status")
	// 	quantity, err := strconv.Atoi(Quantity)

	// 	fmt.Println("Product Name:", ProductName)
	// 	fmt.Println("Product Image Name:", ProductImageName)
	// 	fmt.Println("Quantity:", quantity)

	// 	if err != nil {
	// 		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid quantity"})
	// 	}
	// 	price, err := strconv.ParseFloat(Price, 64)
	// 	if err != nil {
	// 		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid price"})
	// 	}
	// 	active := 0
	// 	if Active == "1" {
	// 		active = 1
	// 	}
	// 	fmt.Println("Product Name:", ProductName)
	// 	fmt.Println("Product Image Name:", ProductImageName)
	// 	fmt.Println("Quantity:", quantity)
	// 	fmt.Println("Price:", price)
	// 	fmt.Println("Active Status:", active)
	// 	db.UpdateProduct(conn, ProductName, ProductImageName, price, quantity, active)
	
	// 	// Return a success message
	// 	return ctx.JSON(http.StatusOK, map[string]string{"message": "Product updated successfully"})
	// })

	e.POST("/deleteProduct", func(ctx echo.Context) error {
		var product struct {
			ProductName      string  `json:"pname"`
			ProductImageName string  `json:"pimage"`
			Quantity         int     `json:"quantity"`
			Price            float64 `json:"price"`
			Active           int     `json:"status"`
		}
		if err := ctx.Bind(&product); err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request data"})
		}

		fmt.Println("Product Name:", product.ProductName)
		fmt.Println("Product Image Name:", product.ProductImageName)
		fmt.Println("Quantity:", product.Quantity)
		fmt.Println("Price:", product.Price)
		fmt.Println("Active Status:", product.Active)
		productID, err := db.GetProductIDByName(conn, product.ProductName)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find product ID"})
		}
		isInOrders, err := db.GetCheck(conn, productID)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find product in orders"})
		}
		if isInOrders {
			fmt.Println("Product is in orders")
		} else {
			db.DeleteProduct(conn, productID)
		}
	
		return ctx.JSON(http.StatusOK, map[string]string{"message": "Product deleted successfully"})
	})

	e.GET("/", func(ctx echo.Context) error {
		errorMessage := ctx.QueryParam("error")

		data := map[string]interface{}{
			"ErrorMessage": errorMessage, 
		}

		return Render(ctx, http.StatusOK, templates.Login(data))
	})

	e.POST("/login", func(ctx echo.Context) error {
		email := ctx.FormValue("email")
		password := ctx.FormValue("password")
		if email == "" || password == "" {
			return ctx.Redirect(http.StatusFound, "/?error=invalid user")
		}
		fmt.Println("Helllloooooo")
	
		if db.Authenticate(conn, email, password) {
			fullName, err := db.GetUserByEmail(conn, email)
			if err != nil {
				if err == sql.ErrNoRows {
					return ctx.JSON(http.StatusNotFound, map[string]interface{}{
						"success": false,
						"message": "User not found",
					})
				}
				return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
					"success": false,
					"message": "Error fetching user information",
				})
			}
			fmt.Println("Name:", fullName)
	
			cookie := new(http.Cookie)
			cookie.Name = "user_email"
			cookie.Value = email
			cookie.Path = "/"        
			cookie.HttpOnly = false   
			cookie.MaxAge = 3600     
			http.SetCookie(ctx.Response(), cookie)
	
			nameCookie := new(http.Cookie)
			nameCookie.Name = "user_full_name"
			nameCookie.Value = fullName
			nameCookie.Path = "/"
			nameCookie.HttpOnly = false
			nameCookie.MaxAge = 3600
			http.SetCookie(ctx.Response(), nameCookie)
	
			fmt.Println("Email Cookie:", cookie.Value)
			fmt.Println("Full Name Cookie:", nameCookie.Value)
	
			fmt.Println("authenticated")
	
			roleNumber, err := db.CheckRole(conn, email)
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to find role using email"})
			}
	
			if roleNumber == 1 {
				return ctx.Redirect(http.StatusFound, "/OrderEntry")
			} else if roleNumber == 2 {
				return ctx.Redirect(http.StatusFound, "/Products")
			}
	
			return ctx.Redirect(http.StatusFound, "/store")
		}
	
		fmt.Println("not authenticated")
		return Render(ctx, http.StatusOK, templates.Login(map[string]interface{}{"ErrorMessage": "Invalid email or password"}))
	})
	
	
	e.GET("/get-user-info", func(ctx echo.Context) error {
		email := ctx.QueryParam("email")
		if email == "" {
			return ctx.JSON(http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": "Email not provided",
			})
		}
	
		fullName, role, err := db.GetUserInfo(conn, email)
		fmt.Println("fullName: ", fullName, " role: ", role)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "Failed to fetch user info",
			})
		}
	
		return ctx.JSON(http.StatusOK, map[string]interface{}{
			"success":  true,
			"fullName": fullName,
			"role":     role,
		})
	})
	
	
	e.GET("/logout", func(ctx echo.Context) error {
		http.SetCookie(ctx.Response(), &http.Cookie{
			Name:   "user_email",
			Value:  "",
			Path:   "/",
			MaxAge: -1, 
		})
		http.SetCookie(ctx.Response(), &http.Cookie{
			Name:   "user_full_name",
			Value:  "",
			Path:   "/",
			MaxAge: -1, 
		})
	
		return ctx.Redirect(http.StatusFound, "/")
	})
	
	
	
   
	

	e.Logger.Fatal(e.Start(":8000"))

}

func Render(ctx echo.Context, statusCode int, t templ.Component) error {
	buf := templ.GetBuffer()
	defer templ.ReleaseBuffer(buf)

	if err := t.Render(ctx.Request().Context(), buf); err != nil {
		return err
	}

	return ctx.HTML(statusCode, buf.String())
}

func RenderWithHeaderFooter(ctx echo.Context, confirmationContent string) error {
	buf := templ.GetBuffer()
	defer templ.ReleaseBuffer(buf)

	if err := templates.Header().Render(ctx.Request().Context(), buf); err != nil {
		return err
	}

	buf.WriteString(confirmationContent)

	if err := templates.Footer().Render(ctx.Request().Context(), buf); err != nil {
		return err
	}

	return ctx.HTML(http.StatusOK, buf.String())
}
