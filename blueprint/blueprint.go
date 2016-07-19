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

package blueprint

import (
	"fmt"
	"io/ioutil"
	"strings"

	log "github.com/Sirupsen/logrus"

	"github.com/Mirantis/northshore/fsm"
	"github.com/Mirantis/northshore/store"
	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/satori/go.uuid"
	"github.com/spf13/viper"

	"golang.org/x/net/context"
)

// Stage represents a Blueprint Stage
type Stage struct {
	//Docker image for bootstrap stage
	Image       string `json:"image"`
	Description string `json:"description"`
	//Ports for exposing to host
	Ports []map[string]string `json:"ports"`
	//Environment variables
	Variables map[string]string `json:"variables"`
	//Provisioner type (docker/...)
	Provisioner string `json:"provisioner"`
	// State is current Blueprint status
	State StageState `json:"state"`
}

// Blueprint represents a Blueprint
type Blueprint struct {
	//API version for processing blueprint
	Version string `json:"version"`
	//Type of blueprint (pipeline/application)
	Type   string           `json:"type"`
	Name   string           `json:"name"`
	Stages map[string]Stage `json:"stages"`
	// State is current Blueprint status
	State State     `json:"state"`
	ID    uuid.UUID `json:"id"`
}

// State represents a state of the Blueprint
type State string

// StageState represents a state of the Stage
type StageState string

const (
	// StateNew is default state of the Blueprint
	StateNew State = "new"
	// StateProvision is the Blueprint status while provisioning
	StateProvision State = "provision"
	// StateActive is the Blueprint status when all Stages are up and ready
	StateActive State = "active"
	// StateInactive is the Blueprint status when some Stage is down
	StateInactive State = "inactive"
)

const (
	// StageStateNew is default state of the Stage
	StageStateNew StageState = "new"
	// StageStateCreated indicates that container is created
	StageStateCreated StageState = "created"
	// StageStateRunning indicates that container is running
	StageStateRunning StageState = "running"
	// StageStatePaused indicates that container is paused
	StageStatePaused StageState = "paused"
	// StageStateStopped indicates that container is stopped
	StageStateStopped StageState = "stopped"
	// StageStateDeleted indicates that container is deleted
	StageStateDeleted StageState = "deleted"
)

// DBBucket defines boltdb bucket for blueprints
const DBBucket = "blueprints"

func state(stagesStates map[string]StageState) State {
	bpState := StateNew

	for _, v := range stagesStates {
		if v == StageStateRunning {
			bpState = StateActive
			break
		}
	}
	for _, v := range stagesStates {
		if v == StageStateCreated {
			bpState = StateProvision
			break
		}
	}
LookInactive:
	for _, v := range stagesStates {
		switch v {
		case
			StageStateDeleted,
			StageStatePaused,
			StageStateStopped:
			bpState = StateInactive
			break LookInactive
		}
	}
	return bpState
}

// ParseFile parses and validates the incoming data
func ParseFile(path string) (bp Blueprint, err error) {
	bpv := viper.New()
	bpv.SetConfigFile(path)
	err = bpv.ReadInConfig()
	if err != nil {
		return bp, fmt.Errorf("Config not found. %s", err)
	}

	err = bpv.Unmarshal(&bp)
	if err != nil {
		return bp, fmt.Errorf("Unable to decode into struct, %v", err)
	}
	return bp, nil
}

// RunBlueprint creates and starts Docker containers
func RunBlueprint(bp Blueprint) {
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}
	ids := []string{}

	for name, stage := range bp.Stages {
		bindings := make(map[nat.Port][]nat.PortBinding)
		for _, ports := range stage.Ports {
			port, _ := nat.NewPort("tcp", ports["fromPort"])
			bindings[port] = []nat.PortBinding{nat.PortBinding{HostIP: "0.0.0.0", HostPort: ports["toPort"]}}
		}

		hostConfig := container.HostConfig{
			PortBindings: bindings,
		}

		config := container.Config{
			Image: bp.Stages[name].Image,
		}
		log.Printf("%s -> Config was built.", name)

		r, err := cli.ContainerCreate(context.Background(), &config, &hostConfig, nil, name)
		if err != nil && strings.Contains(err.Error(), "No such image") {
			log.Println(err)
			log.Println("Start pulling process...")
			rc, e := cli.ImagePull(context.Background(), config.Image, types.ImagePullOptions{})
			if e != nil {
				log.Println(e)
			}
			//TODO: add pretty print of pulling process
			_, re := ioutil.ReadAll(rc)
			if re != nil {
				log.Println(re)
			}
			rc.Close()
			r, err = cli.ContainerCreate(context.Background(), &config, &hostConfig, nil, name)
		}
		if err != nil {
			log.Println(err)
			continue
		}
		log.Printf("%s -> Container was created.", name)
		ids = append(ids, r.ID)

		err = cli.ContainerStart(
			context.Background(),
			r.ID,
			types.ContainerStartOptions{})
		if err != nil {
			log.Println(err)
			continue
		}
		log.Printf("%s -> Container was started.", name)
		log.Printf("%s -> Container ID  %s", name, r.ID)
		log.Printf("%s -> Warnings: %s", name, r.Warnings)
	}
	if len(ids) > 0 {
		//Update list of containers in DB
		//TODO add ability to add one container
		store.Save([]byte(fsm.DBBucketWatcher), []byte(fsm.DBKeyWatcher), strings.Join(ids[:], ","))
	}
}

// DeleteByID deletes blueprint with containers
func DeleteByID(id uuid.UUID) {
	// TODO: stop and remove containers
	log.Debugln("#blueprint,#DeleteBlueprint")

	store.Delete([]byte(DBBucket), []byte(id.String()))
}

// Save stores blueprint in db
func (bp *Blueprint) Save() error {
	ss := map[string]StageState{}
	for s := range bp.Stages {
		ss[s] = bp.Stages[s].State
	}
	bp.State = state(ss)
	zerouuid := uuid.UUID{}
	if bp.ID == zerouuid {
		bp.ID = uuid.NewV4()
	}
	return store.Save([]byte(DBBucket), []byte(bp.ID.String()), bp)
}

// Save stores blueprint in db
func Save(bp Blueprint) error {
	return bp.Save()
}
