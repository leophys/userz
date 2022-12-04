# Userz
> A simple demonstration project for a k8s microservice

The project aims to exemplify a very simple cloud native microservice to handle
a store of users. The json schema is the following:

```
{
    "id": string,
    "first_name": optional<string>,
    "last_name": optional<string>,
    "nickname": string,
    "email": string,
    "password": bytes, // bcrypt of a given string
    "country": optional<string>,
    "created_at": optional<timestamp with timezone>,
    "updated_at": optional<timestamp with timezone>
}
```

The service exposes the expected CRUD interface both on HTTP and via gRPC, has a
notification mechanism based on plugins and offers both and healthcheck endpoint
and prometheus metrics.

### Run the project

To interact with the project, one may use GNU make to spin up a docker compose
stack. These, together with `curl` and `git`, are the only prerequisites.

```
make run
```

compiles the image and starts the project. To stop and clean everything

```
make stop
```

While running the project, some mock data can be added via

```
make data
```

The project has a somehow comprehensive test coverage. The unit tests can be
run with

```
make test
```

The integration tests can be run with

```
make test-integration
```

In case of failure, the stack for the integration tests remains up, so that one
can examine the database (exposed at `localhost:5432`). To clean the stack
after a failed run (no need in case of a successful run)

```
make test-clean
```

## Some details

### The executable


The `userz` executable currently offers the following configurations:

```
NAME:
   userz - manage a list of users

USAGE:
   userz [global options] command [command options] [arguments...]

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --debug                      Set logging to debug level (defaults to info) (default: false) [$DEBUG]
   --console                    Enable pretty (and slower) logging (default: false)
   --http-port value            The port on which the HTTP API will be exposed (default: 6000) [$HTTP_PORT]
   --grpc-port value            The port on which the gRPC API will be exposed (default: 7000) [$GRPC_PORT]
   --grpc-cert value            The path to a TLS certificate to use with the gRPC endpoint [$GRPC_CERT]
   --grpc-key value             The path to a TLS key to use with the gRPC endpoint [$GRPC_KEY]
   --metrics-port value         The port on which the metrics will be exposed (healthcheck and prometheus) (default: 25000) [$METRICS_PORT]
   --pgurl value                The url to connect to the postgres database (if specified, supercedes all other postgres flags) [$POSTGRES_URL]
   --pguser value               The user to connect to the postgres database [$POSTGRES_USER]
   --pghost value               The host to connect to the postgres database (default: "localhost") [$POSTGRES_HOST]
   --pgpassword value           The password to connect to the postgres database [$POSTGRES_PASSWORD]
   --pgport value               The port to connect to the postgres database (default: 5432) [$POSTGRES_PORT]
   --pgdbname value             The dbname to connect to the postgres database [$POSTGRES_DBNAME]
   --pgssl                      Whether to connect to the postgres database in strict ssl mode (default: false) [$POSTGRES_SSL]
   --disable-notifications      Whether to disable notifications (default: false) [$DISABLE_NOTIFICATIONS]
   --notification-plugin value  Specify path to the .so that provides the notification functionality (default: "/pollednotifier.so") [$NOTIFICATION_PLUGIN]
   --help, -h                   show help (default: false)
```

### The HTTP REST API

The api is exposed, by default, at `http://localhost:6000/api` (the port is
configurable). It follows the REST paradigm, so

  - Creation is a `PUT` at `/api` and returns the id of the newly created
    entity.
  - Update is a `POST` at `/api/{id}`, and returns the whole updated entity.
  - Remove is a `DELETE` at `/api/{id}`, and returns the whole deleted entity.
  - Access is a `GET` at `/api`, with an optional `filter` and a mandatory
    `pageSize` and `offset` parameters, expected to be positive integers.

Both the creation and the update expect a JSON body with the following schema

```
{
    "first_name": optional<string>,
    "last_name": optional<string>,
    "nickname": string,
    "email": string,
    "password": string, // the plaintext, will be stored bcrypt'ed
    "country": optional<string>,
}
```

### The gRPC API

The gRPC API follows along the lines of the HTTP one, except for the access: it
is a stream that must be consumed linearly. The protobuf definition is at
[pkg/proto/userz.proto](./pkg/proto/userz.proto).
It is importable externally using


```
github.com/leophys/userz/pkg/proto
```

### The notification system

Notifications follow an extensible mechanism, based on the stdlib `plugin`
module. The default implementation, which can be used as model for future
implementations to support other technologies, lies at
[internal/pollednotifier](./internal/pollednotifier). It exposes a
`/notifications` HTTP endpoint (by default on port 8000) and every `GET`
towards that endpoint returns a JSON array that gets consumed at every access.

The public interface that has to be implemented by other plugins is at
[pkg/notifier/notifier.go](./pkg/notifier/notifier.go)
