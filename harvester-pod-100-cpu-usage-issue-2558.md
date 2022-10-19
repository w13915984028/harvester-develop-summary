# How to troubleshooting when a POD is 100% busy but there is no clear debug/error log

## Issue

### Related issues

https://github.com/harvester/harvester/issues/2558

https://github.com/rancher/dynamiclistener/issues/63

https://github.com/harvester/harvester/issues/2941 (update 2022.10.19)

### Description of the issue

The `Harvester` POD has a stable 100% cpu usage, but from the log, nothing seems wrong.

```
harvester-756cd66d6d-x4gm6:/var/lib/harvester/harvester # 

Tasks:   4 total,   1 running,   3 sleeping,   0 stopped,   0 zombie
%Cpu(s): 17.3 us,  1.4 sy,  0.0 ni, 78.4 id,  2.9 wa,  0.0 hi,  0.0 si,  0.0 st
MiB Mem : 16002.84+total, 1420.094 free, 12233.33+used, 2349.410 buff/cache
MiB Swap:    0.000 total,    0.000 free,    0.000 used. 3433.902 avail Mem 

  PID USER      PR  NI    VIRT    RES    SHR S  %CPU  %MEM     TIME+ COMMAND                                                                                                                              
    8 root      20   0 1111168 232776  27632 S 93.75 1.420   6:18.50 harvester                                                                                                                            
    1 root      20   0    4564    712    620 S 0.000 0.004   0:00.01 tini                                                                                                                                 
   22 root      20   0   15124   4184   2988 S 0.000 0.026   0:00.01 bash                                                                                                                                 
   49 root      20   0   40448   4132   3576 R 0.000 0.025   0:00.01 top                                                                                                                                  
```

### Log all events

Harvester POD is event-driven, all `k8s controllers` inside this POD are looping processing events.

```
vendor/github.com/rancher/lasso/pkg/controller/controller.go
func (c *controller) processSingleItem(obj interface{}) error {
..
}
```

Try to log them in `processSingleItem`, looks normal.

