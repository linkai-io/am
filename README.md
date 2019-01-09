# AM
This is the primary repository for all linkai - AM services.

## Structure
* am/ contains all the domain types. It must not import from *any* other AM packages.
* amtest/ testing helper package and e2e tests (e2e disabled by default)
* clients/ all grpc client implementations
* cmd/ contains all runnable binaries, to be built using Docker
* internal/ internal only packages for the am project
* lambda/ contains all lambda functions
* mock/ mock implementations of services for testing
* modules/ scan modules for the AM service
* pkg/ packages that can be imported from other repositories/projects
* protocservices/ grpc/protobuf services and types automatically generated via make protoc
* services/ the various microservices used by the AM system

## Building
am uses go dep until vgo is better supported.

Building docker images can be done via make {servicename}

## Testing
Testing requires a local or env database is configured. For local testing make sure you check out the database repository, run the cmd/pgm first to create the linkai users and db
then run cmd/amm and deploy all changes prior to running tests.
After which, simply run make test.

## Running locally
First: start consul:
consul agent -dev -config-dir=./consul.d -data-dir=consul_data/ -advertise="127.0.0.1" -client="172.16.238.1 127.0.0.1"
Next: Run docker-compose:
cd testing && docker-compose up
Next: Use amcli to issue commands (see cmd/amcli/README.md for more details)

## Secrets
Secrets are managed via AWS Parameter Store for dev/production. For local testing create environment variables such as:
```json
{
    "go.testEnvVars": {
        "_am_local_db_orgservice_dbstring": "user=orgservice port=5000 dbname=linkai password=XXX sslmode=disable",
        "_am_local_db_jobservice_dbstring": "user=jobservice port=5000 dbname=linkai password=Xxx sslmode=disable",
        "_am_local_db_inputservice_dbstring": "user=inputservice port=5000 dbname=linkai password=Xxx sslmode=disable",
        "_am_local_db_userservice_dbstring": "user=userservice port=5000 dbname=linkai password=Xxx sslmode=disable",
        "_am_local_db_tagservice_dbstring": "user=tagservice port=5000 dbname=linkai password=Xxx sslmode=disable",
        "_am_local_db_scangroupservice_dbstring": "user=scangroupservice port=5000 dbname=linkai password=Xxx sslmode=disable",
        "_am_local_db_hostservice_dbstring": "user=hostservice port=5000 dbname=linkai password=Xxx sslmode=disable",
        "_am_local_db_linkai_admin_dbstring": "user=linkai_admin port=5000 dbname=linkai password=Xxx sslmode=disable",
        "_am_local_db_postgres_dbstring": "user=postgres port=5000 dbname=postgres password=Xxx sslmode=disable"
    }
}
```

## Change Considerations
When updating a database table or the domain types consider the following:
1. Is the domain type stored in DB?
    - Yes, update database schema
    - Yes, update the [service]_statements.go file(s).
2. Is the domain type transfered over grpc?
    - Yes, update the relevant protorepo definitions
        - Don't forget to re-run make protoc
    - Yes, update: github.com/linkai-io/am/pkg/convert/protoc.go
    - Yes, update: *all* services which use the domain type
    - Yes, update: *all* clients which use the domain type
3. Update both local tests and amtests if it references the domain type
4. Is the domain type stored in redis?
    - Yes, update the package(s) redis state system 
5. Extra paranoid: search project for all references
6. Re-run make services
