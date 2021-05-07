# simple-server
A simple Go webserver (intended to be deployed to Google Run) that serves files from Google Cloud Storage (GCS) with simple routing rules. This makes it simple to host the static files for a single-page application (SPA) from Google Cloud.

### Usage / Is this really necessary?
One of the simplest ways to host a static website on Google Cloud is to upload it to Google Cloud Storage and then serve that bucket via a Cloud Load Balancer. This has been documented [here](https://cloud.google.com/storage/docs/hosting-static-website).

**The Catch**
If the static website is an SPA, there are limitations on types of client-side routes you can employ. GCS currently allows a bucket to be configured with an "Index page suffix" and an "Error (404 not found) page". This can be configured to ensure that all requests get routed back to the index.html page which serves as the root of your SPA. However, for any route (other than the root page) this response will be served with a 404 error code. This is non-ideal and results in strange browser behavior.

### Solution
_Work in progress:_ it would be ideal if GCS could simply be configured to suppress the 404 response code when serving the "error" file. Unfortunately, this does not appear to be possible at this time. This solution allows a lightweight alternative to 

### Reference
This solution borrows heavily from [google-storage-proxy](https://github.com/cirruslabs/google-storage-proxy) which has been adapted to solve this particular problem.
