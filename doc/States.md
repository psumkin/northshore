#State Machine

The State Machine handles the state processing and transitions for Stages in the Pipeline.

[Sequence Diagrams](https://goo.gl/ZSSrtP)

For the Iteration 1 we'll support the
[Docker Remote API](https://docs.docker.com/engine/reference/api/docker_remote_api).

![Docker Events](https://docs.docker.com/engine/reference/api/images/event_state.png)

And there is the [FSM for Go](https://github.com/looplab/fsm)


Running the Blueprint requires to run 3 commands - compose & script. This will yield 5 stages in a pipeline.


##State

    {
      "blueprint": {
        "name": "BP Name",
        "uuid": "G4G3G2G1-G6G5-G8G7-G9G10-G11G12G13G14G15G16"
        "stages": {
          "jenkins": {},
          "spinnaker": {},
          "gerrit": {},
          "artifactory": {}
        }
      }
    }


##Event

    {
      "event": {
        "blueprint": "G4G3G2G1-G6G5-G8G7-G9G10-G11G12G13G14G15G16",
        "service": "jenkins",
        "dockerStatus": (created|restarting|running|paused|exited),
        "datetime": "YYYY-MM-DD hh:mm:ss"
      }
    }


##Current State

  Combines _Events_ and produce the _State_
