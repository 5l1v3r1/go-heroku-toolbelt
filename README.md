go-heroku-toolbelt
==================

Implementation of the Heroku Toolbelt in Golang.

## Why
Current version of Heroku Toolbelt is in Ruby. Deploying a version with Golang allows deployment of just one executable, instead of dependencies on Ruby.

## What is working currently?
* only the logs command

## Architecture
Extra commands are easy to implement. Just create the struct, implement Run and register it.

## Example
```
go run heroku.go --app appname logs
```

## Contributions
* Remco (DutchCoders, Github: @nl5887, Twitter: @remco_verhoef)
