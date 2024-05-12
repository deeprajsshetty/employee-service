package empdb_test

import (
	"database/sql"
	"log"
	"os"
	"testing"
	"time"

	"github.com/employees/empdb" // Replace with the actual package path
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	// Setup test database
	testDB, _ = sql.Open("sqlite3", ":memory:")
	empdb.CreateEmployeeTable(testDB)

	// Run tests
	code := m.Run()

	// Clean up
	testDB.Close()
	os.Exit(code)
}

func TestEmpCRUDOperations(t *testing.T) {
	// Create
	emp := empdb.Employee{Name: "John Doe", Position: "Manager", Salary: 50000.0, DateCreated: &time.Time{}}
	empID, err := empdb.CreateEmmployee(testDB, emp)
	if err != nil {
		t.Errorf("Failed to create employee: %v", err)
	}

	// Read
	createdEmp, err := empdb.GetEmployee(testDB, empID)
	if err != nil {
		t.Errorf("Failed to get employee: %v", err)
	}
	if createdEmp.Name != emp.Name || createdEmp.Position != emp.Position || createdEmp.Salary != emp.Salary {
		t.Errorf("Retrieved employee does not match created employee")
	}

	// Update
	createdEmp.Name = "Jane Doe"
	err = empdb.UpdateEmployee(testDB, *createdEmp)
	if err != nil {
		t.Errorf("Failed to update employee: %v", err)
	}

	// Verify update
	updatedEmp, err := empdb.GetEmployee(testDB, empID)
	if err != nil {
		t.Errorf("Failed to get employee after update: %v", err)
	}
	if updatedEmp.Name != "Jane Doe" {
		t.Errorf("Employee name not updated")
	}

	// Delete
	err = empdb.DeleteEmployee(testDB, empID)
	if err != nil {
		t.Errorf("Failed to delete employee: %v", err)
	}

	// Verify delete
	deletedEmp, err := empdb.GetEmployee(testDB, empID)
	if err != nil || deletedEmp != nil {
		t.Errorf("Employee was not deleted")
	}
}

func TestGetEmployeesBatch(t *testing.T) {
	// Create some test employees
	employees := []empdb.Employee{
		{Name: "Alice", Position: "Engineer", Salary: 60000.0, DateCreated: &time.Time{}},
		{Name: "Bob", Position: "Manager", Salary: 70000.0, DateCreated: &time.Time{}},
		{Name: "Charlie", Position: "Director", Salary: 80000.0, DateCreated: &time.Time{}},
	}
	for _, emp := range employees {
		_, err := empdb.CreateEmmployee(testDB, emp)
		if err != nil {
			t.Errorf("Failed to create employee: %v", err)
		}
	}

	// Test pagination
	params := empdb.PaginationParams{Page: 1, Count: 2}
	batch, err := empdb.GetEmployeesBatch(testDB, params)
	if err != nil {
		t.Errorf("Failed to get employees batch: %v", err)
	}
	if len(batch) != 2 {
		t.Errorf("Incorrect number of employees returned")
	}

	// Clean up
	for _, emp := range employees {
		err := empdb.DeleteEmployee(testDB, emp.Id)
		if err != nil {
			log.Printf("Failed to delete employee: %v", err)
		}
	}
}
