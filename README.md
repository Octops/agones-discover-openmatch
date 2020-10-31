# Matchmaking System using Agones and Open Match

This project aims to be an example of how a matchmaking system using Agones and Open Match could be implemented. It is heavily inspired by https://open-match.dev/site/docs/tutorials.

The matchmaking logic and allocation process assume a hypothetical game where players will be matched together based on their skills, latency, region, the desired world they want to play and the gameserver capacity. 

This project does not target any game, company or specific audience. It has been mostly the result of a learning process that put together different tools and technologies.

Real games require way more input information, and a more complex matchmaking algorithm. However, the approach followed in here may be a good starting point.      

### Agones
> An open source, batteries-included, multiplayer dedicated game server scaling and orchestration platform that can run anywhere Kubernetes can run.

### Open Match
> Open Match is a flexible match making system built to scale with your game.

You can find great documentation about each one of these projects on their websites.
- https://agones.dev/site/
- https://open-match.dev/site/

## Requirements
If you want to see the project in action, please make sure you have the requirements set. Basically you need a Kubernetes running Agones 1.9.0 and Open Match 1.0.0.

### Kubernetes cluster

This project will work on any Kubernetes cluster that is supported by Agones 1.9.0 (K8S 1.16). That could be a cluster running in the cloud or locally.

