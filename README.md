# netjailbreak
It solves a problem of connecting from a docker container to the outside world when other ways are unavailable.
It is made of two parts:
## internal 
You run this process inside a container.
- it listens on a tcp port and acts as a proxy to a network resource not accessible from the inside;  
- it exposes a tcp port to the outside world. "external" part uses it to poll data.
## external
You run this process outside of a container.
- it polls the internal part for data and sends it to the remote host
- it accepts data from the remote hosts and sends it to the internal part
# usage
```bash
SERVER_ADDRESS=localhost;SERVER_PORT=3333;API_URL=localhost:9090 ./internal
REMOTE_HOST=touk.pl;REMOTE_PORT=443;API_URL=http://localhost:9090 ./external
```
# docker image
There is a docker image for the internal part. It can be used in a docker compose to act as a gateway.
# executables
Are stored in the repository.
```bash
go build -ldflags "-linkmode external -extldflags -static" -x
```
# disclaimer
This is a two-evenings project. My first program in Golang. There are no external dependencies. 
Polling makes connections 5 times slower. Improvements are welcomed but please keep it simple.
