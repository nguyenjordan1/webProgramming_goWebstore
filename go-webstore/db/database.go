package db

import (
	"database/sql"
	"fmt"
	"go-store/types"
	// "log"
	// _ "github.com/go-sql-driver/mysql"
)

func GetCustomers(db *sql.DB) []types.Customer {
	stmt, _ := db.Prepare("SELECT * FROM customers")
	defer stmt.Close()

	rows, _ := stmt.Query()

	var customers []types.Customer
	var customer types.Customer
	for rows.Next() {
		_ = rows.Scan(&customer.ID, &customer.FirstName, &customer.LastName, &customer.Email)
		customers = append(customers, customer)
	}

	return customers
}

func GetCustomerById(db *sql.DB, id int) (types.Customer, error) {
	stmt, _ := db.Prepare("SELECT * FROM customers WHERE id = ?")
	defer stmt.Close()
	var customer types.Customer
	err := stmt.QueryRow(id).Scan(&customer.ID, &customer.FirstName, &customer.LastName, &customer.Email)
	if err != nil {
		return customer, err
	}
	return customer, nil
}

func GetCustomerByEmail(db *sql.DB, email string) (types.Customer, error) {
	stmt, _ := db.Prepare("SELECT * FROM customers WHERE email = ?")
	defer stmt.Close()
	var customer types.Customer
	err := stmt.QueryRow(email).Scan(&customer.ID, &customer.FirstName, &customer.LastName, &customer.Email)
	if err != nil {
		return customer, err
	}
	return customer, nil
}

//	func AddCustomer(db *sql.DB, firstName, lastName, email string) (int, error) {
//		stmt, _ := db.Prepare("INSERT INTO customers(first_name, last_name, email) VALUES (?, ?, ?)")
//		defer stmt.Close()
//		_, err := stmt.Exec(firstName, lastName, email)
//		return err
//	}
func AddCustomer(db *sql.DB, firstName, lastName, email string) (int64, error) {
	stmt, err := db.Prepare("INSERT INTO customers(first_name, last_name, email) VALUES (?, ?, ?)")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	result, err := stmt.Exec(firstName, lastName, email)
	if err != nil {
		return 0, err
	}

	customerID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return customerID, nil
}

func GetProducts(db *sql.DB) []types.Product {
	stmt, _ := db.Prepare("SELECT id, product_name, image_name, price, in_stock FROM product")
	defer stmt.Close()

	rows, _ := stmt.Query()

	var products []types.Product
	var product types.Product
	for rows.Next() {
		_ = rows.Scan(&product.ID, &product.ProductName, &product.ImageName, &product.Price, &product.InStock)
		products = append(products, product)
	}

	return products
}