```
time="2022-07-22T21:57:08Z" level=error msg="processSingleItem: harv1"
time="2022-07-22T21:57:09Z" level=error msg="processSingleItem: cattle-fleet-system/gitjob"
time="2022-07-22T21:57:09Z" level=error msg="processSingleItem: kube-system/harvester-controllers"
time="2022-07-22T21:57:10Z" level=error msg="processSingleItem: cattle-fleet-system/fleet-controller-lock"
time="2022-07-22T21:57:10Z" level=error msg="processSingleItem: cattle-logging-system/logging-operator.logging.banzaicloud.io"
time="2022-07-22T21:57:10Z" level=error msg="processSingleItem: cattle-fleet-local-system/fleet-agent-lock"
time="2022-07-22T21:57:10Z" level=error msg="processSingleItem: kube-system/rke2"
time="2022-07-22T21:57:10Z" level=error msg="processSingleItem: kube-system/cattle-controllers"
time="2022-07-22T21:57:10Z" level=error msg="processSingleItem: kube-system/harvester-network-controllers"
time="2022-07-22T21:57:10Z" level=error msg="processSingleItem: kube-system/harvester-load-balancer"
time="2022-07-22T21:57:11Z" level=error msg="processSingleItem: cattle-fleet-system/gitjob"
time="2022-07-22T21:57:11Z" level=error msg="processSingleItem: kube-system/harvester-controllers"
time="2022-07-22T21:57:12Z" level=error msg="processSingleItem: cattle-fleet-system/fleet-controller-lock"
time="2022-07-22T21:57:12Z" level=error msg="processSingleItem: cattle-logging-system/logging-operator.logging.banzaicloud.io"
time="2022-07-22T21:57:12Z" level=error msg="processSingleItem: cattle-fleet-local-system/fleet-agent-lock"
time="2022-07-22T21:57:12Z" level=error msg="processSingleItem: kube-system/rke2"
time="2022-07-22T21:57:12Z" level=error msg="processSingleItem: kube-system/cattle-controllers"
time="2022-07-22T21:57:12Z" level=error msg="processSingleItem: kube-system/harvester-network-controllers"
time="2022-07-22T21:57:12Z" level=error msg="processSingleItem: kube-system/harvester-load-balancer"
time="2022-07-22T21:57:13Z" level=error msg="processSingleItem: cattle-fleet-system/gitjob"
time="2022-07-22T21:57:13Z" level=error msg="processSingleItem: kube-system/harvester-controllers"
time="2022-07-22T21:57:14Z" level=error msg="processSingleItem: cattle-fleet-system/fleet-controller-lock"
time="2022-07-22T21:57:14Z" level=error msg="processSingleItem: cattle-logging-system/logging-operator.logging.banzaicloud.io"
time="2022-07-22T21:57:14Z" level=error msg="processSingleItem: cattle-fleet-local-system/fleet-agent-lock"
time="2022-07-22T21:57:14Z" level=error msg="processSingleItem: kube-system/rke2"
time="2022-07-22T21:57:14Z" level=error msg="processSingleItem: kube-system/cattle-controllers"
time="2022-07-22T21:57:14Z" level=error msg="processSingleItem: kube-system/harvester-network-controllers"
time="2022-07-22T21:57:14Z" level=error msg="processSingleItem: kube-system/harvester-load-balancer"
time="2022-07-22T21:57:15Z" level=error msg="processSingleItem: cattle-fleet-system/gitjob"
time="2022-07-22T21:57:15Z" level=error msg="processSingleItem: kube-system/ingress-controller-leader"
time="2022-07-22T21:57:15Z" level=error msg="processSingleItem: kube-system/harvester-controllers"
time="2022-07-22T21:57:16Z" level=error msg="processSingleItem: cattle-fleet-system/fleet-controller-lock"
time="2022-07-22T21:57:16Z" level=error msg="processSingleItem: cattle-logging-system/logging-operator.logging.banzaicloud.io"
time="2022-07-22T21:57:16Z" level=error msg="processSingleItem: cattle-fleet-local-system/fleet-agent-lock"
time="2022-07-22T21:57:16Z" level=error msg="processSingleItem: kube-system/rke2"
time="2022-07-22T21:57:16Z" level=error msg="processSingleItem: kube-system/cattle-controllers"
time="2022-07-22T21:57:16Z" level=error msg="processSingleItem: kube-system/harvester-network-controllers"
time="2022-07-22T21:57:16Z" level=error msg="processSingleItem: kube-system/harvester-load-balancer"
time="2022-07-22T21:57:17Z" level=error msg="processSingleItem: cattle-fleet-system/gitjob"
time="2022-07-22T21:57:17Z" level=error msg="processSingleItem: kube-system/harvester-controllers"
time="2022-07-22T21:57:18Z" level=error msg="processSingleItem: cattle-fleet-system/fleet-controller-lock"
time="2022-07-22T21:57:18Z" level=error msg="processSingleItem: cattle-logging-system/logging-operator.logging.banzaicloud.io"
time="2022-07-22T21:57:18Z" level=error msg="processSingleItem: cattle-fleet-local-system/fleet-agent-lock"
time="2022-07-22T21:57:18Z" level=error msg="processSingleItem: kube-system/rke2"
time="2022-07-22T21:57:18Z" level=error msg="processSingleItem: kube-system/cattle-controllers"
time="2022-07-22T21:57:18Z" level=error msg="processSingleItem: kube-system/harvester-network-controllers"
time="2022-07-22T21:57:18Z" level=error msg="processSingleItem: kube-system/harvester-load-balancer"
time="2022-07-22T21:57:19Z" level=error msg="processSingleItem: cattle-fleet-system/gitjob"
time="2022-07-22T21:57:19Z" level=error msg="processSingleItem: kube-system/harvester-controllers"
time="2022-07-22T21:57:20Z" level=error msg="processSingleItem: cattle-fleet-system/fleet-controller-lock"
time="2022-07-22T21:57:20Z" level=error msg="processSingleItem: cattle-logging-system/logging-operator.logging.banzaicloud.io"
time="2022-07-22T21:57:20Z" level=error msg="processSingleItem: cattle-fleet-local-system/fleet-agent-lock"
time="2022-07-22T21:57:20Z" level=error msg="processSingleItem: kube-system/rke2"
time="2022-07-22T21:57:20Z" level=error msg="processSingleItem: kube-system/cattle-controllers"
time="2022-07-22T21:57:20Z" level=error msg="processSingleItem: kube-system/harvester-network-controllers"
time="2022-07-22T21:57:20Z" level=error msg="processSingleItem: kube-system/harvester-load-balancer"
time="2022-07-22T21:57:21Z" level=error msg="processSingleItem: cattle-fleet-system/gitjob"
time="2022-07-22T21:57:21Z" level=error msg="processSingleItem: kube-system/harvester-controllers"
time="2022-07-22T21:57:22Z" level=error msg="processSingleItem: cattle-fleet-system/fleet-controller-lock"
time="2022-07-22T21:57:22Z" level=error msg="processSingleItem: cattle-logging-system/logging-operator.logging.banzaicloud.io"
time="2022-07-22T21:57:22Z" level=error msg="processSingleItem: cattle-fleet-local-system/fleet-agent-lock"
time="2022-07-22T21:57:22Z" level=error msg="processSingleItem: kube-system/rke2"
time="2022-07-22T21:57:22Z" level=error msg="processSingleItem: kube-system/cattle-controllers"
time="2022-07-22T21:57:22Z" level=error msg="processSingleItem: kube-system/ingress-controller-leader"
time="2022-07-22T21:57:22Z" level=error msg="processSingleItem: kube-system/harvester-network-controllers"
time="2022-07-22T21:57:22Z" level=error msg="processSingleItem: kube-system/harvester-load-balancer"
time="2022-07-22T21:57:23Z" level=error msg="processSingleItem: cattle-fleet-system/gitjob"
time="2022-07-22T21:57:23Z" level=error msg="processSingleItem: kube-system/harvester-controllers"
time="2022-07-22T21:57:23Z" level=error msg="processSingleItem: harv1"

time="2022-07-22T21:57:24Z" level=error msg="processSingleItem: cattle-fleet-system/fleet-controller-lock"
time="2022-07-22T21:57:24Z" level=error msg="processSingleItem: cattle-logging-system/logging-operator.logging.banzaicloud.io"
time="2022-07-22T21:57:24Z" level=error msg="processSingleItem: cattle-fleet-local-system/fleet-agent-lock"
time="2022-07-22T21:57:24Z" level=error msg="processSingleItem: kube-system/rke2"
time="2022-07-22T21:57:24Z" level=error msg="processSingleItem: kube-system/cattle-controllers"
time="2022-07-22T21:57:24Z" level=error msg="processSingleItem: kube-system/harvester-network-controllers"
time="2022-07-22T21:57:24Z" level=error msg="processSingleItem: kube-system/harvester-load-balancer"
time="2022-07-22T21:57:25Z" level=error msg="processSingleItem: cattle-fleet-system/gitjob"
time="2022-07-22T21:57:25Z" level=error msg="processSingleItem: kube-system/harvester-controllers"
time="2022-07-22T21:57:26Z" level=error msg="processSingleItem: cattle-fleet-system/fleet-controller-lock"
time="2022-07-22T21:57:26Z" level=error msg="processSingleItem: cattle-logging-system/logging-operator.logging.banzaicloud.io"
time="2022-07-22T21:57:26Z" level=error msg="processSingleItem: cattle-fleet-local-system/fleet-agent-lock"
time="2022-07-22T21:57:26Z" level=error msg="processSingleItem: kube-system/rke2"
time="2022-07-22T21:57:26Z" level=error msg="processSingleItem: kube-system/cattle-controllers"
time="2022-07-22T21:57:26Z" level=error msg="processSingleItem: kube-system/harvester-network-controllers"
time="2022-07-22T21:57:26Z" level=error msg="processSingleItem: kube-system/harvester-load-balancer"
time="2022-07-22T21:57:27Z" level=error msg="processSingleItem: cattle-fleet-system/gitjob"
time="2022-07-22T21:57:27Z" level=error msg="processSingleItem: kube-system/harvester-controllers"
harv1:~ # 
```


