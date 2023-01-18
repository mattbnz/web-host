# fly.io hugo static hosting wrapper

Basic nginx setup to serve a hugo site checked out from github statically on fly.io.

Includes a webhook endpoint (used as a ping only, not actually processed :p)
for github to hit when a push has been made and the site should be updated.
