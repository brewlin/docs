
package main
 
import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"net/http"
	"net/url"
	"time"
	"log"
)
var apath = "/images/blog/gc-learning/"
var YNOTE_LOGIN = "5||1604911447813"
var YNOTE_SESS  = "v2|V3AsxsVOj0PShMOfk4OfRUMk4OGO4gZ0lf6LOfk4JF0gLn4J4hHeuRJFP4YlhM6uRlWRfPK6MYMRQF646KhMYM0JuRLYGRHwLR"
//http请求
func RandString(len int) string {
    r := rand.New(rand.NewSource(time.Now().UnixNano()))
    bytes := make([]byte, len)
    for i := 0; i < len; i++ {
        b := r.Intn(26) + 65
        bytes[i] = byte(b)
    }
    return string(bytes)
}
func httpHandle(method, urlVal string) string {
	client := &http.Client{}
	var req *http.Request

	urlArr := strings.Split(urlVal,"?")
	if len(urlArr)  == 2 {
		urlVal = urlArr[0] + "?" + getParseParam(urlArr[1])
	}
	req, _ = http.NewRequest(method, urlVal, nil)

 
	//添加cookie，key为X-Xsrftoken，value为df41ba54db5011e89861002324e63af81
        //可以添加多个cookie
	cookie1 := &http.Cookie{Name: "YNOTE_LOGIN",Value: YNOTE_LOGIN, HttpOnly: true}
	cookie5 := &http.Cookie{Name: "YNOTE_SESS",Value: YNOTE_SESS, HttpOnly: true}
	req.AddCookie(cookie1)
	req.AddCookie(cookie5)
 
	resp, err := client.Do(req)
 
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	b, _ := ioutil.ReadAll(resp.Body)
	name := RandString(10)+".png"
	savepath := "/git/docs/source"+ apath +name
	ioutil.WriteFile(savepath, b, 0666)
	fmt.Println(savepath)
	return apath+name

}
 
//将get请求的参数进行转义
func getParseParam(param string) string  {
	return url.PathEscape(param)
}
 
func main() {
	if len(os.Args) != 2 {
		fmt.Println("lack of config file, eg: go run main.go ${path_of_file} ${old_string} ${new_string}")
		os.Exit(-1)
	}
	fileName := os.Args[1]
	in, err := os.Open(fileName)
	if err != nil {
		fmt.Println("open file fail:", err)
		os.Exit(-1)
	}
	defer in.Close()
 
	out, err := os.OpenFile(fileName+".md", os.O_RDWR|os.O_CREATE, 0766)
	if err != nil {
		fmt.Println("Open write file fail:", err)
		os.Exit(-1)
	}
	defer out.Close()
 
	br := bufio.NewReader(in)
	for {
		line, _, err := br.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("read err:", err)
			os.Exit(-1)
		}
		reg := regexp.MustCompile(`\!\[.*?\]\((.*?)\)`)

		findres := reg.FindSubmatch(line)
		if len(findres) != 0{
			newline := httpHandle("GET",string(findres[1]))
			_, err = out.WriteString("![]("+newline + ")\n")
			if err != nil {
				fmt.Println("write to file fail:", err)
				os.Exit(-1)
			}
		}else{
			newline := string(line)
			_, err = out.WriteString(newline + "\n")
			if err != nil {
				fmt.Println("write to file fail:", err)
				os.Exit(-1)
			}
		}
	}
}