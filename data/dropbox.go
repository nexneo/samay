package data

import (
	"encoding/base64"
	"errors"

	"os"
	"os/user"
	"strings"

	"google.golang.org/protobuf/proto"
)

var (
	DB *Dropbox
)

func init() {
	DB = new(Dropbox)
	if err := DB.Init(); err != nil {
		panic(err)
	}
}

type Dropbox struct {
	BasePath string
}

func (d Dropbox) String() string {
	return d.BasePath
}

// reads ~/.dropbox/host.db and detects location of Dropbox folder
// and create BasePath directory if doesn't exists
func (d *Dropbox) Init() error {
	if d.BasePath != "" {
		// nothing to do
		return nil
	}

	stderr := errors.New("Dropbox folder can't be detected")

	// read current user's dropbox host.db file
	current_user, _ := user.Current()
	dat, err := os.ReadFile(
		current_user.HomeDir + "/.dropbox/host.db",
	)
	if err != nil {
		return err
	}

	// read bytes and get second line from host.db
	datStr := string(dat)

	if lines := strings.Split(datStr, "\n"); len(lines) < 2 {
		return stderr
	} else {
		datStr = lines[1]
	}

	// second line is base64 encoded Dropbox path
	dropboxPath, err := base64.StdEncoding.DecodeString(datStr)
	if err != nil {
		return err
	}

	if len(dropboxPath) != 0 {
		d.BasePath = string(dropboxPath) + "/Samay"
		// Create samay data folder
		// TODO make sure user doesn't have folder already
		os.Mkdir(d.BasePath, 0775)
	} else {
		return stderr
	}

	return err
}

func (d *Dropbox) ProjectDirPath(p *Project) string {
	return d.BasePath + "/" + p.GetShaFromName()
}

func (d *Dropbox) EntryDirPath(e *Entry) string {
	return DB.ProjectDirPath(e.GetProject()) + "/entries"
}

func (d *Dropbox) MkProjectDir(p *Project) (err error) {
	if err = os.Mkdir(d.ProjectDirPath(p), 0755); os.IsExist(err) {
		err = nil
	}

	return
}

func (d *Dropbox) Projects() (projects []*Project) {
	folders, err := os.ReadDir(d.BasePath)
	if err != nil {
		return
	}
	for _, dir := range folders {
		project := new(Project)
		project.Sha = proto.String(dir.Name())
		if err = Load(project); err == nil {
			projects = append(projects, project)
		}
	}
	return
}
