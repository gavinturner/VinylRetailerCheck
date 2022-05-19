# vinylretailers
Checks selected vinyl retailers for new listings or prices. Keeps a running list of previous prices found
for artists of interest.

### postgres tweaks
* to get local postgres to accept connections from Docker containers, i had to enable postges to accept connections for outside the local host.
to achieve this, i had to:
  - edit **pg_hba.conf** and add the line: (where 192.168.0.201 is the ip of my laptop)
  > host    all    all  192.168.0.201/16        trust

  - I also had to edit **postgresql.conf** and set:
  > listen_addresses = '*'

### initial project setup (go mod)
* setup required so that the project recognises its root git repo for imports (using go.mod):
  > go mod init github.com/gavinturner/vinylretailers

* allow postgres to accept 'remote' connections (so Docker containers can access it)
  > https://blog.jsinh.in/how-to-enable-remote-access-to-postgresql-database-server/#.YoXL6pNByqC

### building for docker
* encapsulated in the Makefile:
  > **make docker** \
  > (builds and deploys the bridge network, redis and the scanner)

  > **make redis** \
  > (builds and deploys redis only)
  
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

### docker run - useful cli options
https://docs.docker.com/engine/reference/commandline/run

### docker network create - useful cli options
https://docs.docker.com/engine/reference/commandline/network_create

### docker useful network commands

- list the networks available:
  > docker network ls
- inspects a specific network definition:
  > docker network inspect <network name>
- removes a specific network definition:
  > docker network rm <network name>
- get the ip address of a specific container:
  > docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' <container name>

### user-defined docker network: vinylretailers
We have created a specific user-defined bridge network 'vinylretailers' to use. This enables the containers can communicate 
between them, using the container name as the hostname. Also, the host can communicate to published/exported ports using localhost.
This was a pain to get working! To get this to work the network host binding needed to be set to '0.0.0.0', and the
enable_ip_masquerade setting needed to be set to 'true'. I also set '--expose=6379' on the redis container, so that may have been
required as well (should test this).
