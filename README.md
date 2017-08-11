# plumcliraw

This is a CLI that exercises libplumraw - it basically is almost as hard to use as running `curl`s directly, but at least demonstrates some of how to use [libulpmraw](https://github.com/maplebed/libplumraw). For a more interesting daemon, check out [plumd](https://github.com/maplebed/plumd).

## Credit Where Credit Is Due

All the work for this code and the libraries on which it depends are made possible by the fine work of [mikenemat](https://github.com/mikenemat) and [TheSourceLies](https://github.com/TheSourceLies). Mikenemat initially reverse engineered the protocols with which the lightpad and TheSourceLies created excellent documentation of the REST API for the Plum production web service.

See their code at
* https://github.com/mikenemat/plum-probe
* https://github.com/TheSourceLies/plum-lightpad-php-api
