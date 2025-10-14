# Sample API
- On `GET /`, log the request (will update a counter in a PG in the future)


`docker build . -t api -f Containerfile`
`docker run -p 3030:3030 api:latest`
