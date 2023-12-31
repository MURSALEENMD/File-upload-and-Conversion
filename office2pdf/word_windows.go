// +build windows
// word_windows
package office2pdf
import (
"fmt"
"path/filepath"
"strings"
ole "github.com/go-ole/go-ole"
oleutil "github.com/go-ole/go-ole/oleutil"
)
type Word struct {
app *ole.IDispatch
documents *ole.VARIANT
doc *ole.VARIANT
}
func (wd *Word) open(inFile string) (err error) {
ole.CoInitialize(0)
var unknown *ole.IUnknown
unknown, err = oleutil.CreateObject("Word.Application")
if err != nil {
return
}
wd.app, err = unknown.QueryInterface(ole.IID_IDispatch)
if err != nil {
return
}
_, err = oleutil.PutProperty(wd.app, "Visible", false)
if err != nil {
return
}
_, err = oleutil.PutProperty(wd.app, "DisplayAlerts", 0)
if err != nil {
return
}
wd.documents, err = oleutil.GetProperty(wd.app, "Documents")
if err != nil {
return
}
wd.doc, err = oleutil.CallMethod(wd.documents.ToIDispatch(), "Open", inFile)
if err != nil {
fmt.Println(err)
return
}
return
}
func (wd *Word) Export(inFile, outDir string) (outFile string, err error) {
s := strings.Split(filepath.Base(inFile), ".")
outFile = filepath.Join(outDir, filepath.Base(s[0]+".pdf"))
//outFile = filepath.Join(outDir, filepath.Base(inFile+".pdf"))
defer func() {
if err != nil {
outFile = ""
}
wd.close()
}()
err = wd.open(inFile)
if err != nil {
fmt.Println(err)
return
}
_, err = oleutil.CallMethod(wd.doc.ToIDispatch(), "ExportAsFixedFormat", outFile, 17)
if err != nil {
return
}
return
}
func (wd *Word) close() {
if wd.doc != nil {
oleutil.MustPutProperty(wd.doc.ToIDispatch(), "Saved", true)
}
if wd.documents != nil {
oleutil.MustCallMethod(wd.documents.ToIDispatch(), "Close")
}
if wd.app != nil {
oleutil.MustCallMethod(wd.app, "Quit")
wd.app.Release()
}
ole.CoUninitialize()
}