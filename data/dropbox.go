package data

import (
	"encoding/base64"
	"errors"
	"os"
	"os/user"
	"path/filepath"
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

	if envPath := os.Getenv("SAMAY_DATA_DIR"); envPath != "" {
		return d.setBasePath(envPath)
	}

	homeDir, err := resolveHomeDir()
	if err != nil {
		return err
	}

	if dropboxPath, err := detectDropboxBasePath(homeDir); err == nil && dropboxPath != "" {
		return d.setBasePath(filepath.Join(dropboxPath, "Samay"))
	}

	return d.setBasePath(defaultDataDir(homeDir))
}

func (d *Dropbox) setBasePath(path string) error {
	if path == "" {
		return errors.New("base path cannot be empty")
	}

	absolute, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(absolute, 0775); err != nil {
		return err
	}

	d.BasePath = absolute
	return nil
}

// SetBasePath overrides the Dropbox base path. Intended for tests.
func SetBasePath(path string) error {
	return DB.setBasePath(path)
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

func detectDropboxBasePath(homeDir string) (string, error) {
	hostPath := filepath.Join(homeDir, ".dropbox", "host.db")
	dat, err := os.ReadFile(hostPath)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(dat), "\n")
	if len(lines) < 2 {
		return "", errors.New("dropbox host.db missing expected data")
	}

	dropboxPath, err := base64.StdEncoding.DecodeString(lines[1])
	if err != nil {
		return "", err
	}

	return string(dropboxPath), nil
}

func resolveHomeDir() (string, error) {
	if current, err := user.Current(); err == nil && current.HomeDir != "" {
		return current.HomeDir, nil
	}
	return os.UserHomeDir()
}

func defaultDataDir(homeDir string) string {
	if configDir, err := os.UserConfigDir(); err == nil && configDir != "" {
		return filepath.Join(configDir, "samay")
	}
	return filepath.Join(homeDir, ".samay")
}
