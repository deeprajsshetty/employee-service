package main

import (
	"database/sql"
	"log"
	"sync"

	"github.com/employees/api/rest/employee"
	"github.com/employees/empdb"

	"github.com/alexflint/go-arg"
)

var args struct {
	DbPath   string `arg:"env:EMPLOYEE_DB"`
	BindRest string `arg:"env:EMPLOYEE_BIND_REST"`
	BindGrpc string `arg:"env:EMPLOYEE_BIND_GRPC"`
}

func main() {
	arg.MustParse(&args)

	if args.DbPath == "" {
		args.DbPath = "employees.db"
	}
	if args.BindRest == "" {
		args.BindRest = ":8080"
	}
	if args.BindGrpc == "" {
		args.BindGrpc = ":8081"
	}

	log.Printf("using database '%v'\n", args.DbPath)
	db, err := sql.Open("sqlite3", args.DbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	empdb.CreateEmployeeTable(db)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		log.Printf("starting REST API server...\n")
		employee.Serve(db, args.BindRest)
		wg.Done()
	}()

	wg.Wait()
}
