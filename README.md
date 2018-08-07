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
Until vgo is better supported:
* go get -u ./... - to get dependencies

Building docker images can be done via make {servicename}

## Testing
Testing requires a local or env database is configured. For local testing make sure you check out the database repository, run the cmd/pgm first to create the linkai users and db
then run cmd/amm and deploy all changes prior to running tests.
After which, simply run make test.

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


