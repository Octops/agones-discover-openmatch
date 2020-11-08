# Agones Allocator Service

If you are not familiar with the GameServerAllocation concept, check the documentation from the official Agones website.

- https://agones.dev/site/docs/reference/gameserverallocation/

>A GameServerAllocation is used to atomically allocate a GameServer out of a set of GameServers. This could be a single Fleet, multiple Fleets, or a self managed group of GameServers.

- https://agones.dev/site/docs/advanced/allocator-service/

> Agones provides an mTLS based allocator service that is accessible from outside the cluster using a load balancer. The service is deployed and scales independent to Agones controller.
  To allocate a game server, Agones in addition to GameServerAllocations , provides a gRPC service with mTLS authentication, called agones-allocator.

## Allocation Rules using the Agones Allocator Service

If you are running the director using `mode=agones`, the allocation rules will be slightly different.
The criteria used for allocation will be based on the labels `region` and `world`. Whenever a gameserver is `Ready` it is eligible for allocation. Therefore, the player capacity is going to be ignored.
 
## Certificates

The most important part of the setup when using the Agones Allocator service is the correct configuration of the certificates.

There is a very detailed explanation of how this can be setup on the Agones website https://agones.dev/site/docs/advanced/allocator-service/.

Check the [hack/get_certificates.sh](/hack/get_certificates.sh) script if you have installed Agones using Helm

## Update Manifests

Update the [/deploy/install.yaml](/deploy/install.yaml) to reflect the role you want the director to assume. By default the Director will use the Octops Discover service. You will find a commented block with proper instructions.

Running the Director and using the Agones Allocator service require the following manifest modifications:

Set the proper container arguments
```yaml
# Adjust the indentation if required for your manifest
  - image: octops/agones-openmatch:latest
    name: director
    args:
    - --mode=agones
    - --interval=1s
    - --key=/tls/crt/tls.key
    - --cert=/tls/crt/tls.crt
    - --cacert=/tls/ca/tls-ca.crt
    - --allocator-host=192.168.0.110 # IP of the exposed Agones Allocator Service
    - --allocator-port=30304 # PORT of the exposed Agones Allocator Service
    - --verbose
```

Mount the volumes storing the TLS information. Check the [hack/get_certificates.sh](/hack/get_certificates.sh) script if you have installed Agones using Helm.

```yaml
# Adjust the indentation if required for your manifest 
      volumeMounts:
        - name: tls-crt
          mountPath: /tls/crt
        - name: tls-ca
          mountPath: /tls/ca
  volumes:
    - name: tls-crt
      secret:
        secretName: allocator-tls-crt
    - name: tls-ca
      secret:
        secretName: allocator-tls-ca
```

## Fleet Autoscaling

If you have enough capacity on the environment you are running your game servers, you may consider [Fleet Autoscaling](https://agones.dev/site/docs/reference/fleetautoscaler/).

## Demo

This is a five minutes demo that shows all the matchmaking components interacting together and matches allocated to game servers using the [Agones Allocator Service](https://agones.dev/site/docs/advanced/allocator-service/).

Video: https://youtu.be/UkUP44QgfFE