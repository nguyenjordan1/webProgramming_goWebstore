USE jnguyen1;

DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS customers;
DROP TABLE IF EXISTS product;

CREATE TABLE customers (
    id INT AUTO_INCREMENT PRIMARY KEY,
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    email VARCHAR(255)
);

CREATE TABLE product (
    id INT AUTO_INCREMENT PRIMARY KEY,
    product_name VARCHAR(255),
    image_name VARCHAR(255),
    price DECIMAL(6, 2),
    in_stock INT,
    inactive TINYINT DEFAULT 0 
);

CREATE TABLE orders (
    id INT AUTO_INCREMENT PRIMARY KEY,
    product_id INT,
    customer_id INT,
    quantity INT,
    price DECIMAL(6,2),
    tax DECIMAL(6,2),
    donation DECIMAL(4,2),
    timestamp BIGINT,
    FOREIGN KEY (product_id) REFERENCES product(id),
    FOREIGN KEY (customer_id) REFERENCES customers(id)
);

CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    first_name VARCHAR(50) NOT NULL,
    last_name VARCHAR(50) NOT NULL,
    password VARCHAR(255) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    role INT NOT NULL
);

INSERT INTO Users (first_name, last_name, password, email, role)
VALUES
    ('Frodo', 'Baggins', 'fb', 'fb@mines.edu', 1),
    ('Harry', 'Potter', 'hp', 'hp@mines.edu', 2);

INSERT INTO customers (first_name, last_name, email)
VALUES
    ('Mickey', 'Mouse', 'mickeymouse@mines.edu'),
    ('Minnie', 'Mouse', 'minniemouse@mines.edu');

INSERT INTO product (product_name, image_name, price, in_stock, inactive)
VALUES
    ('babyBottle', 'babyBottle.jpg', 5.00, 0, 1),  
    ('pacifier', 'pacifier.jpg', 30.00, 3, 0),     
    ('diapers', 'diapers.jpg', 10.00, 10, 0);
