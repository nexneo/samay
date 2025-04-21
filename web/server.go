package web

import (
	"encoding/json"
	"fmt"
	"go/build"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"os/exec"

	"github.com/gorilla/mux"
	"github.com/nexneo/samay/data"
	"google.golang.org/protobuf/proto"
)

var (
	router *mux.Router
	proxy  *httputil.ReverseProxy
)

func init() {
	router = mux.NewRouter()
	router.HandleFunc("/projects", index)
	router.HandleFunc("/entries/{id}", update)
	router.Handle("/", router.NotFoundHandler)
}

func StartServer(port string) error {
	samayPkg, err := build.Import("github.com/nexneo/samay", "", build.FindOnly)
	if err != nil {
		log.Fatalln(err)
	}
	http.Handle("/", router)
	http.Handle("/a/",
		http.StripPrefix(
			"/a/", http.FileServer(http.Dir(samayPkg.Dir+"/public")),
		),
	)
	url := "http://localhost" + port + "/a/index.html"
	go exec.Command("open", url).Run()
	fmt.Printf("starting %s\n", url)
	return http.ListenAndServe(port, nil)
}

type ProjectSet struct {
	Project *data.Project `json:"project"`
	Entries []*data.Entry `json:"entries"`
}

func index(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	projects := make([]*ProjectSet, 0, 20)
	for _, project := range data.DB.Projects() {
		project.Sha = proto.String(project.GetShaFromName())
		projects = append(projects, &ProjectSet{project, project.Entries()})
	}

	b, err := json.Marshal(projects)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, string(b))
}

func update(w http.ResponseWriter, req *http.Request) {
	fmt.Println(req.URL.Path)
	w.Header().Set("Content-Type", "application/json")
	entry := &data.Entry{}

	b, _ := ioutil.ReadAll(req.Body)

	if err := json.Unmarshal(b, entry); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := data.Update(entry); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
