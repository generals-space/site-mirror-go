windows

```
docker run -it --name site-mirror-go -v %gopath%/src/gitee.com/generals-space/site-mirror-go.git:/project generals/golang_node8 /bin/bash
```

linux 

```
docker run -it --name site-mirror-go -v $GOPATH/src/gitee.com/generals-space/site-mirror-go.git:/project generals/golang_node8 /bin/bash
```
