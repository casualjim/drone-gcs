Use this plugin to upload files and build artifacts
to the [Google Cloud Storage (GCS)](https://cloud.google.com/storage/) bucket.

You will need to a [Service Account](https://developers.google.com/console/help/new/#serviceaccounts)
to authenticate to the GCS.

The following parameters are used to configure this plugin:

* `source` - location of files to upload
* `target` - destination to copy files to, including bucket name
* `ignore` - skip files matching this [pattern](https://golang.org/pkg/path/filepath/#Match), relative to `source`
* `acl` - a list of access rules applied to the uploaded files, in a form of `entity:role`
* `gzip` - files with the specified extensions will be gzipped and uploaded with "gzip" Content-Encoding header
* `cache_control` - Cache-Control header
* `metadata` - an arbitrary dictionary with custom metadata applied to all objects

The following secret values can be set to configure the plugin.
* `GOOGLE_KEY` - corresponds to auth_key

The following is a sample configuration in your .drone.yml file:

```yaml
publish:
  gcs:
    source: dist
    target: bucket/dir/
    ignore: bin/*
    acl:
      - allUsers:READER
    gzip:
      - js
      - css
      - html
    cache_control: public,max-age=3600
    metadata:
      x-goog-meta-foo: bar
```

`GOOGLE_KEY` should be added as secret.
