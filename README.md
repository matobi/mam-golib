# mam-golib

## Maintainer
This service is maintained by **Team Media Provisioning**.

## Depends on
{Add info about docker pods}

## Description
This project contains some basic Go library packages that are used by many mam Go services.

*For reasons described below it is important to only use generic code in this project and no passwords or any secrets.*

This library is included into other projects during build in Jenkins. Since our bitbucket projects are protected with passwords,
it would require ssh configuration on the Jenkins server to automatically fetch this project.
Insted a copy of this project is on an open github account _github.com/matobi/mam-golib_. 

## Building locally
`make`

## Update version
Projects that use this library imports a specific version tag of this lib.
Follow these steps to update the library version for a project.

1. Fix code in this project, and then set a new version tag. Use `git tags`to see old tags. For example to set version v1.0.0:

`
git tag v1.2.0
git push origin v1.2.0
`

2. In the project that is using this library, open file _go.mod_ and remove the reference to this library.
Rebuild the project with `make` and it will update reference to the latest version tag.