### Log long-time processing events

```
func (c *controller) processSingleItem(obj interface{}) error {
	var (
		key string
		ok  bool
	)

	defer c.workqueue.Done(obj)

	if key, ok = obj.(string); !ok {
		c.workqueue.Forget(obj)
		log.Errorf("expected string in workqueue but got %#v", obj)
		return nil
	}
	t1 := time.Now()
	if err := c.syncHandler(key); err != nil {
		t2 := time.Now()
		delta := t2.Sub(t1).Milliseconds()
		if delta > 20 {
			log.Errorf("key %s used %v ms; res ERR", key, delta)
		}
		c.workqueue.AddRateLimited(key)
		return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
	}

		t2 := time.Now()
		delta := t2.Sub(t1).Milliseconds()
		if delta > 20 {
			log.Errorf("key %s used %v ms; res OK", key, delta)
		}

	c.workqueue.Forget(obj)
	return nil
}
```

There is no events taking long time.

```
harv1:~ # kk logs harvester-756cd66d6d-x4gm6 -n harvester-system | grep used
time="2022-07-22T21:36:13Z" level=error msg="key harvester-public/windows-iso-image-base-template used 27 ms; res OK"
time="2022-07-22T21:36:13Z" level=error msg="key harvester-public/iso-image-base-template used 29 ms; res OK"
time="2022-07-22T21:36:13Z" level=error msg="key harvester-public/raw-image-base-template used 37 ms; res OK"
time="2022-07-22T21:36:14Z" level=error msg="key server-version used 1690 ms; res OK"
time="2022-07-22T21:39:46Z" level=error msg="key cattle-monitoring-system/rancher-monitoring-admission-patch used 31 ms; res OK"
time="2022-07-22T21:39:52Z" level=error msg="key harv1 used 35 ms; res OK"
time="2022-07-22T21:40:01Z" level=error msg="key cattle-monitoring-system/rancher-monitoring-admission-patch used 22 ms; res OK"
time="2022-07-22T21:40:01Z" level=error msg="key cattle-monitoring-system/rancher-monitoring-patch-sa used 26 ms; res OK"
harv1:~ # 
```


