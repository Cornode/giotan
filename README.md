[![Build Status](https://travis-ci.org/cornodeledger/gcornoden.svg?branch=master)](https://travis-ci.org/cornodeledger/gcornoden)
[![GoDoc](https://godoc.org/github.com/cornodeledger/gcornodenn?status.svg)](https://godoc.org/github.com/cornodeledger/gcornoden)
[![GitHub license](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/cornodeledger/gcornoden/master/LICENSE)

:frog: gcornode CLI tool
=====

CLI client tool for the cornode reference implementation (IRI) using gcornode lib.

Install
====
```
    $ go get -u github.com/cornodeledger/gcornoden
```

Features
====

1. Sending cornode token using [public nodes](http://cornodesupport.com/lightwallet.shtml) with local PoW.
2. List used and unused Addresses which can be generated from seed.

This CLI mainly focuses on functions using seeds.

If you want to add some functions, please make an issue.

Examples
====

```
    $ gcornoden new
    $ gcornoden addresses 
    $ gcornoden send --recipient=SOMERECIPIENT --amount=1234
    $ gcornoden send --recipient=SOMERECIPIENT --amount=1234 --sender=SOMEADDRESS1,SOMEADDRESS2,SOMEADDRESS3
```

When you use `addresses` and `send`, you will be prompted to input your seed.

When you use --sender, you must specify the addresses which can be generated from `seed`.

Note that `send` takes a long time to calculate Proof of Work.

Development Status: Alpha+
=========================

Tread lightly around here. This tool is still very much
in flux and there are going to be breaking changes.


TODO
=========================

* [ ] More functions(?)
* [ ] More tests :(

<hr>

Released under the [MIT License](LICENSE).
