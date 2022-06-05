package gspider

import (
	"io"
	"log"
	"os"
)

var Log *log.Logger

//var LogW *log.Logger

func init() {
	//Ldate：输出当地时区的日期，如2020/02/07；
	//Ltime：输出当地时区的时间，如11:45:45；
	//Lmicroseconds：输出的时间精确到微秒，设置了该选项就不用设置Ltime了。如11:45:45.123123；
	//Llongfile：输出长文件名+行号，含包名，如github.com/darjun/go-daily-lib/log/flag/main.go:50；
	//Lshortfile：输出短文件名+行号，不含包名，如main.go:50；
	//LUTC：如果设置了Ldate或Ltime，将输出 UTC 时间，而非当地时区。
	Log = log.New(io.MultiWriter(os.Stdout), "gSpider ",
		log.Llongfile|log.LstdFlags)

	//file := "./logs/" + time.Now().Format("20060102") + ".txt"
	//logFile, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0766)
	//if err != nil {
	//	panic(err)
	//}
	//LogW = log.New(io.MultiWriter(os.Stdout, logFile), "gSpider ",
	//	log.Llongfile|log.LstdFlags)
}
