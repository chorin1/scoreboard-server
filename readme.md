# Scoreboard-server

A simple, yet powerful, single instance implementation of a scoreboard server composed of 2 dockers:
1. [Redis](https://redis.io/) to store the scores and ranks in-memory.
2. Golang backend web-handler to receive score requests, based on the excellent [Fiber](https://github.com/gofiber/fiber) framework.

## Quickstart

1. Clone the repo to your localhost or to a cloud instance.
2. Edit `docker-compose.yml` if needed.
3. `docker-compose up`.

## Prerequisites
Docker & [Docker Compose](https://docs.docker.com/compose/install/)


## Usage
### Get highest ranked scores (top 10)
#### Request
`GET /getScores`
```bash
curl -u admin:admin https://localhost/getScores
```
#### Response
```json
{
    "count": 3,
    "Users": [
        {
            "name": "Alice",
            "score": 5000,
            "rank": 1
        },
        {
            "name": "Bob",
            "score": 4000,
            "rank": 2
        },
        {
            "name": "Eve",
            "score": 3000,
            "rank": 3
        }
    ]
}
```

### Register a new high score
#### Request
`POST /newScore`
```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -u admin:admin \
  -d '{"name":"tentacle","uuid":"1feb6f78-6db9-4ef7-97de-5fd2a5eeed62","score":12345}' \
  https://localhost/newScore
```
#### Response
```json
{
    "rank": 17
}
```


## Overwriting scores
* `uuid` is used to allow different high scorers to use the same name.
* A score key is stored as a combination of `uuid` + `name`.
* If a new score is registered with the same `uuid` + `name` combo it will only be overwritten if it is greater than the last score.
* Otherwise, a rank of `-1` will be returned.
* `uuid` is encapsulated and not returned by the server.

## HTTPS
By default, a self-signed certificate will be generated upon starting the service.
To change this behaviour, small modifications need to be made to `docker-compose.yml` and `main.go`

## Scores eviction
Since Redis sorted sets do not support size limiting, periodically, all scores below the 10,000th rank will be evicted.

## Contributing
Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

Please make sure to update tests as appropriate.

## License
[MIT](https://choosealicense.com/licenses/mit/)