The development and testing has been done using a Kubernetes cluster provisioned with [k3s](https://k3s.io/) on a separated Linux server. 

However, the exactly same results could be achieved running on [Minikube](https://github.com/kubernetes/minikube), [Kind](https://kind.sigs.k8s.io/) or any cloud environment.

Running a Kubernetes cluster provisioned with k3s (Linux only):
```bash
export INSTALL_K3S_VERSION=v1.16.15+k3s1
curl -sfL https://get.k3s.io | sh -s - --docker
# Kubeconfig is written to /etc/rancher/k3s/k3s.yaml
```

### Agones

This project has been developed and tested using Agones 1.9.0. I can't guarantee compatibility to any previous version since it has not been tested.

Although the documentation covers the deployment on a single cluster, the solution can be deployed on a multi cluster topology. The documentation for that scenario will be added soon.

The [PlayerTracking](https://agones.dev/site/docs/guides/player-tracking/) will provide the information about Players (Capacity and Count) which is used by the Allocator when looking for available gameservers. Different matchmaking systems might not require that Agones feature. 
- Enable Player Capacity feature gate: 
```bash
$ helm install agones --namespace agones-system --create-namespace agones/agones --set agones.featureGates=PlayerTracking=true
```

### Open Match

The manifests to install Open Match 1.0.0 can be found on [deploy/openmatch](deploy/openmatch) folder, they also include Prometheus and Grafana. 

These files have been downloaded from the official Open Match repo. The only difference is the number of replicas set to 1 for the core components. 

To install all the Open Match components run the following command:
```bash
$ ./deploy/openmatch/deploy.sh
```

You can also check the official Open Match installation docs on https://open-match.dev/site/docs/installation/.

## Development

Below you can find details about the local setup used for the development and testing of this project.

- Dedicated machine for coding the required matchmaking components and extras
- Local Kubernetes cluster running Agones and Open Match - Dedicated Linux Server
- Player simulator, MMF and Director running locally but interacting with Open Match services deployed on Kubernetes
- Services exposed via `NodePort` but the same could be achieved using `kubectl port-forward`

During the development of this project I've exposed some Open Match services using services of type NodePort. That eliminates the need of using `port-forward` to communicate with those services running within the cluster.
```bash
# Check the manifest and update the nodePort field if required. 
$ kubectl -n open-match apply -f open-match-services-nodeport.yaml
```  

If you are a [direnv](https://direnv.net/) user check the [.envrc.template](.envrc.template) file.

## Matchmaking Components
Below there is a list of all the services and components that put together deliver the match making system. They can be part of this repo, Open Match built in or third party services.

- Repo
    - Player Simulator
    - Game FrontEnd
    - Match Making Function (a.k.a MMF)
    - Director and Allocator

- Open Match
    - Builtin: Backend, Frontend, Query Service, Evaluator, Synchronizer

- Third Party
    - Octops/Discover

## Hypothetical Game Scenario

The scenario considered for the matchmaking takes into consideration the following input:
 
- GameSession Open world: Players can join and leave a match at any time 
- Each Fleet deploys GameServers hosting one specific "World" and "Region":
    - Fleet: us-east-1: Dune
    - Fleet: us-east-2: Nova
    - Fleet: us-west-1: Pandora
    - Fleet: us-west-2: Orion 
- Players can join if a GameServer still has enough free capacity (Players.Capacity - Players.Count). Check the [PlayerTracking](https://agones.dev/site/docs/guides/player-tracking/) feature.
- Player's tickets will be grouped by Region, World, Skill and Latency. Skill and Latency are range based.
- Game Client provides the desired match and player's info: Region, World, Skill and Latency. There is nothing crazy going on here. The Player simulator just pick up a fake latency from a range and assign to the ticket.

## Match Making Rules
 
- Ticket Details: created by the Player simulator and pushed to the Game Service Frontend
    - Game Mode Tag: mode.session
    - Region: us-east-1, us-east-2, us-west-1, us-west-2
    - World: Dune, Nova, Pandora, Orion
    - Skill Level: 1-1000
    - Latency: 0-100  

- MMF Criteria
    - Open Match should create PoolTickets based on the above criteria 
    - Player capacity: 10 Tickets/Players per match

- Director Profiles
    - Every 5s (interval flag) the [director](pkg/director/openmatch) will generate profiles and request matches
    - Skill and Latency are range based.

## Allocation Rules
    
The allocation service will try to find a GameServer that matches the criteria found on the `Extension` field of the `AssignTicketsRequest`. This information must match with Labels (Region and World) from the Fleet and GameServer.

Rules: 
   - GameServer running on the desired Region
   - GameServer hosting the desired World
   - GameServer Capacity can accommodate all the tickets from the match
   - Assign connections to matches or clean up those not allocated

## Looking for GameServers using the Octops Discover Service

The Octops Discover service works like a central GameServers state store. 

One of the core components of the Octops Discover is the https://github.com/Octops/agones-event-broadcaster together with a data store and an HTTP API to expose the GameServers states.

The director leverages the searching of the GameServers to the Octops/Discover service.

This service exposes an HTTP API to be consumed by clients looking for GameServer information using some filtering. That includes labels or any [GameServer](https://github.com/googleforgames/agones/blob/master/pkg/apis/agones/v1/gameserver.go) field present on the data struct. 

**More details about the Octops Discover project and how it could be used for multi cluster topology will be available soon.**

The manifest to deploy the service can be found on [deploy/third-party/octops.yaml](deploy/third-party/octops.yaml).

An example of how to query for GameServers is shown below. The client requests the service endpoint passing the right query params.
```bash
# Decoded Version
GET http://octops-discover.agones-openmatch.svc.cluster.local:8081/api/v1/gameservers?fields=status.state=Ready&labels=region=us-east-1,world=Dune

# Encoded Version
GET http://octops-discover.agones-openmatch.svc.cluster.local:8081/api/v1/gameservers?fields=status.state%3DReady&labels=region%3Dus-east-1%2Cworld%3DDune
``` 

Example of a response:

```json
{
    "data": [
        {
            "uid": "6cb220c0-ca9d-4164-bca9-a8aa1a879fd3",
            "name": "fleet-us-east-1-dune-8g8dl-vw9l7",
            "namespace": "default",
            "resource_version": "161047",
            "labels": {
                "agones.dev/fleet": "fleet-us-east-1-dune",
                "agones.dev/gameserverset": "fleet-us-east-1-dune-8g8dl",
                "region": "us-east-1",
                "world": "Dune"
            },
            "status": {
                "state": "Ready",
                "address": "192.168.0.110:7019",
                "players": {
                    "count": 0,
                    "capacity": 10,
                    "ids": null
                }
            }
        }
    ]
}
```

The Allocator service will use the list of GameServers from the response and assign a connection to ticket based on the GameServer capacity. If any GameServer can accommodate the ticket, the connection will not be assigned.

Alternatively, the director could be extended and use any other sort of allocation mechanism. However, this is not covered by this project due to the game use case presented previously. Future work may include different games scenarios and will explore alternatives to the Octops Discover.

You can find more details of how to allocate GameServers using the Agones Allocation Service on https://agones.dev/site/docs/advanced/allocator-service/.

## Running
All the components will be built into a single binary. The passing argument when running the binary together with a few flags will start the proper process.

Make sure you have the expected `environment variabels` set and pointing to valid endpoints. Check the [.envrc.template](.envrc.template) file for references. 

**Important**
Running this project requires the Octops Discover service. Details about this project and service will be added soon.

* Deploy Octops Discover Service
```bash
$ kubectl -n agones-openmatch apply -f deploy/third-party/octops.yaml

# port-forward - This endpoint will be used by the Director when allocating gameservers
$ kubectl -n agones-openmatch port-forward svc/octops-discover 8081
```

Player Simulator
```bash
# Generate 10 random profiles/players every 5 seconds
go run main.go player simulate --players-pool 10 --interval 5s
```

Matchmaking Function - MMF
```bash
$ go run main.go function --verbose
```

Director
```bash
# Generate profiles and Fetch matches every 5 seconds
$ go run main.go director --interval 5s --verbose
```

## Install

The steps below covers the deployment of the whole solution to a Kubernetes cluster.

* Create namespace
```bash
$ kubectl create ns agones-openmatch
```

* Deploy Octops Discover Service
```bash
$ kubectl -n agones-openmatch apply -f deploy/third-party/octops.yaml
```

* Deploy Matchmaking components: MMF, Director and Players Simulator
```bash
# Player simulator replicas=0 check section below
$ kubectl -n agones-openmatch apply -f deploy/install.yaml
```

* Deploy Fleets
```bash
# Default namespace
$ kubectl apply -f demo/fleets/fleets.yaml 
```

## Matchmaking

The Player Simulator will generate 10 tickets every 5 seconds. The details of the ticket will be randomly assigned: World, Region, Skill and Latency. 

Scale Players Simulator and check logs
```bash
$ kubectl -n agones-openmatch scale deployment agones-openmatch-players --replicas=1
$ kubectl -n agones-openmatch logs -f $(kubectl -n agones-openmatch get pods -l app=agones-openmatch-players -o jsonpath="{.items[0].metadata.name}")
```

Output:
```bash
DEBU[0094] ticketID=buaq0f9m0k8lmh52dp2g playerUID=1ac50018-4ae3-41eb-9540-91eec508fe6e stringArgs=map[region:us-west-2 world:Nova] doubleArgs=map[latency:25 skill:10]
DEBU[0094] ticketID=buaq0f9m0k8lmh52dp30 playerUID=2ee0179d-d0d1-4a63-aa09-e110255a46eb stringArgs=map[region:us-east-2 world:Pandora] doubleArgs=map[latency:25 skill:1000]
DEBU[0094] ticketID=buaq0f9m0k8lmh52dp3g playerUID=26bc9990-b5c1-4b6b-a3d2-7997e7c6244a stringArgs=map[region:us-east-1 world:Orion] doubleArgs=map[latency:50 skill:100]
DEBU[0094] ticketID=buaq0f9m0k8lmh52dp40 playerUID=36f07d2e-565f-4430-aa3f-efae59f5220e stringArgs=map[region:us-east-1 world:Orion] doubleArgs=map[latency:75 skill:10]
```

Check logs from the Director
```bash
kubectl -n agones-openmatch logs -f $(kubectl -n agones-openmatch get pods -l app=agones-openmatch-director -o jsonpath="{.items[0].metadata.name}")
```

Output:
```bash
time="2020-10-27T18:34:03Z" level=info msg="fetching matches for profile world_based_profile_Dune_us-east-2" command=fetch component=director
time="2020-10-27T18:34:03Z" level=info msg="fetching matches for profile world_based_profile_Nova_us-east-2" command=fetch component=director
time="2020-10-27T18:34:03Z" level=info msg="fetching matches for profile world_based_profile_Dune_us-west-1" command=fetch component=director
```

The logs from the Director will show 2 possible situations:
- A match has been created but could not be assigned because a GameServer could not be found (based on the criteria region and world), or it does not have capacity to accommodate all the Players from the Ticket.
```bash
# There is no Fleet hosting Orion on us-east-1
time="2020-10-25T17:10:39Z" level=debug msg="gameservers not found for request with filter map[fields:status.state=Ready labels:region=us-east-1,world=Orion]" component=allocator
```

- A GameServer matches with the match criteria (world, region and capacity) and the connection has been assigned to the tickets.
```bash
# GameServer hosting Orion on us-west-2 assigned to 8 Players
time="2020-10-25T17:09:14Z" level=info msg="gameserver fleet-us-west-2-orion-q42z5-8sjff connection 192.168.0.10:7015 assigned to request, total tickets: 8" component=allocator
```

Check logs from the MMF
```bash
kubectl -n agones-openmatch logs -f $(kubectl -n agones-openmatch get pods -l app=agones-openmatch-mmf -o jsonpath="{.items[0].metadata.name}")
```

Output:
```bash
time="2020-10-27T18:35:23Z" level=debug msg="creating match for ticket buc6f80rcraqq57mlhe0" command=matchmaker component=match_function
time="2020-10-27T18:35:23Z" level=debug msg="total matches for profile world_based_profile_Orion_us-east-1: 1" command=matchmaker component=match_function
time="2020-10-27T18:35:23Z" level=debug msg="total matches for profile world_based_profile_Dune_us-west-1: 0" command=matchmaker component=match_function
time="2020-10-27T18:35:23Z" level=debug msg="creating match for ticket buc6f98rcraqq57mlhhg" command=matchmaker component=match_function
time="2020-10-27T18:35:23Z" level=debug msg="total matches for profile world_based_profile_Nova_us-west-2: 1" command=matchmaker component=match_function
```

Open Match Dashboards

Check metrics from Open Match services using the Grafana dashboards deployed together with the other components.

```bash
$ kubectl -n open-match port-forward svc/open-match-grafana 3000
# Username: Admin Password: openmatch
```

## Roadmap

- [ ] Improve test cases and coverage
- [ ] Implement Allocator using Agones Allocation Service via gRPC
- [ ] Instrument with Prometheus
- [ ] Explore different MMF logics
- [ ] Record demo
- [ ] Document multi cluster deployment
- [ ] Extract Profile Generator to a service
- [ ] Extract Allocator to a service
- [ ] Delete assigned tickets - Player Simulator
- [ ] Add CI

## Contributions or Feedbacks

Issues and pull requests are welcome.

The project's goal is not to be a plug and play matchmaking system that works for every single use case. Instead it could inspire other folks who are looking for some examples which use the same or similar stack.

Some components could be easily extracted or extended in a way that it would cover different uses cases.

You can reach out on Agones or Open Match Slack if you have any question.
