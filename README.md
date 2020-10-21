# repull

A tool to restart a Docker container with a newer version of an image used by the container
---

Often you may need to pull a newer version of an image to re-create an existing container, 
and you don't care much about the checkpoints because the application is stateless.

The usual way of doing that is to use 
```
docker pull repo/app
docker kill containerId
docker rm containerId 
docker run --name xyz --link -v ... -p ... etc
```

This is quite tedious and requires to remember all the command line options that were used to start the container. 
With `k8s` or `docker-compose` it is possible to solve that, but for vanilla Docker containers it could be tricky.

`repull` simplifies it to the point of `repull <containerid> <containername>...`.

It will

1. iterate over the list of passed container ids or names
2. fetch the description of the container
3. identify the image used in the container
4. lookup the information about the authentication in `~/.docker/config.json`
5. pull the new version of the image, if available for the current tag.
6. kill the existing container
7. remove the existing container
8. start the new container with the newest image available **preserving all configuration/startup options**
9. repeat for all images passed in the arguments.

### Usage
```
./repull -h
Usage of ./repull:
  -v    verbose

```
