package employee

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/employees/empdb"
)

func setHeader(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
}

func fromBody[T any](body io.Reader, target T) {
	buf := new(bytes.Buffer)
	buf.ReadFrom(body)
	json.Unmarshal(buf.Bytes(), &target)
}

func returnEmployee[T any](w http.ResponseWriter, withData func() (T, error)) {
	setHeader(w)

	data, serverErr := withData()

	if serverErr != nil {
		w.WriteHeader(500)
		serverErrEmployee, err := json.Marshal(&serverErr)
		if err != nil {
			log.Print(err)
			return
		}
		w.Write(serverErrEmployee)
		return
	}

	dataEmployee, err := json.Marshal(&data)
	if err != nil {
		log.Print(err)
		w.WriteHeader(500)
		return
	}

	w.Write(dataEmployee)
}

func returnErr(w http.ResponseWriter, err error, code int) {
	returnEmployee(w, func() (interface{}, error) {
		errorMessage := struct {
			Err string
		}{
			Err: err.Error(),
		}
		w.WriteHeader(code)
		return errorMessage, nil
	})
}

func CreateEmployee(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "POST" {
			return
		}

		employee := empdb.Employee{}
		fromBody(req.Body, &employee)

		empId, err := empdb.CreateEmmployee(db, employee)
		if err != nil {
			returnErr(w, err, 400)
			return
		}

		returnEmployee(w, func() (interface{}, error) {
			log.Printf("REST CreateEmployee: %v\n", empId)
			return empdb.GetEmployee(db, empId)
		})
	})
}

func GetEmployee(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "GET" {
			return
		}
		employee := empdb.Employee{}
		fromBody(req.Body, &employee)

		returnEmployee(w, func() (interface{}, error) {
			log.Printf("REST GetEmployee: %v\n", employee)
			return empdb.GetEmployee(db, employee.Id)
		})
	})
}

func GetEmployeesBatch(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "GET" {
			return
		}

		queryOptions := empdb.PaginationParams{}
		fromBody(req.Body, &queryOptions)

		if queryOptions.Count <= 0 || queryOptions.Page <= 0 {
			returnErr(w, errors.New("page and count fields are required and must be > 0"), 400)
			return
		}

		returnEmployee(w, func() (interface{}, error) {
			log.Printf("REST GetEmployeesBatch: %v\n", queryOptions)
			return empdb.GetEmployeesBatch(db, queryOptions)
		})
	})
}

func UpdateEmployee(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "PUT" {
			return
		}
		employee := empdb.Employee{}
		fromBody(req.Body, &employee)

		if err := empdb.UpdateEmployee(db, employee); err != nil {
			returnErr(w, err, 400)
			return
		}

		returnEmployee(w, func() (interface{}, error) {
			log.Printf("REST UpdateEmployee: %v\n", employee)
			return empdb.GetEmployee(db, employee.Id)
		})
	})
}

func DeleteEmployee(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "DELETE" {
			return
		}
		employee := empdb.Employee{}
		fromBody(req.Body, &employee)

		if err := empdb.DeleteEmployee(db, employee.Id); err != nil {
			returnErr(w, err, 400)
			return
		}

		returnEmployee(w, func() (interface{}, error) {
			log.Printf("REST DeleteEmployee: %v\n", employee.Id)
			return nil, nil
		})
	})
}

func Serve(db *sql.DB, bind string) {
	http.Handle("/employee/create", CreateEmployee(db))
	http.Handle("/employee/get", GetEmployee(db))
	http.Handle("/employee/get_batch", GetEmployeesBatch(db))
	http.Handle("/employee/update", UpdateEmployee(db))
	http.Handle("/employee/delete", DeleteEmployee(db))
	log.Printf("REST API server listening on %v\n", bind)
	err := http.ListenAndServe(bind, nil)
	if err != nil {
		log.Fatalf("REST server error: %v", err)
	}
}
