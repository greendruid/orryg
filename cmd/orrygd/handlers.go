package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/vrischmann/orryg"
)

func writeJSON(w http.ResponseWriter, data []byte) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_, err := w.Write(data)

	return err
}

func writeString(w http.ResponseWriter, data string) error {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)

	_, err := w.Write([]byte(data))

	return err
}

func marshalAndWriteJSON(w http.ResponseWriter, val interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	enc := json.NewEncoder(w)
	return enc.Encode(val)
}

func writeError(w http.ResponseWriter, format string, args ...interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_, err := fmt.Fprintf(w, format, args...)

	return err
}

func handleCopiersList(w http.ResponseWriter, req *http.Request) error {
	scpConfs, err := store.getAllSCPCopierConfs()
	if err != nil {
		return err
	}

	var res orryg.CopiersConf
	for _, el := range scpConfs {
		res = append(res, orryg.CopierConf{
			Type: orryg.SCPCopierType,
			Conf: el,
		})
	}

	return marshalAndWriteJSON(w, res)
}

func handleCopiersAdd(w http.ResponseWriter, req *http.Request) error {
	defer req.Body.Close()
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err
	}

	var body orryg.UCopierConf
	err = json.Unmarshal(data, &body)
	if err != nil {
		return err
	}

	switch body.Type {
	case orryg.SCPCopierType:
		var scpConf orryg.SCPCopierConf
		if err = json.Unmarshal(body.Conf, &scpConf); err != nil {
			return err
		}

		if err = store.mergeSCPCopierConf(scpConf); err != nil {
			return err
		}
	}

	return writeString(w, "OK")
}

func handleCopiersRemove(w http.ResponseWriter, req *http.Request) error {
	name := req.URL.Path[len("/copiers/remove/"):]
	if name == "" {
		return errors.New("no name defined")
	}

	if err := store.removeCopier(name); err != nil {
		return err
	}

	return writeString(w, "OK")
}

func handleDirectoriesList(w http.ResponseWriter, req *http.Request) error {
	res, err := store.getDirectories()
	if err != nil {
		return err
	}

	return marshalAndWriteJSON(w, res)
}

func handleDirectoriesAdd(w http.ResponseWriter, req *http.Request) error {
	defer req.Body.Close()
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err
	}

	var d orryg.Directory
	err = json.Unmarshal(data, &d)
	if err != nil {
		return err
	}

	if err = store.mergeDirectory(d); err != nil {
		return err
	}

	// Force the first backup
	e.oodCh <- d

	return writeString(w, "OK")
}

func handleDirectoriesRemove(w http.ResponseWriter, req *http.Request) error {
	name := req.URL.Path[len("/directories/remove/"):]
	if name == "" {
		return errors.New("no name defined")
	}

	if err := store.removeDirectory(name); err != nil {
		return err
	}

	return writeString(w, "OK")
}
