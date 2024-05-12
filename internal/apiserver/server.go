package apiserver

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/remram44/vogon/internal/database"
	"github.com/remram44/vogon/internal/versioning"
)

type ApiServer struct {
	db database.Database
}

func runServer(config Config) error {
	db, err := config.Database.Connect()
	if err != nil {
		log.Fatalf("connecting to database: %v", err)
	}

	apiServer := ApiServer{
		db: db,
	}

	server := http.Server{
		Addr:    fmt.Sprintf("%v:%v", config.ListenAddr, config.ListenPort),
		Handler: &apiServer,
	}
	return server.ListenAndServe()
}

var pathFormat = regexp.MustCompile("^(/[a-z0-9][a-z0-9-]*)+$")

func sendJson(res http.ResponseWriter, status int, object interface{}) error {
	res.Header().Set("Content-type", "application/json")
	res.WriteHeader(status)
	encoder := json.NewEncoder(res)
	return encoder.Encode(object)
}

func sendMessage(res http.ResponseWriter, status int, message string) {
	type JsonMessage struct {
		Message string `json:"message"`
	}
	res.Header().Set("Content-type", "application/json")
	res.WriteHeader(status)
	encoder := json.NewEncoder(res)
	err := encoder.Encode(JsonMessage{Message: message})
	if err != nil {
		log.Printf("Error sending JSON message: %v", err)
	}
}

func boolParam(param string) (bool, error) {
	switch strings.ToLower(param) {
	case "1", "true", "yes":
		return true, nil
	case "0", "false", "no":
		return false, nil
	default:
		return false, fmt.Errorf("invalid boolean")
	}
}

func (s *ApiServer) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	log.WithFields(log.Fields{
		"remote_addr": req.RemoteAddr,
		"method":      req.Method,
		"path":        req.URL.Path,
	}).Print("request")

	if req.URL.Path == "/" {
		if req.Method == "GET" {
			res.Header().Set("Content-type", "text/plain")
			res.WriteHeader(200)
			io.WriteString(res, "Welcome to this Vogon server.\nSee https://github.com/remram44/vogon for information.\n")
		} else {
			res.WriteHeader(400)
		}
		return
	}

	if req.URL.Path == "/_version" && req.Method == "GET" {
		sendJson(res, 200, struct {
			Version string `json:"version"`
		}{
			Version: versioning.NameAndVersionString(),
		})
		return
	}

	if !pathFormat.MatchString(req.URL.Path) {
		sendMessage(res, 400, "Invalid path")
		return
	}

	name := req.URL.Path[1:]

	if req.Method == "GET" {
		object, err := s.db.Get(name)
		if err != nil {
			if _, ok := err.(*database.DoesNotExist); ok {
				sendMessage(res, 404, "No such object")
				return
			}

			log.WithField("name", name).Printf("GET: %v", err)
			sendMessage(res, 500, "error")
			return
		}

		err = sendJson(res, 200, object)
		if err != nil {
			log.WithField("name", name).Printf("GET send: %v", err)
		}
	} else if req.Method == "PUT" {
		create, err := boolParam(req.URL.Query().Get("create"))
		if err != nil {
			sendMessage(res, 400, fmt.Sprintf("invalid query parameter 'create'"))
			return
		}
		replace, err := boolParam(req.URL.Query().Get("replace"))
		if err != nil {
			sendMessage(res, 400, fmt.Sprintf("invalid query parameter 'replace'"))
			return
		}

		var object database.Object
		decoder := json.NewDecoder(req.Body)
		decoder.DisallowUnknownFields()
		err = decoder.Decode(&object)
		if err != nil {
			sendMessage(res, 400, fmt.Sprintf("error reading input: %v", err))
			return
		}
		if object.Metadata.Name != name {
			sendMessage(res, 400, "Mismatched name")
			return
		}
		var meta database.MetadataResponse
		if create && replace {
			meta, err = s.db.Create(object, true)
		} else if create {
			meta, err = s.db.Create(object, false)
		} else if replace {
			meta, err = s.db.Update(object)
		} else {
			sendMessage(res, 400, "Nothing to do if both create and replace are 0")
			return
		}
		if err != nil {
			sendMessage(res, 400, fmt.Sprintf("%v", err))
			return
		}
		sendJson(res, 200, meta)
	} else if req.Method == "DELETE" {
		id := req.URL.Query().Get("id")
		revision := req.URL.Query().Get("revision")
		meta, err := s.db.Delete(name, id, revision)
		if err != nil {
			status := 400
			if _, ok := err.(*database.DoesNotExist); ok {
				status = 404
			}
			sendMessage(res, status, fmt.Sprintf("%v", err))
			return
		}
		sendJson(res, 200, meta)
	}
}
