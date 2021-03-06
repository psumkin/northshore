// Copyright 2016 The NorthShore Authors All rights reserved.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"io/ioutil"
	"net/http"

	"github.com/Mirantis/northshore/blueprint"
	log "github.com/Sirupsen/logrus"
	"github.com/manyminds/api2go/jsonapi"
)

// BlueprintYAML represents form data
type BlueprintYAML struct {
	Data string `json:"data"`
}

//GetName implements jsonapi interface
func (BlueprintYAML) GetName() string {
	return "blueprintYAML"
}

//SetID implements jsonapi interface
func (*BlueprintYAML) SetID(string) error {
	return nil
}

// ParseBlueprintHandler creates and stores a blueprint
func ParseBlueprintHandler(w http.ResponseWriter, r *http.Request) {

	var body []byte
	var err error
	var yaml BlueprintYAML

	if body, err = ioutil.ReadAll(r.Body); err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Errorln("#ParseBlueprintHandler,#ReadRequestError")
		APIError(w, err, http.StatusInternalServerError)
		return
	}

	if err = jsonapi.Unmarshal(body, &yaml); err != nil {
		log.WithFields(log.Fields{
			"body": body,
			"err":  err,
		}).Errorln("#ParseBlueprintHandler,#UnmarshalBodyError")
		APIError(w, err, http.StatusBadRequest)
		return
	}

	bp, err := blueprint.ParseBytes([]byte(yaml.Data))
	if err != nil {
		log.WithFields(log.Fields{
			"yaml": yaml,
			"err":  err,
		}).Errorln("#ParseBlueprintHandler,#ParseBlueprintError")
		APIError(w, err, http.StatusBadRequest)
		return
	}

	log.WithFields(log.Fields{
		"body": string(body),
		"err":  err,
		"yaml": yaml,
		"bp":   bp,
	}).Debugln("#ParseBlueprintHandler")

	//TODO: Handle case if already exists
	if err := bp.Save(); err != nil {
		log.WithFields(log.Fields{
			"bp":  bp,
			"err": err,
		}).Errorln("#ParseBlueprintHandler")
		APIError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Location", "/api/v1/blueprints/"+bp.GetID())
	APIData(w, bp, http.StatusCreated)
}
