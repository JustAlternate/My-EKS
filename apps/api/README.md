# Sample API with its postgres server in a compose file
- On `GET /`, log the request (will update a counter in a PG in the future)

```
cp env.dist .env
```

Launch the postgres and api server in local
```
docker compose up
```
