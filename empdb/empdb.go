package empdb

import (
	"database/sql"
	"log"
	"time"

	"github.com/mattn/go-sqlite3"
)

type Employee struct {
	Id          int64      `json:"id"`
	Name        string     `json:"name"`
	Position    string     `json:"position"`
	Salary      float64    `json:"salary"`
	DateCreated *time.Time `json:"date_created"`
}

func CreateEmployeeTable(db *sql.DB) {
	_, err := db.Exec(`
		CREATE TABLE employees (
			id            INTEGER PRIMARY KEY,
			name		  TEXT,
			position	  TEXT,
			salary		  NUMERIC,
			date_created  INTEGER
		);
	`)
	if err != nil {
		if sqlError, ok := err.(sqlite3.Error); ok {
			// code 1 == "table already exists"
			if sqlError.Code != 1 {
				log.Fatal(sqlError)
			}
		} else {
			log.Fatal(err)
		}
	}
}

func employeeFromRows(row *sql.Rows) (*Employee, error) {
	var id int64
	var name string
	var position string
	var salary float64
	var dateCreated int64

	err := row.Scan(&id, &name, &position, &salary, &dateCreated)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	t := time.Unix(dateCreated, 0)
	return &Employee{Id: id, Name: name, Position: position, Salary: salary, DateCreated: &t}, nil
}

func CreateEmmployee(db *sql.DB, emp Employee) (int64, error) {
	result, err := db.Exec(`INSERT INTO
		employees(name, position, salary, date_created)
		VALUES(?, ?, ?, 0)`, emp.Name, emp.Position, emp.Salary)

	if err != nil {
		log.Println(err)
		return 0, err
	}
	empId, err := result.LastInsertId()
	log.Println(empId)
	if err != nil {
		log.Println(err)
		return 0, err
	}

	return empId, err
}

func GetEmployee(db *sql.DB, id int64) (*Employee, error) {
	rows, err := db.Query(`
		SELECT id, name, position, salary, date_created
		FROM employees
		WHERE id = ?`, id)

	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		return employeeFromRows(rows)
	}
	return nil, nil
}

func UpdateEmployee(db *sql.DB, emp Employee) error {
	t := emp.DateCreated.Unix()
	_, err := db.Exec(`UPDATE employees SET 
		  name=?,
		  position=?,
		  salary=?,
		  date_created=?
		  WHERE id = ?`, emp.Name, emp.Position, emp.Salary, t, emp.Id)

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func DeleteEmployee(db *sql.DB, id int64) error {
	_, err := db.Exec(`
		DELETE FROM employees
		WHERE id=?`, id)

	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

type PaginationParams struct {
	Page  int
	Count int
}

func GetEmployeesBatch(db *sql.DB, params PaginationParams) ([]Employee, error) {
	var empty []Employee

	rows, err := db.Query(`
		SELECT id, name, position, salary, date_created
		FROM employees
		ORDER BY id ASC
		LIMIT ? OFFSET ?`, params.Count, (params.Page-1)*params.Count)
	if err != nil {
		log.Println(err)
		return empty, err
	}

	defer rows.Close()

	emps := make([]Employee, 0, params.Count)

	for rows.Next() {
		emp, err := employeeFromRows(rows)
		if err != nil {
			return nil, err
		}
		emps = append(emps, *emp)
	}

	return emps, nil
}
