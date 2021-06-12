weechat-go
==========

Weechat-go is a relay client for [Weechat](https://weechat.org).


Dependencies
------------

- Golang 1.16


Build
-----

To build the code, you can run:

```bash
$ git checkout https://github.com/maxking/weechat-go
$ cd weechat-go
$ make
```

Then, you can run with:
```bash
$ ./weechat-go
```

Testing Relay
-------------

This repo comes with a testing CLI that can be used to interact
with the Weechat Relay. Currently, the information printed is really
not that much, but in future, it should be printed better.

```bash
$ ./testclient
```


License
-------
All the contents of this repo are licensed under Apache 2.0 license. 
Please see the LICENSE file included for a full copy of the License.