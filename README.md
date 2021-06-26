weeclient
==========

Weeclient is a relay client for [Weechat](https://weechat.org).

[![Go Documentation](https://pkg.go.dev/badge/github.com/maxking/weeclient.svg)](https://pkg.go.dev/github.com/maxking/weeclient) [![Gituhb Action](https://github.com/maxking/weeclient/actions/workflows/go.yml/badge.svg)](https://github.com/maxking/weeclient/actions/workflows/go.yml)


![Weeclient Screenshot](images/screenshot.png)

Dependencies
------------

- Golang 1.16


Build
-----

To build the code, you can run:

```bash
$ git checkout https://github.com/maxking/weeclient
$ cd weeclient
$ make
```

Then, you can run with:
```bash
$ ./weeclient
```

KeyBindings
-----------

**Buffer list**

- <kbd>Ctrl</kbd> + <kbd>b</kbd>: Move focus to buffer list.
- <kbd>Enter</kbd> : When in buffer list, this will move focus to the input box of the buffer.

**Buffer view**

- <kbd>Ctrl</kbd> + <kbd>s</kbd>: Move focus to the chat and allow scrolling up and down
- <kbd>Ctrl</kbd> + <kbd>i</kbd>: Move focus to the input box.

**Input box**
- <kbd>Esc</kbd>: Clear the box.
- <kbd>Enter</kbd>: Send the message in the box.


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
