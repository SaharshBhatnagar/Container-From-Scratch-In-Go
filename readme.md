# Container From Scratch in Go

A minimal Linux container runtime engineered from scratch in Go using native Linux kernel system calls and virtual filesystems. This project demonstrates the core kernel primitives that power modern containerization engines like Docker and `runc`—specifically Linux namespaces, `chroot` filesystem isolation, private mount propagation, and Control Groups (cgroups v2)—using only the Go standard library.

---

## System Architecture & Execution Flow

### 1. Two-Phase Self-Execution (`/proc/self/exe`)
To ensure the target process initializes inside the isolated namespaces rather than the host context, the runtime uses a two-stage bootstrap pattern:
* **Parent Stage (`run`):** Parses user arguments, configures kernel namespace flags on the child process descriptor (`syscall.SysProcAttr`), and executes `/proc/self/exe` (a symlink to the currently running binary in memory) with the internal `child` subcommand.
* **Child Stage (`child`):** Executes inside the newly cloned kernel namespaces. It sets the container hostname, configures and attaches cgroup v2 resource constraints on the host hierarchy, pivots the root filesystem into `alpine-fs`, mounts `/proc`, and executes the target command (`/bin/sh`).

## Prerequisites & Rootfs Setup

### Environment Requirements
 * Operating System: Linux or Windows Subsystem for Linux (WSL2 with Ubuntu) running Linux Kernel 5.8+ with cgroup v2 unified hierarchy enabled.
 * Toolchain: Go (amd64 Linux toolchain, zero external dependencies).
 * Privileges: Root privileges (sudo) are required to invoke clone, chroot, mount, and write to /sys/fs/cgroup.
 
 1. Clone the Repository
 ```Bash
 git clone https://github.com/SaharshBhatnagar/Container-From-Scratch-In-Go.git
cd Container-From-Scratch-In-Go
```
2. Provision the Alpine Linux Root Filesystem
 > Download and extract the minimal Alpine Linux root filesystem (~3MB) into alpine-fs/:
```Bash
mkdir -p alpine-fs
wget https://dl-cdn.alpinelinux.org/alpine/v3.18/releases/x86_64/alpine-minirootfs-3.18.4-x86_64.tar.gz
tar -C alpine-fs -xf alpine-minirootfs-3.18.4-x86_64.tar.gz
rm alpine-minirootfs-3.18.4-x86_64.tar.gz
```

### Building & Running the Runtime

#### Option 1: Compile Standalone Binary (Recommended)
Compile the runtime as your standard user and execute the binary with sudo:
```Bash
go build -o container main.go
sudo ./container run /bin/sh
```
#### Option 2: Run Directly via Go Toolchain
```Bash
sudo go run main.go run /bin/sh
```

### Benchmarks & Verification Suite1.
1. Verify PID, Hostname, and Filesystem Isolation
Inside the container shell (/ #), run:
```Bash
hostname
ps
ls /
```
* Expected Output:
```Plaintext/ 
# hostname
my-container
/ # ps
PID   USER     TIME  COMMAND
    1 root      0:00 /proc/self/exe child /bin/sh
    7 root      0:00 /bin/sh
    8 root      0:00 ps
/ # ls /
bin    dev    etc    home   lib    media  mnt    opt    proc   root   run    sbin   srv    sys    tmp    usr    var
```

2. Verify Mount Namespace Isolation
While the container is active in Terminal 1, inspect the mount table inside the container vs. on the Ubuntu host:
* Inside Container (/ #):
```Bash
mount
# Output: proc on /proc type proc (rw,relatime)
```
* On Host Terminal:
```Bash
mount | grep alpine-fs
# Output: (empty - private mount propagation prevents host table leakage)
```

3. Verify Cgroup v2 Fork Bomb Defense (pids.max = 20)
Attempt to spawn 30 background processes inside the container:
```Bash
for i in $(seq 1 30); do sleep 100 & done
```
* Expected Result: Around the 20th process (including the parent init and shell processes), the Linux kernel blocks fork() with:
```Plaintext
/bin/sh: can't fork: Resource temporarily unavailable
```

4. Verify Cgroup v2 Memory Hard Cap & OOM Killer (memory.max = 20M)
> Spawn a child shell inside the container that exponentially doubles an in-memory string until it crosses the 20M RAM ceiling:
```Bash
sh -c 'x="a"; while true; do x="$x$x"; done'
```
Expected Result:
* Inside Container: The kernel immediately terminates the runaway child process and outputs Killed, while keeping the main /bin/sh container session responsive.
* On Host Terminal: Inspect the kernel OOM counter before exiting the container:
```Bash
cat /sys/fs/cgroup/my-container/memory.events
# Confirms oom and oom_kill counters incremented
```

5. Verify Crash-Resilient Resource Teardown (defer)
Force an abnormal non-zero exit inside the container shell:
```Bash
exit 1
```
Then verify on the host machine that both the /proc mount and the /sys/fs/cgroup/my-container directory were automatically cleaned up during stack unwinding:
```Bash
mount | grep alpine-fs
ls /sys/fs/cgroup/my-container
# Expected: ls: cannot access '/sys/fs/cgroup/my-container': No such file or directory
```

## Project Structure

```Plaintext

Container-From-Scratch-In-Go/
├── alpine-fs/       
├── container        
├── .gitignore       
├── go.mod           
├── main.go          
└── README.md        

```