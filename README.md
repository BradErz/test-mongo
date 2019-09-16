# test-mongo
Reproduction of the issues:
https://jira.mongodb.org/browse/GODRIVER-1297

Run mongo locally with authentication enabled

```bash
docker run -p 27017:27017 -d \
    -e MONGO_INITDB_ROOT_USERNAME=mongoadmin \
    -e MONGO_INITDB_ROOT_PASSWORD=secret \
    --name=mongo \
    mongo:3.6
```

build the test-mongo container
```bash
docker build .
```

Run the container with a low amount of CPU
You will very quickly see the error (for me its within the first 10-20 queries)
```bash
docker run --cpus=0.1 --memory=100m 839a8a21e450
time="2019-09-16T11:08:26Z" level=info msg="created 10000 users"
time="2019-09-16T11:08:26Z" level=info msg="dropping previous users in database"
time="2019-09-16T11:08:26Z" level=info msg="adding 10000 users to db"
time="2019-09-16T11:09:07Z" level=info msg="found test-nickname-1 users from query"
time="2019-09-16T11:09:07Z" level=info msg="found test-nickname-9 users from query"
time="2019-09-16T11:09:07Z" level=info msg="found test-nickname-12 users from query"
time="2019-09-16T11:09:07Z" level=info msg="found test-nickname-13 users from query"
time="2019-09-16T11:09:08Z" level=error msg="error getting users from database" error="connection(172.17.0.1:27017[-3]) unable to read full message: read tcp 172.17.0.3:32840->172.17.0.1:27017: i/o timeout"
time="2019-09-16T11:09:08Z" level=error msg="error getting users from database" error="connection() : auth error: sasl conversation error: unable to authenticate using mechanism \"SCRAM-SHA-1\": context deadline exceeded"
time="2019-09-16T11:09:09Z" level=error msg="error getting users from database" error="connection() : auth error: sasl conversation error: unable to authenticate using mechanism \"SCRAM-SHA-1\": context deadline exceeded"

```


if you do the same test with 1 core the problem isn't present
```bash
docker run --cpus=1 --memory=100m 839a8a21e450
```

MONGODB_URI can be set if not running docker locally with the provided command.

When running the same code but with minPoolSize specified it is able to easily handle the requests without causing 
issues or timeouts to the database.
```bash
docker run --cpus=0.1 --memory=100m \
    -e MONGODB_URI="mongodb://mongoadmin:secret@172.17.0.1:27017/test-mongo?authSource=admin&minPoolSize=10" \
    839a8a21e450
```

Im not sure how the driver connections work in other drivers but it seems bad that the call to the database from an 
operation is in charge of opening a connection to the database and there isn't some other process watching the 
connections and adding some preemptively or always able to keep a buffer of x without having to specify a minimum poolSize