# vinylretailers
Checks selected vinyl retailers for new listings or prices. Keeps a running list of previous prices found
for artists of interest.

## running in Docker
### with postgres running on the host
* to get a local (host) version of postgres to accept connections from Docker containers, i had to enable postges
to accept connections for outside the local host. 
* allowing postgres to accept 'remote' connections (so Docker containers can access it) is described here: 
https://blog.jsinh.in/how-to-enable-remote-access-to-postgresql-database-server/#.YoXL6pNByqC
* So to achieve this, i had to:
  - edit **pg_hba.conf** and add the line: (where 192.168.0.201 is the ip of my laptop)
  > host    all    all  192.168.0.201/16        trust

  - I also had to edit **postgresql.conf** and set:
  > listen_addresses = '*'

### with postgres running in a container
* when you run in a container you need to be mindful of the ports that are being exposed.
  - we're running with exposure settings: **-p 5400:5432**
  - this means the port is **5432** within docker (between containers)
  - and the port exposed to the local host is **5400**
* also when communicating between containers using a user-defined bridge network, the hostname for the 
postgres db is the container-name (in our case **vinylretailers-postgres**).

### using a bridge network between containers
The use of a bridging network is described here: https://docs.docker.com/network/network-tutorial-standalone. 
* Using a bridge network allows communication between containers 
* We use the flag **--network='bridge-name'** on each call to docker run to attach a container to the bridge.
* When we create a container that exposes a port, we use the flag **-p=PORTA:PORTB** on the call to docker run where 
  - PORTA: is the port exposed to the local host
  - PORTB: is the port exposed within the conytainer bridge network
* To get the containers on the bridge network to be visible from the local host, you need to use the following options:
> docker network create --driver bridge \ \
> -o "com.docker.network.bridge.host_binding_ipv4"="0.0.0.0" \ \
> -o "com.docker.network.bridge.enable_ip_masquerade"="true" \ \
> -o "com.docker.network.bridge.enable_icc"="true" <network-name>
* Using this setup the host should be able to see ports exposed on containers using localhost:PORTA

### docker useful network commands
- list the networks available:
  > docker network ls
- inspects a specific network definition:
  > docker network inspect <network name>
- removes a specific network definition:
  > docker network rm <network name>
- get the ip address of a specific container:
  > docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' <container name>

### docker network create - useful cli options
https://docs.docker.com/engine/reference/commandline/network_create


### user-defined docker network: vinylretailers
We have created a specific user-defined bridge network 'vinylretailers' to use. This enables the containers can communicate
between them, using the container name as the hostname. Also, the host can communicate to published/exported ports using localhost.
This was a pain to get working! To get this to work the network host binding needed to be set to '0.0.0.0', and the
enable_ip_masquerade setting needed to be set to 'true'. I also set '--expose=6379' on the redis container, so that may have been
required as well (should test this).

### docker run - useful cli options
https://docs.docker.com/engine/reference/commandline/run


## building and running for Docker
* encapsulated in the Makefile:
  > **make docker** \
  > (builds and deploys the bridge network, redis, postgres and the scanner)

  > **make redis** \
  > (builds and deploys redis only)

  > **make postgres** \
  > (builds and deploys postgres only using a local volume)

  > **make pgadmin** \
  > (builds and deploys postgres web ui only)

  > **make scanner** \
  > (builds and deploys the scanner only)

* create a bridge network for images to communicate
  > docker network create --driver bridge \ \
  > -o "com.docker.network.bridge.host_binding_ipv4"="0.0.0.0" \ \
  > -o "com.docker.network.bridge.enable_ip_masquerade"="true" \ \
  > -o "com.docker.network.bridge.enable_icc"="true" vinylretailers

* grab and run the redis image without persistent data storage (port 6379) and using the vinylretailers network bridge:
  > docker pull redis:latest \
  > docker run --name vinylretailers-redis --network=vinylretailers -p 6379:6379 --expose=6379 -d redis

* build and run the scanner docker image using the vinylretailers network bridge:
  > docker build -t vinylretailers-scanner -f cmd/scanner/Dockerfile.scanner . \
  > docker run run --name vinylretailers-scanner --network=vinylretailers -d vinylretailers-scanner
  

### golang initial project setup (go mod)
* setup required so that the golang project recognises its root git repo for imports (ie using go.mod):
  > go mod init github.com/gavinturner/vinylretailers
  
### logging into the postgres pgAdmin web interface
* use username/password "vinylretailers"
* use the hostname "vinylretailers-postgres"
* the port is the default 5432
