# Images
There are four components(Nsqd/Nsqlookupd/Nsqadmin/Nginx) used in Nsq-Operator.
Either to the images related to Nsq-Operator.

## PID 1 process
A `SIGTERM` signal is sent to the PID 1 process inside container when
[`docker stop` a container](https://docs.docker.com/engine/reference/commandline/stop/#extended-description)
or [`kubectl delete` a pod](https://kubernetes.io/docs/concepts/workloads/pods/pod/#termination-of-pods).

[Nsqd needs `SIGTERM` or `SIGINT` to do a graceful shutdown](https://github.com/nsqio/nsq/releases/tag/v0.3.8).

There are two solutions to run Nsqd:
* Use a hosting service, such as [`systemd`](https://www.freedesktop.org/wiki/Software/systemd/),
[`supervisord`](http://supervisord.org/), to guard Nsqd instance
* Make Nsqd instance be the PID 1 process

Both `systemd` and `supervisor` do not support transmiting received
signals to its hosted services. Actually, most hosting service does
not support transmiting received `SIGTERM` signal to its guarded
services. So it seems that solution one is not an option.

By making Nsqd instance be the PID 1 process, `SIGTERM` signal
can be sent directly to Nsqd and a graceful shutdown will be done.

## Log rotation
Currently Nsqd/Nsqlookupd/Nsqadmin will output logs to stderr/stdout.
There is no log rotation.

[cronolog](https://linux.die.net/man/1/cronolog) is a simple program
that read log messages from its input and writes them to a set of output
 files, the names of which are constructed using template and the
 current date and time. The template uses the same format specifiers as
 the Unix [date](https://linux.die.net/man/1/date) command.

Thus, by redirecting stderr/stdout of Nsqd/Nsqlookupd/Nsqadmin to the
input of cronolog, we can simply rotate logs.
