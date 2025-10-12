# Sample web-server
- Expose a index.html
- On `GET /send`, send a request to `$apiUrl` env var

`docker build . -t web-server -f Containerfile`
`docker run -p 8080:8080 -e API_URL=http://127.0.0.1:3030 web-server:latest`
