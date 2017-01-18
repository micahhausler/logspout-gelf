# Graylog GELF Module for Logspout
This module allows Logspout to send Docker logs in the GELF format to Graylog via UDP.

## Build
To build, you'll need to fork [Logspout](https://github.com/gliderlabs/logspout), add the following code to `modules.go` 

```
_ "github.com/micahhausler/logspout-gelf"
```
and run `docker build -t $(whoami)/logspout:gelf`

## Run

```
docker run \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -p 8000:80 \
    micahhausler/logspout:gelf \
    gelf://<graylog_host>:12201

```

## A note about GELF parameters
The following docker container attributes are mapped to the corresponding GELF extra attributes.

```
{
        "_container_id":   <container-id>,
        "_container_name": <container-name>,
        "_image_id":       <container-image-sha>,
        "_image_name":     <container-image-name>,
        "_command":        <container-cmd>,
        "_created":        <container-created-date>,
        "_swarm_node":     <host-if-running-on-swarm>
}
```

You can also add extra custom fields by adding labels to the containers.

for example 
a container with label ```gelf_service=servicename``` will have the extra field service



## License
MIT. See [License](LICENSE)
