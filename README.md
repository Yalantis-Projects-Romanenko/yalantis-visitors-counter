## Visitors count test project
This service uses websockets to notify all clients about visitors number update  
Before starting the service be sure that go version >= 1.11 
Service workdir should be the root directory of the project because of `static` dir that stores static content  
To get project dependencies run:  
```
go mod download
```
to start service run
```
go run server.go
```
Service will start on 8808 port. Visit http://localhost:8808 to check it out



### TODO
 - use cookies to ignore tab duplicates in browser
 - contextual logging to log request id