## Look help from stacktrace

We know, go program has tons of go routine, traditional `kill -s SIGABRT` won't help

```
get strack trace log

https://pkg.go.dev/runtime

The GOTRACEBACK variable controls the amount of output generated when a Go program fails due to an unrecovered panic or an unexpected runtime condition. By default, a failure prints a stack trace for the current goroutine, eliding functions internal to the run-time system, and then exits with exit code 2. The failure prints stack traces for all goroutines if there is no current goroutine or the failure is internal to the run-time. 

GOTRACEBACK=none omits the goroutine stack traces entirely. 
GOTRACEBACK=single (the default) behaves as described above. 
GOTRACEBACK=all adds stack traces for all user-created goroutines. 
GOTRACEBACK=system is like “all” but adds stack frames for run-time functions and shows goroutines created internally by the run-time. 
GOTRACEBACK=crash is like “system” but crashes in an operating system-specific manner instead of exiting. 

For example, on Unix systems, the crash raises SIGABRT to trigger a core dump. For historical reasons, the GOTRACEBACK settings 0, 1, and 2 are synonyms for none, all, and system, respectively. The runtime/debug package's SetTraceback function allows increasing the amount of output at run time, but it cannot reduce the amount below that specified by the environment variable. See https://golang.org/pkg/runtime/debug/#SetTraceback.
```

### Enable stacktrace in Harvester go program

