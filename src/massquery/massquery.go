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
	formatArg           string
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
			EnvVar:      "MASSQUERY_CNN",
			Destination: &connectionStringArg,
		},
		cli.StringFlag{
			Name:        "format",
			Usage:       "output format",
			Destination: &formatArg,
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

			var connectionString string

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

			status := "success"
			res, err := runQuery(connectionString, query)
			if err != nil {
				status = "error"
				log.Println(err.Error())
			}

			if len(formatArg) > 0 {
				res := formatArg
				res = strings.Replace(res, "\\t", "\t", -1) // лишнее экранирование при получении аргумента программы
				res = strings.Replace(res, "\\n", "\n", -1) // лишнее экранирование при получении аргумента программы
				res = strings.Replace(res, "{input}", string(row), -1)
				res = strings.Replace(res, "{res}", res, -1)
				res = strings.Replace(res, "{id}", params[0], -1)
				res = strings.Replace(res, "{cnn}", connectionString, -1)
				res = strings.Replace(res, "{status}", status, -1)
				fmt.Println(res)
			} else {
				fmt.Printf(params[0]+"\t"+status+"\t%s\n", res)
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
