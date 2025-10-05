package data

import (
	"errors"
	"fmt"
	"os"
	"sort"

	"google.golang.org/protobuf/proto"
)

type ProjectRepository interface {
	All() ([]*Project, error)
}

type EntryRepository interface {
	ForProject(*Project) ([]*Entry, error)
}

type repositorySet struct {
	projects ProjectRepository
	entries  EntryRepository
}

func newFilesystemRepositories(d *Dropbox, store *Store) repositorySet {
	return repositorySet{
		projects: &filesystemProjectRepository{
			store:   NewGenericStore[*Project](store),
			baseDir: func() string { return d.BasePath },
		},
		entries: &filesystemEntryRepository{
			store:    NewGenericStore[*Entry](store),
			entryDir: func(p *Project) string { return d.EntryDirForProject(p) },
		},
	}
}

type filesystemProjectRepository struct {
	store   GenericStore[*Project]
	baseDir func() string
}

// All returns every project directory beneath the configured Dropbox base path.
// It tolerates unreadable projects by logging and continuing, mirroring legacy behaviour.
func (r *filesystemProjectRepository) All() ([]*Project, error) {
	items, err := r.store.store.list(r.baseDir())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	projects := make([]*Project, 0, len(items))
	for _, item := range items {
		if !item.IsDir {
			continue
		}
		project := new(Project)
		project.Sha = proto.String(item.Name)
		if err := r.store.Load(project); err != nil {
			// Skip unreadable projects but keep going.
			fmt.Println("Failed:", item.Name, ":", err)
			continue
		}
		projects = append(projects, project)
	}
	return projects, nil
}

type filesystemEntryRepository struct {
	store    GenericStore[*Entry]
	entryDir func(*Project) string
}

// ForProject lists and loads all entries associated with the given project, newest first.
// Entries that cannot be decoded are skipped so partial histories still render.
func (r *filesystemEntryRepository) ForProject(project *Project) ([]*Entry, error) {
	items, err := r.store.store.list(r.entryDir(project))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	entries := make([]*Entry, 0, len(items))
	for _, item := range items {
		if item.IsDir || item.Name == ".DS_Store" {
			continue
		}
		entry := new(Entry)
		entry.Project = project
		entry.Id = proto.String(item.Name)
		if err := r.store.Load(entry); err != nil {
			fmt.Println("Failed:", item.Name, ":", err)
			continue
		}
		entry.GetEnded()
		entries = append(entries, entry)
	}
	sort.Sort(byEndTime(entries))
	return entries, nil
}
