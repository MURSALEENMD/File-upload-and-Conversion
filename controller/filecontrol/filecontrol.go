package filecontrol
import (
 "fmt"
 "io/ioutil"
 "net/http"
 "os"
 "github.com/gorilla/sessions"
)
var store = sessions.NewCookieStore([]byte("mysession"))
func Upload(w http.ResponseWriter, r *http.Request) {
 r.ParseMultipartForm(10 << 20)
 file, handler, err := r.FormFile("myfile")
 if err != nil {
 fmt.Println("Error getting the file")
 fmt.Println(err)
 return
 }
 defer file.Close()
 fmt.Printf("Uploaded File: %+v\n", handler.Filename)
 fmt.Printf("File Size: %+v\n", handler.Size)
 fmt.Printf("MIME Header: %+v\n", handler.Header)
 session, _ := store.Get(r, "mysession")
 username := session.Values["username"]
 str := fmt.Sprintf("%v", username)
 path, _ := os.Getwd()
 filepath := path + "\\Uploads\\" + str + "\\"
 dst, err := os.Create(filepath + handler.Filename)
 defer dst.Close()
 if err != nil {
 fmt.Println(err)
 }
 fileBytes, err := ioutil.ReadAll(file)
 if err != nil {
 fmt.Println(err)
 }
 _, err = dst.Write(fileBytes)
 if err != nil {
 fmt.Println(err)
 }
 http.Redirect(w, r, "/account/welcome", http.StatusSeeOther)
}
