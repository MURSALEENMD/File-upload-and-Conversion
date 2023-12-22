package accountcontrol

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"office2pdf"
	"os"
	"path/filepath"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/sessions"
)

var store = sessions.NewCookieStore([]byte("mysession"))
var DbCon *sql.DB = GetConnection()

type UserInfo struct {
	f_name   string
	l_name   string
	username string
	password string
	email    string
	contact  string
}

func Index(w http.ResponseWriter, r *http.Request) {
	tpl, _ := template.ParseFiles("views/index.html")
	tpl.Execute(w, nil)
}
func Login(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.Form.Get("username")
	password := r.Form.Get("password")
	if CheckUser(username, password) != 0 {
		session, _ := store.Get(r, "mysession")
		session.Values["username"] = username
		b, err := json.Marshal(time.Now())
		if err != nil {
			log.Fatal(err)
		}
		session.Values["timeval"] = b
		session.Save(r, w)
		http.Redirect(w, r, "/account/welcome", http.StatusSeeOther)
	} else {
		data := map[string]interface{}{
			"err": "Invalid Username/Password",
		}
		tpl, _ := template.ParseFiles("views/login.html")
		tpl.Execute(w, data)
	}
}
func Register(w http.ResponseWriter, r *http.Request) {
	tpl, _ := template.ParseFiles("views/register.html")
	tpl.Execute(w, nil)
}
func Registration(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	f_name := r.Form.Get("f_name")
	l_name := r.Form.Get("l_name")
	username := r.Form.Get("username")
	password := r.Form.Get("password1")
	email := r.Form.Get("email")
	contact := r.Form.Get("contact")
	var user UserInfo = UserInfo{f_name, l_name, username, password, email, contact}
	RegisterUser(&user)
	data := map[string]interface{}{
		"msg": "Succesfully Registered",
	}
	tpl, _ := template.ParseFiles("views/register.html")
	tpl.Execute(w, data)
}
func RegisterUser(user *UserInfo) {
	tx, err := DbCon.Begin()
	if err != nil {
		log.Fatal(err)
	}
	stmt, err := tx.Prepare("INSERT INTO users (f_name,l_name,username,password,email,contact) VALUES (?,?,?,?,?,?)")
	if err != nil {
		tx.Rollback()
		panic(err.Error())
	}
	_, err = stmt.Exec(user.f_name, user.l_name, user.username, user.password, user.email, user.contact)
	if err != nil {
		tx.Rollback()
		panic(err.Error())
	}
	stmt.Close()
	path, _ := os.Getwd()
	err = os.Mkdir(path+"/Uploads/"+user.username, 0755)
	if err != nil {
		tx.Rollback()
		panic(err.Error())
	}
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}
}
func CheckTime(w http.ResponseWriter, r *http.Request) string {
	session, _ := store.Get(r, "mysession")
	username := session.Values["username"]
	v := session.Values["timeval"]
	tm := &time.Time{}
	if b, ok := v.([]byte); ok {
		err := json.Unmarshal(b, tm)
		if err != nil {
			log.Fatal(err)
		}
	}
	if time.Since(*tm) > 3*time.Minute {
		session.Options.MaxAge = -1
		session.Save(r, w)
		return "nil"
	} else {
		str := fmt.Sprintf("%v", username)
		b, err := json.Marshal(time.Now())
		if err != nil {
			log.Fatal(err)
		}
		session.Values["timeval"] = b
		session.Save(r, w)
		return str
	}
}
func Welcome(w http.ResponseWriter, r *http.Request) {
	str := CheckTime(w, r)
	if str != "nil" {
		path, _ := os.Getwd()
		files, err := ioutil.ReadDir(path + "/Uploads/" + str)
		if err != nil {
			log.Fatal(err)
		}
		var file []string
		var c int = 0
		for _, f := range files {
			file = append(file, f.Name())
			c++
		}
		data := map[string]interface{}{
			"username":  str,
			"file_list": file,
			"file_no":   c,
		}
		tpl, _ := template.ParseFiles("views/welcome.html")
		tpl.Execute(w, data)
	} else {
		data := map[string]interface{}{
			"err": "Session Over, Please Login !!",
		}
		tpl, _ := template.ParseFiles("views/login.html")
		tpl.Execute(w, data)
	}
}
func Download(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	session, _ := store.Get(r, "mysession")
	username := session.Values["username"]
	str := fmt.Sprintf("%v", username)
	filename := r.Form.Get("file_name")
	check := r.Form.Get("choice")
	path, _ := os.Getwd()
	filepath := path + "\\Uploads\\" + str + "\\"
	if check == "download" {
		http.ServeFile(w, r, path+"/Uploads/"+str+"/"+filename)
	}
	if check == "convert" {
		outdir := path + "\\Uploads\\" + str
		export(filepath+filename, outdir)
		http.Redirect(w, r, "/account/welcome", http.StatusSeeOther)
	}
	if check == "delete" {
		err := os.Remove(path + "/Uploads/" + str + "/" + filename)
		if err != nil {
			log.Fatal(err)
		}
		http.Redirect(w, r, "/account/welcome", http.StatusSeeOther)
	}
}
func Logout(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "mysession")
	session.Options.MaxAge = -1
	session.Save(r, w)
	http.Redirect(w, r, "/account/index", http.StatusSeeOther)
}
func export(inFile, outDir string) {
	if fileIsExist(inFile) && fileIsExist(outDir) {
		exporter := exporterMap()[filepath.Ext(inFile)]
		if _, ok := exporter.(office2pdf.Exporter); ok {
			outFile, err := exporter.(office2pdf.Exporter).Export(inFile, outDir)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("Output File: " + outFile)
		}
	}
}
func fileIsExist(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}
func exporterMap() (m map[string]interface{}) {
	m = map[string]interface{}{
		".doc":  new(office2pdf.Word),
		".docx": new(office2pdf.Word),
		".xls":  new(office2pdf.Excel),
		".xlsx": new(office2pdf.Excel),
	}
	return
}
func GetConnection() *sql.DB {
	db, err := sql.Open("mysql", "root:Murs@2000tcp(127.0.0.1:3306)/go_web")
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("Database Connection Done")
	return db
}
func CheckUser(username, password string) int {
	users, err := DbCon.Query("select * from users where username=? and password=?", username, password)
	if err != nil {
		panic(err.Error())
	}
	var c int = 0
	for users.Next() {
		c++
		if err != nil {
			panic(err.Error())
		}
	}
	return c
}
