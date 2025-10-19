# Sample web-server
- Expose a index.html
- On `GET /send`, send a request to `$apiUrl` env var

```
cp env.dist .env
```

Using go

```
go run main.go
```

```
firefox http://127.0.0.1:8080
```

OR

```
curl 127.0.0.1:8080/send
```

Or using docker:

```
docker build . -t web-server -f Containerfile
docker run -p 8080:8080 -e API_URL=http://127.0.0.1:3030 web-server:latest
```

