
# JumpCloud Hash and Encode a Password String

A simple REST http server that listens on a given port and encodes a password. Provides the following endpoints:

## Endpoints

| Endpoint  | Request   | Description                                                                                                                                                                                    |
|-----------|-----------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| /hash     | POST      | Handles POST requests on the /hash endpoint with a form field "password" provding the value to hash. Returns an incrementing identifier immediately but the password is not hashed for 5 secs. |
| /hash/    | GET       | Handles GET requests to retrieve a hashed password by its id.                                                                                                                                  |
| /stats    | GET       | Handles GET requests for basic information about password hashes.                                                                                                                              |
| /shutdown | GET       | Handles GET “graceful shutdown request”.                                                                                                                                                       |

## To Run

- Clone https://github.com/rumyanaruseva/jumpcloud_password_hash
- In jumpcloud_password_hash folder, type:
    - `go run main.go` to start the server on default port 8080, or
    - `go run main.go -port <port num>`, to start the server on port `<port num>`, e.g. `go run main.go -port 1234`


## Notes

- Build with Go 1.17 on Windows