func GetProductByName(db *sql.DB, name string) (types.Product, error) {
	stmt, err := db.Prepare("SELECT id, product_name, price, in_stock FROM product WHERE product_name = ?")
	if err != nil {
		return types.Product{}, fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer stmt.Close()

	var product types.Product
	err = stmt.QueryRow(name).Scan(&product.ID, &product.ProductName, &product.Price, &product.InStock)
	if err != nil {
		return product, err
	}
	return product, nil
}

func GetProductInstock(db *sql.DB, productName string) (types.Product, error) {
	stmt, err := db.Prepare("SELECT InStock WHERE product_name = ?")
	if err != nil {
		return types.Product{}, fmt.Errorf("failed to prepare statement: %v", err)
	}

	defer stmt.Close()
	var inStock int
	err = stmt.QueryRow(productName).Scan(&inStock)
	if err != nil {
		if err == sql.ErrNoRows {
			return types.Product{}, fmt.Errorf("no product found with name %s", productName)
		}
		return types.Product{}, fmt.Errorf("failed to execute query: %v", err)
	}
	product := types.Product{
		ProductName: productName,
		InStock:     inStock,
	}

	return product, nil

}

func SellProduct(db *sql.DB, productID int, quantity int) (types.Product, error) {
	stmt, err := db.Prepare("UPDATE product SET in_stock = in_stock - ? WHERE in_stock >= ? AND id = ?")
	if err != nil {
		return types.Product{}, fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer stmt.Close()
	result, err := stmt.Exec(quantity, quantity, productID)
	if err != nil {
		return types.Product{}, fmt.Errorf("failed to execute statement: %v", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return types.Product{}, fmt.Errorf("failed to get rows affected: %v", err)
	}
	if rowsAffected == 0 {
		return types.Product{}, fmt.Errorf("no product sold: either product not found or insufficient stock")
	}
	var product types.Product
	err = db.QueryRow("SELECT id, product_name, price, in_stock FROM product WHERE id = ?", productID).
		Scan(&product.ID, &product.ProductName, &product.Price, &product.InStock)
	if err != nil {
		return types.Product{}, fmt.Errorf("failed to retrieve product: %v", err)
	}

	return product, nil
}
func GetOrders(db *sql.DB) []types.Order {
	stmt, _ := db.Prepare(`
		SELECT 
			o.id, 
			o.quantity, 
			o.price, 
			o.tax, 
			o.donation, 
			o.timestamp, 
			c.first_name, 
			c.last_name, 
			p.product_name
		FROM 
			orders o
		JOIN 
			customers c ON o.customer_id = c.id
		JOIN 
			product p ON o.product_id = p.id
	`)
	defer stmt.Close()

	rows, _ := stmt.Query()

	var orders []types.Order
	for rows.Next() {
		var order types.Order
		_ = rows.Scan(&order.ID, &order.Quantity, &order.Price, &order.Tax, &order.Donation, &order.Timestamp, &order.CustomerFirstName, &order.CustomerLastName, &order.ProductName)
		orders = append(orders, order)
	}

	return orders
}

func AddOrder(db *sql.DB, productID, customerID int, quantity int, price, tax, donation float64) error {
	stmt, err := db.Prepare("INSERT INTO orders (product_id, customer_id, quantity, price, tax, donation, timestamp) VALUES (?, ?, ?, ?, ?, ?, UNIX_TIMESTAMP())")
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()
	_, err = stmt.Exec(productID, customerID, quantity, price, tax, donation)
	if err != nil {
		return fmt.Errorf("failed to execute statement: %w", err)
	}

	return nil
}

func GetCustomersByNameFragment(conn *sql.DB, search string) ([]types.Customer, error) {
	var customers []types.Customer
	rows, err := conn.Query("SELECT id, first_name, last_name, email FROM customers WHERE first_name LIKE ? OR last_name LIKE ?", "%"+search+"%", "%"+search+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var customer types.Customer
		if err := rows.Scan(&customer.ID, &customer.FirstName, &customer.LastName, &customer.Email); err != nil {
			return nil, err
		}
		customers = append(customers, customer)
	}
	return customers, nil
}

func AddProduct(db *sql.DB, ProductName string, ProductImageName string, price float64, quantity int, active int) (int64, error) {
    stmt, err := db.Prepare("INSERT INTO product(product_name, image_name, price, in_stock, inactive) VALUES (?, ?, ?, ?, ?)")
    if err != nil {
        return 0, err
    }
    defer stmt.Close()
    result, err := stmt.Exec(ProductName, ProductImageName, price, quantity, active)
    if err != nil {
        return 0, err
    }

    productID, err := result.LastInsertId()
    if err != nil {
        return 0, err
    }

    return productID, nil
}
func GetCheck(db *sql.DB, productID int) (bool, error) {
    fmt.Println("is in check")
    row := db.QueryRow("SELECT id FROM orders WHERE product_id = ?", productID)

    var orderID int
    err := row.Scan(&orderID)
    if err == sql.ErrNoRows {
        fmt.Println("No rows found for product_id:", productID)
        return false, nil
    }
    if err != nil {
        fmt.Println("Error scanning row:", err)
        return false, err
    }
    return true, nil
}

func GetProductIDByName(db *sql.DB, productName string) (int, error) {
	stmt, err := db.Prepare("SELECT id FROM product WHERE product_name = ?")
    if err != nil {
        return 0, err
    }
    defer stmt.Close()  
    var productID int
    err = stmt.QueryRow(productName).Scan(&productID)
    if err != nil {
        if err == sql.ErrNoRows {
            return 0, nil
        }
        return 0, err
    }
	fmt.Println("productID: ", productID)
    return productID, nil
}



func UpdateProduct(db *sql.DB, ProductName string, ProductImageName string, price float64, quantity int, active int) error {
    fmt.Println("gets here")
    query := "UPDATE product SET image_name = ?, price = ?, in_stock = ?, inactive = ? WHERE product_name = ?"
    _, err := db.Exec(query, ProductImageName, price, quantity, active, ProductName)
    
    if err != nil {
        fmt.Println("exit 1")
        return fmt.Errorf("failed to execute update: %w", err)
    }

    fmt.Println("update successful")
    return nil
}



func DeleteProduct(db *sql.DB, productID int) error {
	_, err := db.Exec("DELETE FROM product WHERE id = ?", productID)
	if err != nil {
		return err
	}
	return nil
}


func Authenticate(db *sql.DB, email string, password string) bool {
	
	stmt, err := db.Prepare("SELECT * FROM users WHERE email = ? AND password = ?")
	if err != nil {
		return false
	}
	defer stmt.Close()

	var id int
	var firstName, lastName, dbPassword, dbEmail string
	var role int

	err = stmt.QueryRow(email, password).Scan(&id, &firstName, &lastName, &dbPassword, &dbEmail, &role)
	if err != nil {
		if err == sql.ErrNoRows {
			return false
		}
		return false 
	}

	return true
}
func CheckRole(db *sql.DB, email string) (int, error) {
	stmt, err := db.Prepare("SELECT role FROM users WHERE email = ?")
	if err != nil {
		return 0, err 
	}
	defer stmt.Close()

	var role int

	err = stmt.QueryRow(email).Scan(&role)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("no user found with email: %s", email)
		}
		return 0, err 
	}

	fmt.Println("The Role number: ", role)
	return role, nil
}

func GetUserByEmail(db *sql.DB, email string) (string, error) {
    stmt, err := db.Prepare("SELECT first_name, last_name FROM users WHERE email = ?")
    if err != nil {
        return "", fmt.Errorf("failed to prepare statement: %v", err)
    }
    defer stmt.Close()
    var firstName, lastName string

    err = stmt.QueryRow(email).Scan(&firstName, &lastName)
    if err != nil {
        if err == sql.ErrNoRows {
            return "", fmt.Errorf("no user found with email: %s", email)
        }
        return "", fmt.Errorf("query failed: %v", err)
    }

    fullName := fmt.Sprintf("%s %s", firstName, lastName)
    return fullName, nil
}
func GetUserInfo(db *sql.DB, email string) (string, int, error) {
	var firstName, lastName string
	var role int

	fmt.Println("enters GetUserInfo")
	query := `SELECT first_name, last_name, role FROM users WHERE email = ?`
	err := db.QueryRow(query, email).Scan(&firstName, &lastName, &role)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", 0, fmt.Errorf("no user found with email: %s", email)
		}
		return "", 0, fmt.Errorf("query failed: %v", err)
	}

	fullName := fmt.Sprintf("%s %s", firstName, lastName)
	return fullName, role, nil
}