```
build_binary () {
    local BINARY="$1"
    local PKG_PATH="$2"

    CGO_ENABLED=0 GOTRACEBACK=all go build -ldflags "$LINKFLAGS $OTHER_LINKFLAGS"  -o "bin/${BINARY}" "${PKG_PATH}"
    if [ "$CROSS" = "true" ] && [ "$ARCH" = "amd64" ]; then
        GOOS=darwin go build -ldflags "$LINKFLAGS"  -o "bin/${BINARY}-darwin" "${PKG_PATH}"
        GOOS=windows go build -ldflags "$LINKFLAGS" -o "bin/${BINARY}-windows" "${PKG_PATH}"
    fi
}

build_binary "harvester" "."
build_binary "harvester-webhook" "./cmd/webhook"
```

Build and replace the image

### Trigger kill

Login into POD

```
kk exec -i -t -n harvester-system harvester-756cd66d6d-x4gm6 -- /bin/bash

harvester-756cd66d6d-x4gm6:/var/lib/harvester/harvester # ps aux
USER       PID %CPU %MEM    VSZ   RSS TTY      STAT START   TIME COMMAND
root         1  0.0  0.0   4564   724 ?        Ss   15:42   0:00 tini -- harvester
root         8  100  1.4 1179076 231988 ?      Sl   15:42   2:33 harvester
root        21  0.0  0.0  15116  4356 pts/0    Ss   15:44   0:00 /bin/bash
root        50  100  0.0  38184  4004 pts/0    R+   15:44   0:00 ps aux


when ps cmd is not there, use ls /proc and cat related process cmd line to check which process is the main one

```

Kill.

```
harvester-756cd66d6d-x4gm6:/var/lib/harvester/harvester # kill -s SIGABRT 8

or

kill -n 6 pid

```

Quickly scrap log

```
  kk logs -n harvester-system harvester-756cd66d6d-x4gm6 > /home/rancher/trace1.txt


update at 2022.10.19:

the best way is:

fetch the log from the NODE which is running the POD, the file path will be like

..# ls /var/log/pods/harvester-system_harvester-network-webhook-68b68f67df-2vbfz_0b2c58ae-e65e-4b35-bf1f-1a5f542ffde3/harvester-network-webhook/ -alth
total 340K
-rw-r----- 1 root root 3.0K Oct 19 15:34 2.log
drwxr-xr-x 2 root root 4.0K Oct 19 15:34 .
-rw-r----- 1 root root 325K Oct 19 15:34 1.log
drwxr-xr-x 3 root root 4.0K Oct 19 15:19 ..


Normally, there will be  2 files here, the big one will contain the stack trace
```
  

### Analyze the backtrace

There are big amout of goroutines, each has a backtrace like following.

We need to identify which one is the root cause.

