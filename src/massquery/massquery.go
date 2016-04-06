package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/codegangsta/cli"
	_ "github.com/go-sql-driver/mysql"
	"github.com/karamani/iostreams"
)

var (
	debugMode bool
	queryArg  string
)

func main() {
	app := cli.NewApp()
	app.Name = "massquery"
	app.Usage = "massquery"
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:        "debug",
			Usage:       "debug mode",
			EnvVar:      "MASSQUERY_DEBUG",
			Destination: &debugMode,
		},
		cli.StringFlag{
			Name:        "query",
			Usage:       "sql-query",
			Destination: &queryArg,
		},
	}

	app.Action = func(c *cli.Context) {

		// this func's called for each stdin's row
		process := func(row []byte) error {

			debug(string(row))

			params := strings.Split(string(row), "\t")
			if len(params) < 2 {
				log.Printf("[ERROR] Не хватает параметров в stdin. Нужно минимум 2. Получено %d\n", len(params))
			}

			res, err := runQuery(params[1], c.String("query"))
			if err != nil {
				fmt.Printf(os.Stdout, "%s\terror\t\n", params[0])
				log.Println(err.Error())
			} else {
				fmt.Printf("%s\tsuccess\t%s\n", params[0], res)
			}
			return nil
		}

		err := iostreams.ProcessStdin(process)
		if err != nil {
			log.Panicln(err.Error())
		}
	}

	app.Run(os.Args)
}

func runQuery(connectionString, query string) (string, error) {

	var res string

	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		log.Println(err.Error())
		return "", err
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		return "", err
	}

	rows, err := db.Query(query)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&res); err != nil {
			return "", err
		}
		break
	}
	if err := rows.Err(); err != nil {
		return "", err
	}

	return res, nil
}

func debug(format string, args ...interface{}) {
	if debugMode {
		log.Printf("[DEBUG] "+format+"\n", args...)
	}
}
