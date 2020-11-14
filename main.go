package main

import (
	"bytes"
	"database/sql"
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	_ "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	"go/ast"
	"go/build"
	"go/format"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"time"
)

var pkgInfo *build.Package

var db *sql.DB

var (
	target    string
	table     string
	dbAddress string
	schema    string
)

var dbGolangType = map[string]string{
	"int":     "int",
	"tinyint": "int",
	"varchar": "string",
	"float":   "float64",
	"decimal": "float64",
	"char":    "string",
	"enum":    "string",
	"text":    "string",
}

func InitFlag() {
	flag.StringVar(&target, "target", "", "target struct of entity")
	flag.StringVar(&table, "table", "", "the table name")
}

type config struct {
	DbAddress string
	Db        string
}

var conf config

func InitConf() error {
	user, _ := user.Current()
	if _, err := toml.DecodeFile(fmt.Sprintf("%s/.entity-helper/conf.toml", user.HomeDir), &conf); err != nil {
		return err
	}
	return nil
}

func main() {
	err := InitConf()
	if err != nil {
		log.Fatal("[Error] fail to load config", err.Error())
	}
	dbAddress = conf.DbAddress
	schema = conf.Db
	var dbErr error
	db, dbErr = sql.Open("mysql", dbAddress)
	if dbErr != nil {
		log.Fatal("[Error]access database failed." + dbAddress + dbErr.Error())
	}

	InitFlag()
	flag.Parse()
	if target == "" {
		log.Println("must point out target")
		return
	}
	if table == "" {
		log.Println("must point out table")
		return
	}

	goBytes, err := covertTableSchemaToEntity(table, schema)
	if err != nil {
		log.Fatal("[Error]fail to covert table Schema " + err.Error())
	}
	startIdx, _, filename, err := findEntityPositon(target)
	if err != nil {
		log.Fatal("[Error]fail to find entity position " + err.Error())
	}

	if filename == "" { // 找不到命中的 或者找到的struct 已经有值
		return
	}

	// 开始写文件
	f, err := os.OpenFile(fmt.Sprintf("./%s", filename), os.O_RDWR, 0666)
	if err != nil {
		log.Fatal("[Error]open file error" + err.Error())
	}
	defer f.Close()
	stat, _ := f.Stat()
	fmt.Println(fmt.Sprintf("[Debug] file size => %d", stat.Size()))
	front := make([]byte, startIdx)
	_, err = f.Read(front)
	if err != nil {
		log.Fatal("[Error] fail to read front", err.Error())
	}
	_, err = f.Seek(int64(startIdx), 0)
	if err != nil {
		log.Fatal("[Error] fail to seed to struct positon ", err.Error())
	}
	remainder, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal("[Error] fail to read remainder", err.Error())
	}
	buf := bytes.NewBuffer(front)
	buf.Write(goBytes)
	buf.Write(remainder)
	final, err := format.Source(buf.Bytes())
	fmt.Println(string(final))
	if err != nil {
		log.Fatal("[Error]fail to format", err.Error())
	}
	err = f.Truncate(0)
	if err != nil {
		log.Fatal("[Error] fail to truncate", err.Error())
	}
	_, err = f.Seek(0, 0)
	if err != nil {
		log.Fatal("[Error] fail to seek to start", err.Error())
	}
	_, err = f.Write(final)
	if err != nil {
		log.Fatal("[Error] fail to write file", err.Error())
	}

}

func covertTableSchemaToEntity(tableName string, schema string) ([]byte, error) {
	sql := "select * from information_schema.columns where table_name = ? AND table_schema = ? "
	rows, err := db.Query(sql, tableName, schema)
	if err != nil {
		return nil, errors.Wrap(err, "fail to access db")
	}
	info := rows2maps(rows)
	buf := bytes.NewBufferString(fmt.Sprintf("\n// Entity-Helper Start => %s \n ", time.Now().Format("20060102 15:04")))
	for _, r := range info {
		columnName,_ := r["COLUMN_NAME"].(string)
		dataType,_ := r["DATA_TYPE"].(string)
		comment, _ := r["COLUMN_COMMENT"].(string)
		fieldName := toCamelInitCase(columnName, true)
		goType := dbGolangType[dataType]
		goDesc := fmt.Sprintf("%s %s `json:\"%s\"` //%s\n", fieldName, goType, columnName, comment)
		buf.Write([]byte(goDesc))
	}
	buf.Write([]byte(fmt.Sprintf("// Entity-Helper End => %s", time.Now().Format("20060102 15:04"))))
	return buf.Bytes(), nil
}

func findEntityPositon(structName string) (startIdx int64, endIdx int64, filename string, err error) {
	pkgInfo, err = build.ImportDir("./", 0)
	if err != nil {
		return 0, 0, "", errors.Wrap(err, "fail to import Dir")
	}
	fset := token.NewFileSet()
	for _, file := range pkgInfo.GoFiles {
		f, err := parser.ParseFile(fset, file, nil, 0)
		if err != nil {
			log.Fatal(err)
		}
		//ast.Print(fset, f)
		ast.Inspect(f, func(node ast.Node) bool {
			decl, ok := node.(*ast.GenDecl)
			if !ok {
				return true
			}
			if decl.Tok != token.TYPE {
				return true
			}
			specs := decl.Specs
			if len(specs) < 0 {
				return true
			}
			tSpec, ok := specs[0].(*ast.TypeSpec)
			if !ok {
				return true
			}
			typeName := tSpec.Name.Name
			if typeName != structName {
				return true
			}
			structType, ok := tSpec.Type.(*ast.StructType)
			if !ok {
				return true
			}
			if numFields := structType.Fields.NumFields(); numFields > 0 {
				return true
			}
			closing := structType.Fields.Closing
			openning := structType.Fields.Opening
			fileStart := f.Package
			filename = file
			startIdx = int64(openning) - int64(fileStart) + 1
			endIdx = int64(closing) - int64(fileStart)
			log.Println("[DEBUG]", file, typeName, openning, closing, startIdx, endIdx)
			return false
		})
	}
	return
}
