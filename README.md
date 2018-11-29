# Kubernetes Event Alert


## What is this?

- listwatch kubernetes event resource
- send events to influxdb and rabbitmq
- send alert for important event
 
## Running

Clone repo:

```
go build -o ealert .
```

Prepare build environment:

```
configs/kubepub.conf
configs/default.yaml
```

Run:
```
./ealert 
```
