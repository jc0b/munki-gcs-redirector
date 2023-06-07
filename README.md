# Munki GCS Redirector

Some sample source for 'converting' BasicAuth-authenticated GET requests to signed GCS URLs.

This is intended to be run on Google Cloud Run behind a load balancer, with a service account with URL signing permissions attached to the deployment.
The code will attempt to use Application Default Credentials provided by the attached service account when attempting to authenticate.

Lastly, I don't guarantee that this code works perfectly. I don't run this code in production, and it may require a little bit of tweaking to fit your use-case.

## Configuration
There are 3 environment variables that need to be set:
`AUTH_USERNAME`: the desired basic auth username
`AUTH_PASSWORD`: the desired basic auth password
`GCS_BUCKET_NAME`: the bucket you are using as the Munki repo (no `gs://` required - e.g. just `jc0b-munki-repo`)
