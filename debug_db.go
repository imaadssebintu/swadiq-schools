package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	connStr := "host=129.80.199.242 port=5432 user=imaad password=Ertdfgxc dbname=swadiq sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	fmt.Println("Testing Expenses query...")
	query := `SELECT e.id, e.category_id, e.title, e.amount, e.currency, e.date, 
			  e.created_at, e.updated_at, c.id, c.name
			  FROM expenses e
			  LEFT JOIN categories c ON e.category_id = c.id
			  WHERE e.deleted_at IS NULL
			  ORDER BY e.date DESC`

	rows, err := db.Query(query)
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id, catID, title, currency string
		var amount interface{} // Use interface to see actual type
		var date, createdAt, updatedAt interface{}
		var cID, cName sql.NullString

		err := rows.Scan(&id, &catID, &title, &amount, &currency, &date, &createdAt, &updatedAt, &cID, &cName)
		if err != nil {
			fmt.Printf("Scan failed: %v\n", err)
			continue
		}
		fmt.Printf("Found: %s, Amount: %v, Type: %T\n", title, amount, amount)
	}
	fmt.Println("Test complete.")
}
