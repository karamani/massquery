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
	execArg             string
	connectionStringArg string
	formatArg           string
	fakeMode            bool
	isExec              bool
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
			Name:        "exec",
			Usage:       "exec-string (insert, update or delete)",
			Destination: &execArg,
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

		if len(connectionStringArg) == 0 {
			log.Println("[ERROR] 'cnn' arg is required")
			return
		}

		if len(queryArg) == 0 && len(execArg) == 0 {
			log.Println("[ERROR] It should be one of the arguments: 'query' or 'exec'")
			return
		}

		query := queryArg
		isExec = len(query) == 0
		if isExec {
			query = execArg
		}

		debug("%#v", isExec)
		if !iostreams.StdinReady() {
			res, err := runQuery(connectionStringArg, queryArg)
			if err != nil {
				log.Println(err.Error())
				printRes(formatArg, "", "", connectionStringArg, "error", nil)
			} else {
				debug("%#v", res)
				for _, resrow := range res {
					printRes(formatArg, "", "", connectionStringArg, "success", resrow)
				}
			}
			return
		}

		// this func's called for each stdin's row
		process := func(row []byte) error {

			debug(string(row))

			params := strings.Split(string(row), "\t")

			rowQuery := query
			rowCnn := connectionStringArg
			for i, param := range params {
				tpl := fmt.Sprintf("{%d}", i)
				rowQuery = strings.Replace(rowQuery, tpl, param, -1)
				rowCnn = strings.Replace(rowCnn, tpl, param, -1)
			}

			debug(rowQuery)
			debug(rowCnn)

			status := "success"
			res, err := runQuery(rowCnn, rowQuery)
			if err != nil {
				status = "error"
				log.Println(err.Error())
				printRes(formatArg, string(row), params[0], rowCnn, status, nil)
			}

			for _, resrow := range res {
				printRes(formatArg, string(row), params[0], rowCnn, status, resrow)
			}

			return nil
		}

		if err := iostreams.ProcessStdin(process); err != nil {
			log.Panicln(err.Error())
		}
	}

	app.Run(os.Args)
}

func printRes(format, input, id, cnn, status string, res []string) {

	resString := strings.Join(res, "\t")

	if len(format) > 0 {
		s := format
		s = strings.Replace(s, "\\t", "\t", -1) // unnecessary quotes from command line
		s = strings.Replace(s, "\\n", "\n", -1) // unnecessary quotes from command line
		s = strings.Replace(s, "{input}", input, -1)
		s = strings.Replace(s, "{res}", resString, -1)
		s = strings.Replace(s, "{cnn}", cnn, -1)
		s = strings.Replace(s, "{status}", status, -1)
		for i, r := range res {
			tpl := fmt.Sprintf("{res%d}", i)
			s = strings.Replace(s, tpl, r, -1)
		}
		fmt.Println(s)
	} else {
		if len(resString) > 0 {
			fmt.Println(resString)
		}
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
