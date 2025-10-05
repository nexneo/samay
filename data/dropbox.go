package data

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
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
	store    *Store
	repos    repositorySet
}

// String returns the resolved Dropbox data directory for debugging and logging.
func (d Dropbox) String() string {
	return d.BasePath
}

// Init discovers and prepares the on-disk store, falling back to ~/.samay when Dropbox is unavailable.
// It is safe to call multiple times and only reconfigures when BasePath has not been seeded.
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

	if err := d.setBasePath(defaultDataDir(homeDir)); err != nil {
		return err
	}
	d.store = defaultStore
	d.repos = newFilesystemRepositories(d, d.store)
	return nil
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
	if d.store == nil {
		d.store = defaultStore
	}
	d.repos = newFilesystemRepositories(d, d.store)
	return nil
}

// SetBasePath overrides the Dropbox base path and rebuilds the repository set. Intended for tests.
func SetBasePath(path string) error {
	return DB.setBasePath(path)
}

// ProjectDirPath returns the folder dedicated to the given project's data within the Dropbox tree.
func (d *Dropbox) ProjectDirPath(p *Project) string {
	return d.BasePath + "/" + p.GetShaFromName()
}

// EntryDirPath returns the directory where entries for the entry's project are stored.
func (d *Dropbox) EntryDirPath(e *Entry) string {
	return d.EntryDirForProject(e.GetProject())
}

// EntryDirForProject resolves the entries directory for the provided project.
func (d *Dropbox) EntryDirForProject(p *Project) string {
	return d.ProjectDirPath(p) + "/entries"
}

// MkProjectDir ensures the project directory exists, ignoring the error when it already does.
func (d *Dropbox) MkProjectDir(p *Project) (err error) {
	if err = os.Mkdir(d.ProjectDirPath(p), 0755); os.IsExist(err) {
		err = nil
	}

	return
}

// Projects returns all discoverable projects under the Dropbox base path, logging unreadable entries.
func (d *Dropbox) Projects() []*Project {
	projects, err := d.repos.projects.All()
	if err != nil {
		fmt.Println("Failed to list projects:", err)
		return nil
	}
	return projects
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
