Drone plugin to publish files and artifacts to Google Cloud Storage. For the usage information and a listing of the available options please take a look at [the docs](DOCS.md).

Local test:
```
mkdir dist
touch dist/test
go get ./...
go build 
./drone-gcs --auth_key [your_google_authentication_info] --source ./dist --target [gcs_bucket_name/[dir]] --ignore bin/* --acl allUsers:READER --gzip js --cache_control public,max-age=3600 --metadata '{"x-goog-meta-foo": "bar"}'
```

Or use docker:

```
mkdir dist
touch dist/test
docker run --env GOOGLE_KEY=[your_google_authentication_info] --env PLUGIN_SOURCE=./dist --env PLUGIN_TARGET=[gcs_bucket_name/[dir]] maplain/drone-gcs:latest
```

Or if your drone is configured:
```
drone exec --secret GOOGLE_KEY=[your_google_authentication_info] --secret PLUGIN_SOURCE=./dist --secret PLUGIN_TARGET=[gcs_bucket_name/[dir]] --plugin gcs
```
