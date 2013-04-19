package data

import (
	"code.google.com/p/goprotobuf/proto"
	"io/ioutil"
	"os"
)

// Saves Protocol Message to file
func Save(ps Persistable) error {
	_, err := os.Stat(ps.Location())
	if os.IsNotExist(err) {
		return write(ps)
	}

	return err
}

func Persisted(ps Persistable) bool {
	if Load(ps) != nil {
		return false
	}
	return true
}

func write(ps Persistable) error {
	data, err := proto.Marshal(ps)
	if err != nil {
		return err
	}

	// create dir if doesn't exists, etc
	err = ps.prepareForSave()
	if err != nil {
		return err
	}

	return ioutil.WriteFile(ps.Location(), data, 0666)
}

func Load(ps Persistable) error {
	return LoadFromPath(ps.Location(), ps)
}

func LoadFromPath(path string, ps Persistable) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	return proto.Unmarshal(data, ps)
}

func Destroy(ps Destroyable) error {
	err := os.Remove(ps.Location())
	if os.IsNotExist(err) {
		err = nil
	} else if err != nil {
		return err
	}
	err = ps.cleanupAfterRemove()
	return err
}
