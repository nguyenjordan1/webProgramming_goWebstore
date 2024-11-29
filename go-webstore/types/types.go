package types

type CustomerResults struct {
	Customers  []Customer
	Customer2 Customer
	Customer3 string
	Customer4 Customer
	Customer5Find string
	Customer5 Customer
}

type Customer struct {
	ID        int
	FirstName string
	LastName  string
	Email     string
}

type ProductResults struct {
	Product []Product
	Product2 Product
	Product3 Product
}

type Product struct {
	ID int
    ProductName string
    InStock     int
    ImageName   string
	Price float64
	ImagePath string
	Inactive int
}

type OrderResults struct {
	Order []Order
	Order2 Order
}
type Order struct {
	ID              int
	ProductID       int
	CustomerID      int
	Quantity        int
	Price           float64
	Tax             float64
	Donation        float64
	Timestamp       int64
	CustomerFirstName string
	CustomerLastName  string
	ProductName     string
}


type PurchaseResponse struct {
    Message string `json:"message"`
}

