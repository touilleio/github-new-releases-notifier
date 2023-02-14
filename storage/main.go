package storage

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
)

type Handler struct {
	dbStoragePath string
	dbBucketName  string
	bucketName    []byte
	db            *bolt.DB
	log           *logrus.Logger
}

var (
	dummyPayload = []byte("42")
)

func NewStorageHandler(dbStoragePath string, dbBucketName string, log *logrus.Logger) (*Handler, error) {

	db, err := bolt.Open(dbStoragePath, 0666, nil)
	if err != nil {
		return nil, err
	}

	bucketName := []byte(dbBucketName)
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketName)
		return err
	})
	if err != nil {
		return nil, err
	}

	handler := &Handler{
		log:        log,
		bucketName: bucketName,
		db:         db,
	}

	return handler, nil
}

type RepoAndTag struct {
	Repo string `json:"repo"`
	Tag  string `json:"tag"`
}

func (o RepoAndTag) String() string {
	return o.Repo + "|" + o.Tag
}

func repoAndKeyToString(repo string, tag string) string {
	return repo + "|" + tag
}

func repoAndKeyFromString(repoAndTagStr string) RepoAndTag {
	rat := strings.Split(repoAndTagStr, "|")
	repoAndTag := RepoAndTag{
		Repo: rat[0],
		Tag:  rat[1],
	}
	return repoAndTag
}

func (h *Handler) TagExists(repo string, tag string) (bool, error) {
	var fileExists = false
	err := h.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(h.bucketName)
		v := b.Get([]byte(repoAndKeyToString(repo, tag)))
		fileExists = v != nil
		return nil
	})
	return fileExists, err
}

func (h *Handler) StoreTag(repo string, tag string) error {
	err := h.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(h.bucketName)
		err := b.Put([]byte(repoAndKeyToString(repo, tag)), dummyPayload)
		return err
	})
	return err
}

func (h *Handler) RemoveTag(repo string, tag string) error {
	err := h.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(h.bucketName)
		err := b.Delete([]byte(repoAndKeyToString(repo, tag)))
		return err
	})
	return err
}

func (h *Handler) ListTags() ([]RepoAndTag, error) {
	tags := make([]RepoAndTag, 0, 64)
	err := h.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(h.bucketName)
		c := b.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			tags = append(tags, repoAndKeyFromString(string(k)))
		}
		return nil
	})
	return tags, err
}

func (h *Handler) Close() {
	if h.db != nil {
		h.db.Close()
	}
}

// curl -X PUT -d '{"repo": "https://github.com/x/y", "tag":"x.v.z"}' http://.../delete (and hence re-process the given file.
func DeleteTagOnDemand(storageHandler *Handler) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodPut {
			if req.Body == nil {
				http.Error(w, "Ignoring request. Required non-empty request body.\n", http.StatusBadRequest)
				return
			}
			defer req.Body.Close()
			body, err := ioutil.ReadAll(req.Body)
			if err != nil {
				http.Error(w, fmt.Sprintf("Got an error reading body, %v\n", err), http.StatusBadRequest)
				return
			}
			var repoAndTag RepoAndTag
			err = json.Unmarshal(body, &repoAndTag)
			if err != nil {
				http.Error(w, fmt.Sprintf("Got an error unmarshalling body %s, %v\n", string(body), err), http.StatusBadRequest)
				return
			}
			storageHandler.RemoveTag(repoAndTag.Repo, repoAndTag.Tag)
			http.Error(w, "Ok\n", http.StatusOK)
		} else {
			http.Error(w, "Use PUT method instead\n", http.StatusMethodNotAllowed)
		}
	}
}

// curl -X GET http://... list all the stored files.
func ListAllNotifiedTags(storageHandler *Handler) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodGet {
			files, err := storageHandler.ListTags()
			if err != nil {
				storageHandler.log.Warnf("Got an error while listing all the files, %v", err)
				http.Error(w, fmt.Sprintf("Error %v\n", err), http.StatusInternalServerError)
			} else {
				b, err := json.Marshal(files)
				if err != nil {
					storageHandler.log.Warnf("Got an error while marshaling the files, %v", err)
					http.Error(w, fmt.Sprintf("Error %v\n", err), http.StatusInternalServerError)
				} else {
					w.Header().Set("Content-Type", "application/json; charset=utf-8")
					w.WriteHeader(http.StatusOK)
					_, err = w.Write(b)
					if err != nil {
						storageHandler.log.Warnf("Got an error while writing the reponse, %v", err)
					}
				}
			}
		} else {
			http.Error(w, "Use PUT method instead\n", http.StatusMethodNotAllowed)
		}
	}
}
