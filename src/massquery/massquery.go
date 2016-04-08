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
	debugMode           bool
	queryArg            string
	connectionStringArg string
	throughMode         bool
	fakeMode            bool
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
		cli.StringFlag{
			Name:        "cnn",
			Usage:       "db connection string",
			Destination: &connectionStringArg,
		},
		cli.BoolFlag{
			Name:        "through",
			Usage:       "through mode",
			Destination: &throughMode,
		},
		cli.BoolFlag{
			Name:        "fake",
			Usage:       "fake mode",
			Destination: &fakeMode,
		},
	}

	app.Action = func(c *cli.Context) {

		// this func's called for each stdin's row
		process := func(row []byte) error {

			var (
				connectionString string
				resPrefix        string
			)

			debug(string(row))

			params := strings.Split(string(row), "\t")
			if len(connectionStringArg) == 0 && len(params) < 2 {
				log.Printf("[ERROR] Не хватает параметров в stdin. Нужно минимум 2. Получено %d\n", len(params))
			}

			if len(connectionStringArg) > 0 {
				connectionString = connectionStringArg
			} else {
				connectionString = params[1]
			}

			query := c.String("query")

			for i, param := range params {
				paramTpl := fmt.Sprintf("{%d}", i+1)
				query = strings.Replace(query, paramTpl, param, -1)
			}

			debug(query)

			if throughMode {
				resPrefix = string(row)
			} else {
				resPrefix = params[0]
			}

			res, err := runQuery(connectionString, query)
			if err != nil {
				fmt.Printf(resPrefix + "\terror\t\n")
				log.Println(err.Error())
			} else {
				fmt.Printf(resPrefix+"\tsuccess\t%s\n", res)
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

	if fakeMode {
		return "", nil
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
