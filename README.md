# grpc-todo-list
https://medium.com/@amsokol.com/tutorial-how-to-develop-go-grpc-microservice-with-http-rest-endpoint-middleware-kubernetes-daebb36a97e9

## Add import path to `api/proto/v1/todo-service.proto``:
```protobuf
option go_package = "./";
```

## Generate go code from proto
```sh
protoc --proto_path=api/proto/v1 --proto_path=third_party --go_out=plugins=grpc:pkg/api/v1 todo-service.proto
```

### Create database table:
```sql
CREATE TABLE ToDo (
    ID serial primary key unique,
    Title varchar (200) DEFAULT NULL,
    Description varchar(1024) DEFAULT NULL,
    Reminder timestamp NULL DEFAULT NULL
  );
```
