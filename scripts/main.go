package scripts

import (
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func RunMigrations() {
	if len(os.Args) < 2 {
		println("Usage: go run main.go [normal|alert|delete]")
		return
	}

	dsn := "postgres://optimizer_core_local:@localhost:5432/optimizer_core_live?sslmode=disable"

	switch os.Args[1] {
	case "normal":
		fmt.Println("Inserting normal payments")
		InsertNormalPayments(dsn)
	case "alert":
		fmt.Println("Inserting alert payments")
		InsertAlertPayments(dsn)
	case "delete":
		fmt.Println("Deleting all payments")
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			fmt.Printf("Failed to connect to database: %v\n", err)
			return
		}
		result := db.Exec("DELETE FROM payments")
		if result.Error != nil {
			fmt.Printf("Failed to delete payments: %v\n", result.Error)
			return
		}
		fmt.Printf("Successfully deleted %d payments\n", result.RowsAffected)
	default:
		println("Invalid argument. Use 'normal', 'alert', or 'delete'")
	}
}
