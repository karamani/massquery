package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
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
			res, err := runQuery(connectionStringArg, queryArg, isExec)
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
			rowQuery := parameterizedString(query, "{%d}", params)
			rowCnn := parameterizedString(connectionStringArg, "{%d}", params)

			debug("connection:" + rowCnn)
			debug("query:" + rowQuery)

			res, err := runQuery(rowCnn, rowQuery, isExec)
			if err != nil {
				log.Println(err.Error())
				printRes(formatArg, string(row), params[0], rowCnn, "error", nil)
			}

			for _, resrow := range res {
				printRes(formatArg, string(row), params[0], rowCnn, "success", resrow)
			}

			return nil
		}

		if err := iostreams.ProcessStdin(process); err != nil {
			log.Panicln(err.Error())
		}
	}

	app.Run(os.Args)
}

func parameterizedString(s, tpl string, params []string) (res string) {
	res = s
	for i, param := range params {
		t := fmt.Sprintf(tpl, i)
		res = strings.Replace(res, t, param, -1)
	}
	return
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
		s = parameterizedString(s, "{res%d}", res)
		fmt.Println(s)
	} else {
		if len(resString) > 0 {
			fmt.Println(resString)
		}
	}
}

func runQuery(connectionString, query string, isExec bool) (res [][]string, resErr error) {

	res, resErr = nil, nil

	db, resErr := sql.Open("mysql", connectionString)
	if resErr != nil {
		log.Println(resErr.Error())
		return
	}
	defer db.Close()

	resErr = db.Ping()
	if resErr != nil {
		return
	}

	if fakeMode {
		return
	}

	if isExec {

		resErr = func() error {
			execRes, err := db.Exec(query)
			if err != nil {
				return err
			}
			affected, err := execRes.RowsAffected()
			if err != nil {
				return err
			}
			lastInsertId, err := execRes.LastInsertId()
			if err != nil {
				return err
			}
			res = append(res, []string{
				strconv.FormatInt(affected, 10),
				strconv.FormatInt(lastInsertId, 10),
			})
			return nil
		}()

	} else {

		resErr = func() error {

			var (
				container []sql.RawBytes
				pointers  []interface{}
				values    []string
			)

			rows, err := db.Query(query)
			if err != nil {
				return err
			}
			defer rows.Close()

			cols, err := rows.Columns()
			if err != nil {
				return err
			}

			colsCount := len(cols)

			for rows.Next() {

				pointers = make([]interface{}, colsCount)
				container = make([]sql.RawBytes, colsCount)
				for i, _ := range pointers {
					pointers[i] = &container[i]
				}

				if err := rows.Scan(pointers...); err != nil {
					return err
				}

				values = make([]string, colsCount)
				for i, elem := range container {
					values[i] = ""
					if elem != nil {
						values[i] = string(elem)
					}
				}

				res = append(res, values)
			}

			if err := rows.Err(); err != nil {
				res = nil
				return err
			}
			return nil
		}()
	}

	return
}

func debug(format string, args ...interface{}) {
	if debugMode {
		log.Printf("[DEBUG] "+format+"\n", args...)
	}
}
