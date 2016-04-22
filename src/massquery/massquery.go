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

		connectionString := connectionStringArg
		queryArg := c.String("query")

		if !iostreams.StdinReady() {
			res, err := runQuery(connectionString, queryArg)
			if err != nil {
				log.Println(err.Error())
				printRes(formatArg, "", "", connectionString, "error", "")
			} else {
				debug("%#v", res)
				for _, resrow := range res {
					s := strings.Join(resrow, "\t")
					printRes(formatArg, "", "", connectionString, "success", s)
				}
			}
			return
		}

		// this func's called for each stdin's row
		process := func(row []byte) error {

			debug(string(row))

			params := strings.Split(string(row), "\t")

			if len(connectionStringArg) == 0 && len(params) < 2 {
				log.Printf("[ERROR] Не хватает параметров в stdin. Нужно минимум 2. Получено %d\n", len(params))
				return nil
			}

			if len(connectionStringArg) == 0 {
				connectionString = params[1]
			}

			query := queryArg
			for i, param := range params {
				paramTpl := fmt.Sprintf("{%d}", i)
				query = strings.Replace(query, paramTpl, param, -1)
			}

			debug(query)

			status := "success"
			res, err := runQuery(connectionString, query)
			if err != nil {
				status = "error"
				log.Println(err.Error())
				printRes(formatArg, string(row), params[0], connectionString, status, "")
			}

			for _, resrow := range res {
				s := strings.Join(resrow, "\t")
				printRes(formatArg, string(row), params[0], connectionString, status, s)
			}

			return nil
		}

		if err := iostreams.ProcessStdin(process); err != nil {
			log.Panicln(err.Error())
		}
	}

	app.Run(os.Args)
}

func printRes(format, input, id, cnn, status, res string) {
	if len(format) > 0 {
		s := format
		s = strings.Replace(s, "\\t", "\t", -1) // лишнее экранирование при получении аргумента программы
		s = strings.Replace(s, "\\n", "\n", -1) // лишнее экранирование при получении аргумента программы
		s = strings.Replace(s, "{input}", input, -1)
		s = strings.Replace(s, "{res}", res, -1)
		s = strings.Replace(s, "{id}", id, -1)
		s = strings.Replace(s, "{cnn}", cnn, -1)
		s = strings.Replace(s, "{status}", status, -1)
		fmt.Println(s)
	} else {
		fmt.Printf("%s\t%s\t%s\n", id, status, res)
	}
}

func runQuery(connectionString, query string) ([][]string, error) {

	var (
		res       [][]string
		container []string
		pointers  []interface{}
	)

	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	if fakeMode {
		return nil, nil
	}

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	colsCount := len(cols)

	for rows.Next() {

		pointers = make([]interface{}, colsCount)
		container = make([]string, colsCount)
		for i, _ := range pointers {
			pointers[i] = &container[i]
		}

		if err := rows.Scan(pointers...); err != nil {
			return nil, err
		}
		res = append(res, container)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

func debug(format string, args ...interface{}) {
	if debugMode {
		log.Printf("[DEBUG] "+format+"\n", args...)
	}
}
