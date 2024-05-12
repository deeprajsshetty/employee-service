package employee_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/employees/api/rest/employee"
	"github.com/employees/empdb"
)

func TestCreateEmployee(t *testing.T) {
	mockDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open mock database: %v", err)
	}
	defer mockDB.Close()

	empdb.CreateEmployeeTable(mockDB)

	tests := []struct {
		name       string
		body       string
		statusCode int
		want       *empdb.Employee
	}{
		{
			name:       "Valid request",
			body:       `{"name": "John Doe", "position": "Manager", "salary": 500000}`,
			statusCode: http.StatusOK,
			want: &empdb.Employee{
				Id:          1,
				Name:        "John Doe",
				Position:    "Manager",
				Salary:      500000,
				DateCreated: nil, // The date_created field will be set by the database
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/employee/create", strings.NewReader(tt.body))
			rr := httptest.NewRecorder()

			handler := employee.CreateEmployee(mockDB)
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.statusCode {
				t.Errorf("handler returned wrong status code: got %v want %v",
					rr.Code, tt.statusCode)
			}

			if tt.statusCode == http.StatusOK {
				var actual empdb.Employee
				if err := json.Unmarshal(rr.Body.Bytes(), &actual); err != nil {
					t.Errorf("error unmarshalling response: %v", err)
				}

				if actual.Id != tt.want.Id {
					t.Errorf("handler returned unexpected id: got %v want %v",
						actual.Id, tt.want.Id)
				}
				if actual.Name != tt.want.Name {
					t.Errorf("handler returned unexpected name: got %v want %v",
						actual.Name, tt.want.Name)
				}
				if actual.Position != tt.want.Position {
					t.Errorf("handler returned unexpected position: got %v want %v",
						actual.Position, tt.want.Position)
				}
				if actual.Salary != tt.want.Salary {
					t.Errorf("handler returned unexpected salary: got %v want %v",
						actual.Salary, tt.want.Salary)
				}
				if !actual.DateCreated.After(time.Time{}) {
					t.Errorf("handler returned unexpected date_created: got %v want > %v",
						actual.DateCreated, time.Time{})
				}
			}
		})
	}
}

func TestGetEmployee(t *testing.T) {
	mockDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open mock database: %v", err)
	}
	defer mockDB.Close()

	empdb.CreateEmployeeTable(mockDB)
	mockEmployee := empdb.Employee{Id: 1, Name: "Alice", Position: "Engineer", Salary: 60000}
	empdb.CreateEmmployee(mockDB, mockEmployee)

	tests := []struct {
		name       string
		body       string
		id         int64
		statusCode int
		want       *empdb.Employee
	}{
		{
			name:       "Valid request",
			body:       `{"id": 1}`,
			id:         1,
			statusCode: http.StatusOK,
			want:       &mockEmployee,
		},
		{
			name:       "Employee not found",
			body:       `{"id": 2}`,
			id:         2,
			statusCode: http.StatusOK, // Assuming it returns nil and 200 for simplicity
			want:       &empdb.Employee{Id: 0, Name: "", Position: "", Salary: 0, DateCreated: nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/employee/get", strings.NewReader(tt.body))
			rr := httptest.NewRecorder()

			handler := employee.GetEmployee(mockDB)
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.statusCode {
				t.Errorf("handler returned wrong status code: got %v want %v",
					rr.Code, tt.statusCode)
			}

			if tt.statusCode == http.StatusOK {
				var actual empdb.Employee
				if err := json.Unmarshal(rr.Body.Bytes(), &actual); err != nil {
					t.Errorf("error unmarshalling response: %v", err)
				}

				// Resetting a Date Created
				if tt.want != nil {
					tt.want.DateCreated = actual.DateCreated
				}

				if tt.want == nil && &actual != nil {
					t.Errorf("handler returned unexpected body: got %v want %v",
						actual, tt.want)
				} else if tt.want != nil && !reflect.DeepEqual(actual, *tt.want) {
					t.Errorf("handler returned unexpected body: got %v want %v",
						actual, *tt.want)
				}
			}
		})
	}
}

func TestUpdateEmployee(t *testing.T) {
	mockDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open mock database: %v", err)
	}
	defer mockDB.Close()

	empdb.CreateEmployeeTable(mockDB)
	mockEmployee := empdb.Employee{Id: 1, Name: "Alice", Position: "Engineer", Salary: 60000, DateCreated: nil}
	empId, _ := empdb.CreateEmmployee(mockDB, mockEmployee)
	insertedMockEmp, _ := empdb.GetEmployee(mockDB, empId)

	newEmployee := empdb.Employee{Id: 1, Name: "Alice Smith", Position: "Senior Engineer", Salary: 70000, DateCreated: insertedMockEmp.DateCreated}

	body, err := json.Marshal(newEmployee)
	if err != nil {
		t.Fatalf("Failed to marshal request body: %v", err)
	}

	req := httptest.NewRequest("PUT", "/employee/update", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	handler := employee.UpdateEmployee(mockDB)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			rr.Code, http.StatusOK)
	}

	updatedEmployee, err := empdb.GetEmployee(mockDB, 1)
	if err != nil {
		t.Fatalf("Failed to get updated employee: %v", err)
	}

	if !reflect.DeepEqual(*updatedEmployee, newEmployee) {
		t.Errorf("handler did not update employee correctly: got %v want %v",
			*updatedEmployee, newEmployee)
	}
}

func TestGetEmployeesBatch(t *testing.T) {
	mockDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open mock database: %v", err)
	}
	defer mockDB.Close()

	empdb.CreateEmployeeTable(mockDB)
	mockEmployees := []empdb.Employee{
		{Id: 1, Name: "Alice", Position: "Engineer", Salary: 60000, DateCreated: nil},
		{Id: 2, Name: "Bob", Position: "Manager", Salary: 70000, DateCreated: nil},
		{Id: 3, Name: "Charlie", Position: "Director", Salary: 80000, DateCreated: nil},
	}
	for _, emp := range mockEmployees {
		empdb.CreateEmmployee(mockDB, emp)
	}

	body := `{"page": 1, "count": 2}`
	req := httptest.NewRequest("GET", "/employee/get_batch", strings.NewReader(body))
	rr := httptest.NewRecorder()

	handler := employee.GetEmployeesBatch(mockDB)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			rr.Code, http.StatusOK)
	}

	var actual []empdb.Employee
	if err := json.Unmarshal(rr.Body.Bytes(), &actual); err != nil {
		t.Errorf("error unmarshalling response: %v", err)
	}

	// As of now I am resetting the Date
	for i, emp := range actual {
		mockEmployees[i].DateCreated = emp.DateCreated
	}
	expected := mockEmployees[:2] // First 2 employees

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("handler returned unexpected body: got %v want %v",
			actual, expected)
	}
}