```
goroutine 133 [chan receive, 3 minutes]:
github.com/harvester/harvester/vendor/github.com/rancher/lasso/pkg/controller.(*SharedHandler).Register.func1()
	/go/src/github.com/harvester/harvester/vendor/github.com/rancher/lasso/pkg/controller/sharedhandler.go:49 +0x65
created by github.com/harvester/harvester/vendor/github.com/rancher/lasso/pkg/controller.(*SharedHandler).Register
	/go/src/github.com/harvester/harvester/vendor/github.com/rancher/lasso/pkg/controller/sharedhandler.go:48 +0x245
	
	
goroutine 769 [select, 3 minutes]:
github.com/harvester/harvester/vendor/golang.org/x/net/http2.(*clientStream).writeRequest(0xc000d5ea80, 0xc000d6c200)
	/go/src/github.com/harvester/harvester/vendor/golang.org/x/net/http2/transport.go:1340 +0x9c9
github.com/harvester/harvester/vendor/golang.org/x/net/http2.(*clientStream).doRequest(0x0?, 0x0?)
	/go/src/github.com/harvester/harvester/vendor/golang.org/x/net/http2/transport.go:1202 +0x1e
created by github.com/harvester/harvester/vendor/golang.org/x/net/http2.(*ClientConn).RoundTrip
	/go/src/github.com/harvester/harvester/vendor/golang.org/x/net/http2/transport.go:1131 +0x30a

goroutine 170 [chan receive, 3 minutes]:
github.com/harvester/harvester/vendor/github.com/rancher/lasso/pkg/controller.(*SharedHandler).Register.func1()
	/go/src/github.com/harvester/harvester/vendor/github.com/rancher/lasso/pkg/controller/sharedhandler.go:49 +0x65
created by github.com/harvester/harvester/vendor/github.com/rancher/lasso/pkg/controller.(*SharedHandler).Register
	/go/src/github.com/harvester/harvester/vendor/github.com/rancher/lasso/pkg/controller/sharedhandler.go:48 +0x245



goroutine 113 [IO wait, 3 minutes]:
internal/poll.runtime_pollWait(0x7f120c574a18, 0x72)
	/usr/local/go/src/runtime/netpoll.go:302 +0x89
internal/poll.(*pollDesc).wait(0xc00052e000?, 0x0?, 0x0)
	/usr/local/go/src/internal/poll/fd_poll_runtime.go:83 +0x32
internal/poll.(*pollDesc).waitRead(...)
	/usr/local/go/src/internal/poll/fd_poll_runtime.go:88
internal/poll.(*FD).Accept(0xc00052e000)
	/usr/local/go/src/internal/poll/fd_unix.go:614 +0x22c
net.(*netFD).accept(0xc00052e000)
	/usr/local/go/src/net/fd_unix.go:172 +0x35
net.(*TCPListener).accept(0xc0005961e0)
	/usr/local/go/src/net/tcpsock_posix.go:139 +0x28
net.(*TCPListener).Accept(0xc0005961e0)
	/usr/local/go/src/net/tcpsock.go:288 +0x3d
net/http.(*Server).Serve(0xc000370000, {0x28aa358, 0xc0005961e0})
	/usr/local/go/src/net/http/server.go:3039 +0x385
net/http.(*Server).ListenAndServe(0xc000370000)
	/usr/local/go/src/net/http/server.go:2968 +0x7d
net/http.ListenAndServe(...)
	/usr/local/go/src/net/http/server.go:3222
github.com/harvester/harvester/pkg/cmd.initProfiling.func1()
	/go/src/github.com/harvester/harvester/pkg/cmd/app.go:83 +0x6f
created by github.com/harvester/harvester/pkg/cmd.initProfiling
	/go/src/github.com/harvester/harvester/pkg/cmd/app.go:82 +0x5d
```


[The collected full stack trace:](./resources/issue-2558-stack-trace.md)



When goroutine is in `sync`, or `select`, or `chan`, which is in a block state, lets exclude them.

There are only few goroutines left.

```
..$ ga "goroutine" | grep -v select | grep -v sync | grep -v cha

trace1.txt:goroutine 0 [idle]:
trace1.txt:goroutine 113 [IO wait, 3 minutes]:
trace1.txt:goroutine 32 [syscall, 3 minutes]:
trace1.txt:goroutine 48 [IO wait]:
trace1.txt:goroutine 9048 [IO wait]:

trace1.txt:goroutine 6139 [runnable]:

trace1.txt:goroutine 5872 [IO wait]:
trace1.txt:goroutine 6140 [IO wait, 3 minutes]:

trace1.txt:goroutine 9033 [runnable]:
```

### The root cause

Clearly, the `6139 [runnable]` looks special.

It's stack.

```
goroutine 6139 [runnable]:
runtime.Gosched(...)
	/usr/local/go/src/runtime/proc.go:317
github.com/harvester/harvester/vendor/github.com/rancher/dynamiclistener.(*listener).WrapExpiration.func1()
	/go/src/github.com/harvester/harvester/vendor/github.com/rancher/dynamiclistener/listener.go:168 +0x59
created by github.com/harvester/harvester/vendor/github.com/rancher/dynamiclistener.(*listener).WrapExpiration
	/go/src/github.com/harvester/harvester/vendor/github.com/rancher/dynamiclistener/listener.go:165 +0xb8
```

The source code is as such, a `spin-lock` similar behaviour, but it does not end as expected as only in a very short time.


```
func (l *listener) WrapExpiration(days int) net.Listener {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		// busy-wait for certificate preload to complete
		for l.cert == nil {
			runtime.Gosched()  /// FIXME, here
		}
```

We should try avoid such code in user-land program. It's hard to make sure that it can always end loop in expected time.


