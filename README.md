# nomnomlog

[![Download nomnomlog](http://papertrail.github.io/nomnomlog/images/download.png)][releases]

nomnomlog tails one or more log files and sends syslog messages to a
remote central syslog server. It generates packets itself, ignoring the system
syslog daemon, so its configuration doesn't affect system-wide logging.

Uses:

* Collecting logs from servers & daemons which don't natively support syslog
* When reconfiguring the system logger is less convenient than a purpose-built daemon (e.g., automated app deployments)
* Aggregating files not generated by daemons (e.g., package manager logs)

This code is tested with the hosted log management service [Papertrail]
and should work for transmitting to any syslog server.


## Installing

Precompiled binaries for Mac (Darwin), Linux and Windows are available on the
[nomnomlog releases page][releases].

Untar the package, copy the "nomnomlog" executable into your $PATH,
and then customize the included example_config.yml with the log file paths
to read and the host/port to log to.

Optionally, move and rename the configuration file to `/etc/log_files.yml` so
that nomnomlog picks it up automatically. For example:

    sudo cp ./nomnomlog /usr/local/bin
    sudo cp example_config.yml /etc/log_files.yml
    sudo vi /etc/log_files.yml

Configuration directives can also be specified as command-line arguments (below).

## Usage

    Usage of nomnomlog:
      -c, --configfile string             Path to config (default "/etc/log_files.yml")
          --debug-log-cfg string          The debug log file; overridden by -D/--no-detach
      -d, --dest-host string              Destination syslog hostname or IP
      -p, --dest-port int                 Destination syslog port (default 514)
          --eventmachine-tail             No action, provided for backwards compatibility
      -f, --facility string               Facility (default "user")
          --hostname string               Local hostname to send from (default: OS hostname)
          --log string                    Set loggo config, like: --log="<root>=DEBUG" (default "<root>=INFO")
          --new-file-check-interval int   How often to check for new files (seconds) (default 10)
      -D, --no-detach                     Don't daemonize and detach from the terminal; overrides --debug-log-cfg
          --no-eventmachine-tail          No action, provided for backwards compatibility
          --pid-file string               Location of the PID file
          --poll                          Detect changes by polling instead of inotify
      -s, --severity string               Severity (default "notice")
          --tcp                           Connect via TCP (no TLS)
          --tls                           Connect via TCP with TLS

## Example

Daemonize and collect messages from files listed in `./example_config.yml` as
well as the file `/var/log/mysqld.log`. Write PID to `/tmp/nomnomlog.pid`
and send to port `logs.papertrailapp.com:12345`:

    $ nomnomlog -c example_config.yml -p 12345 --pid-file=/tmp/nomnomlog.pid /var/log/mysqld.log

Stay attached to the terminal, look for and use `/etc/log_files.yml` if it
exists, and send with facility local0 to `a.example.com:514`:

    $ nomnomlog -D -d a.example.com -f local0 /var/log/mysqld.log

## Auto-starting at boot

Sample init files can be found [in the examples directory](examples/). You may be able to:

    $ cp examples/nomnomlog.init.d /etc/init.d/nomnomlog
    $ chmod 755 /etc/init.d/nomnomlog

And then ensure it's started at boot, either by using:

    $ sudo update-rc.d nomnomlog defaults

or by creating a link manually:

    $ sudo ln -s /etc/init.d/nomnomlog /etc/rc3.d/S30nomnomlog

nomnomlog will daemonize by default.

Additional information about init files (`init.d`, `supervisor`, `systemd` and `upstart`) are
available [in the examples directory](examples/).

## Sending messages securely

If the receiving system supports sending syslog over TCP with TLS, you can
pass the `--tls` option when running `nomnomlog`:

    $ nomnomlog -D --tls -p 1234 /var/log/mysqld.log

or add `protocol: tls` to your configuration file.

## Configuration

By default, nomnomlog looks for a configuration in `/etc/log_files.yml`.

The archive comes with a [sample config](https://github.com/shadowbq/nomnomlog/blob/master/example_config.yml). Optionally:

    $ cp example_config.yml.example /etc/log_files.yml

`log_files.yml` has filenames to log from (as an array) and hostname and port
to log to (as a hash). Wildcards are supported using * and standard shell
globbing. Filenames given on the command line are additive to those in
the config file.

Only 1 destination server is supported; the command-line argument wins.

    files:
     - /var/log/httpd/access_log
     - /var/log/httpd/error_log
     - /var/log/mysqld.log
     - /var/run/mysqld/mysqld-slow.log
    destination:
      host: logs.papertrailapp.com
      port: 12345
      protocol: tls

nomnomlog sends the name of the file without a path ("mysqld.log") as
the syslog tag (program name).

After changing the configuration file, restart `nomnomlog` using the
init script or by manually killing and restarting the process. For example:

    /etc/init.d/nomnomlog restart

## Advanced Configuration (Optional)

Here's an [advanced config](https://github.com/shadowbq/nomnomlog/blob/master/examples/log_files.yml.example.advanced) which uses all options.

### Override hostname

Provide `--hostname somehostname` or use the `hostname` configuration option:

```yml
    hostname: somehostname
```

### Detecting new files

nomnomlog automatically detects and activates new log files that match
its file specifiers. For example, `*.log` may be provided as a file specifier,
and nomnomlog will detect a `some.log` file created after it was started.

By default, globs are re-checked every 10 seconds. To check for new files more
frequently, use the `--new-file-check-interval` argument. For example, to
recheck globs every 1 second, use:

    --new-file-check-interval 1

Note: messages may be written to new files in the period between when the
file is created and when the periodic glob check detects it. This data is not
transmitted.

If globs are specified on the command-line, enclose each one in single-quotes
(`'*.log'`) so the shell passes the raw glob string to nomnomlog (rather
than the current set of matches). This is not necessary for globs defined in
the config file.

### Log rotation and the behavior of nomnomlog

External log rotation scripts often move or remove an existing log file
and replace it with a new one (at a new inode). The Linux standard script
[logrotate](http://iain.cx/src/logrotate/) supports a `copytruncate` config
option.  With that option, `logrotate` will copy files, operate on the copies,
and truncate the original so that the inode remains the same.

`nomnomlog` will handle both approaches seamlessly, so it should be no
concern as to which method is used. If a log file is moved or renamed,
and a new file is created (at a new inode), `nomnomlog` will follow that
new file at the new inode (assuming it has the same absolute path name). If
a file is copied then truncated, `nomnomlog` will seek to the beginning of
the truncated file and continue to read it.

#### Log rotation edge cases to be aware of

Some logging programs such as Java's gclog (`-XX:+PrintGC` or `-verbose:gc`)
do not log in append mode, so if another program such as `logrotate` (set to
`copytruncate`) truncates the file, on the next write of the Java logger, the
OS will fill the file with NUL bytes upto the current offset of the file descriptor.
More info on that [here](http://stackoverflow.com/questions/8353401/garbage-collector-log-loggc-file-rotation-with-logrotate-does-not-work-properl).
`nomnomlog` will detect those leading NUL bytes, discard them, and log the discard count.

### Excluding files from being sent

Provide one or more regular expressions to prevent certain files from being
matched.

    exclude_files:
      - \.\d$
      - .bz2
      - .gz

### Excluding lines matching a pattern

There may be certain log messages that you do not want to be sent.  These may be
repetitive log lines that are "noise" that you might not be able to filter out
easily from the respective application.  To filter these lines, use the
exclude_patterns with an array or regexes:

    exclude_patterns:
     - exclude this
     - \d+ things

### Multiple instances

Run multiple instances to specify unique syslog hostnames.

To do that, provide an alternate PID path as a command-line option to the
additional instance(s). For example:

    --pid-file=/var/run/nomnomlog_2.pid

Note: Daemonized programs use PID files to identify whether the program is already
running ([more](http://unix.stackexchange.com/questions/12815/what-are-pid-and-lock-files-for/12818#12818)). Like other daemons, nomnomlog will refuse to run as a
daemon (the default mode) when a PID file is present. If a .pid file is
present but the daemon is not actually running, remove the PID file.

### Choosing app name

nomnomlog uses the log file name (like "access_log") as the syslog
program name, or what the syslog RFCs call the "tag." This is ideal unless
nomnomlog watches many files that have the same name.

In that case, tell nomnomlog to set another program name using the
`tag` attribute in the configuration file:

```yaml
files:
  - path: /var/log/httpd/access_log
    tag: apache
destination:
  host: logs.papertrailapp.com
  port: 12345
  protocol: tls
```

... or on the command line:
`nomnomlog apache=/var/log/httpd/access_log`

This functionality was introduced in version 0.17

## Troubleshooting

### Generate debug log

To output debugging events with maximum verbosity, run:

```shell
nomnomlog --debug-log-cfg=logfile.txt --log="<root>=DEBUG"
```

.. as well as any other arguments which are used in normal operation. This
will set [loggo](https://github.com/juju/loggo#func-parseconfigurationstring)'s
root logger to the `DEBUG` level and output to `logfile.txt`.

### Truncated messages

To send messages longer than 1024 characters, use TCP (either TLS or cleartext
TCP) instead of UDP. See "[Sending messages securely](#sending-messages-securely)" to
use TCP with TLS for messages of any length.

[Here's why](http://help.papertrailapp.com/kb/configuration/troubleshooting-remote-syslog-reachability/#message-length) longer UDP messages are impossible to send over the Internet.

### inotify

When running nomnomlog in the foreground using the `-D` switch, if you
receive the error:

    Error creating fsnotify watcher: inotify_init: too many open files

determine the maximum number of inotify instances that can be created using:

    cat /proc/sys/fs/inotify/max_user_instances

and then increase this limit using:

    echo VALUE >> /proc/sys/fs/inotify/max_user_instances

where VALUE is greater than the present setting. Confirm that nomnomlog starts
up and then apply this new value permanently by adding the following to
`/etc/sysctl.conf:`:

    fs.inotify.max_user_instances = VALUE

### "No space left on device"

When monitoring a large number of files, this error may occur:

    FATAL -- Error watching /path/here : no space left on device

To solve this, determine the maximum number of user watches that can be
created using:

    cat /proc/sys/fs/inotify/max_user_watches

and then increase them using:

    echo VALUE >> /proc/sys/fs/inotify/max_user_watches

Once again, confirm that nomnomlog starts and then apply this value permanently by adding the following to `/etc/sysctl.conf:`:

    fs.inotify.max_user_watches = VALUE

## Credits

* [Paul Morton](https://twitter.com/mortonpe)
* [Papertrail](https://papertrailapp.com/) staff
* [Paul Hammond](http://paulhammond.org/)

## Reporting bugs

1. See whether the issue has already been reported: <https://github.com/shadowbq/nomnomlog/issues/>
2. If you don't find one, create an issue with a repro case.

## Development

nomnomlog is written in go, and uses [govendor] to manage
dependencies. To get everything set up, [install go][goinstall] then
run:

    go get github.com/kardianos/govendor
    go get github.com/mitchellh/gox
    go get github.com/shadowbq/nomnomlog

To run tests:

    # run all tests
    go test ./...
    # run all tests except the slower syslog reconnection tests
    go test -short ./...

## Building

    make

### ARM support

As of 0.18, we introduced ARM support for nomnomlog. Current ARM builds
support all ARM platforms with hardware floating point instruction sets. This
includes All Raspberry PI devices, most ARMv6 chips (Cortex), and ARMv7 and
beyond.

## Contributing

Once you've made your great commits:

1. [Fork][fk] nomnomlog
2. Create a topic branch - `git checkout -b my_branch`
3. Commit the changes without changing the Rakefile or other files unrelated to your enhancement.
4. Push to your branch - `git push origin my_branch`
5. Create a Pull Request or an [Issue][is] with a link to your branch
6. That's it!

[Papertrail]: http://papertrailapp.com/
[nomnomlog]: https://github.com/papertrail/nomnomlog
[nomnomlog]: https://github.com/papertrail/nomnomlog
[release fork]: https://github.com/shadowbq/nomnomlog/releases

[govendor]: https://github.com/kardianos/govendor
[goinstall]: http://golang.org/doc/install

[fk]: http://help.github.com/forking/
[is]: https://github.com/shadowbq/nomnomlog/issues/
