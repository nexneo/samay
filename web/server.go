package web

import (
	"code.google.com/p/goprotobuf/proto"
	"encoding/json"
	"github.com/nexneo/samay/data"
	"io"
	"net/http"
	"os/exec"
)

func StartServer() error {
	http.Handle("/",
		http.FileServer(http.Dir("./public")),
	)
	http.HandleFunc("/app.json", appJson)
	go exec.Command("open", "http://localhost:8080/").Run()
	return http.ListenAndServe(":8080", nil)
}

type ProjectSet struct {
	Project *data.Project `json:"project"`
	Entries []*data.Entry `json:"entries"`
}

func appJson(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	projects := make([]*ProjectSet, 0, 20)
	for _, project := range data.DB.Projects() {
		project.Sha = proto.String(project.GetShaFromName())
		projects = append(projects, &ProjectSet{project, project.Entries()})
	}
	by, err := json.Marshal(projects)
	if err != nil {
		panic(err)
	}
	io.WriteString(w, string(by))
}
