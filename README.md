```
make dist
touch dist/test
go install 
drone-gcs --auth_key [your_google_authentication_file] --source dist --target bucket/dir --ignore bin/* --acl allUsers:READER --gzip js --cache_control public,max-age=3600 --metadata '{"x-goog-meta-foo": "bar"}'
```
