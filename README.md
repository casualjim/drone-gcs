Drone plugin to publish files and artifacts to Google Cloud Storage. For the usage information and a listing of the available options please take a look at [the docs](DOCS.md).

```
mkdir dist
touch dist/test
go get ./...
go install 
drone-gcs --auth_key [your_google_authentication_file] --source dist --target gcs_bucket_name/dir --ignore bin/* --acl allUsers:READER --gzip js --cache_control public,max-age=3600 --metadata '{"x-goog-meta-foo": "bar"}'
```
