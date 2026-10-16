# Container From Scratch in Go (Still Developing)

A lightweight Linux container runtime built from scratch in Go using native Linux kernel system calls. This project explores the low-level primitives that power containerization—specifically Linux namespaces, `chroot` filesystem isolation, and Control Groups (cgroups)—without relying on third-party container libraries.

**Status:** Work in Progress (Core Isolation & Filesystem Jail Implemented)

---

## Architecture & Implemented FeaturesV

### 1. Two-Phase Self-Execution (`/proc/self/exe`)
To configure kernel namespaces before executing the user's target process, the runtime uses a two-stage execution pipeline:
* **Parent Process (`run`):** Requests new kernel namespaces via `Cloneflags` (`syscall.SysProcAttr`) and re-executes the current binary via `/proc/self/exe` passing the `child` argument.
* **Child Process (`child`):** Initializes inside the newly created namespaces, configures the isolated environment (hostname, root filesystem, and `/proc` mount), and executes the target command.

### 2. UTS Namespace Isolation (`CLONE_NEWUTS`)
* Isolates the hostname and domain name inside the container (`my-container`) using `syscall.Sethostname`, preventing modifications from affecting the host machine.

### 3. PID Namespace Isolation (`CLONE_NEWPID`)
* Isolates the process ID number space so the container's initialization process runs as `PID 1` inside the sandbox while maintaining a standard unprivileged PID on the host kernel.

### 4. Filesystem Isolation (`chroot`) & `/proc` Lifecycle
* **Root Filesystem Jail:** Uses `syscall.Chroot` and `syscall.Chdir("/")` to lock the container process inside a minimal Alpine Linux root filesystem (`alpine-fs`), preventing traversal into the host filesystem.
* **Process Visibility:** Dynamically mounts the `/proc` pseudo-filesystem (`syscall.Mount`) inside the container root so internal utilities like `ps` can query the kernel for isolated process tables, and cleanly unmounts it (`syscall.Unmount`) upon container exit.

---

## Roadmap (In Progress)

* [ ] **Mount Namespace Isolation (`CLONE_NEWNS`):** Isolate mount points from the host mount table using private mount propagation (`MS_PRIVATE | MS_REC`).
* [ ] **Control Groups (cgroups):** Implement resource constraints to restrict memory allocation and cap maximum spawned processes (`pids.max`) to defend against host resource exhaustion (e.g., fork bombs).

---

## Prerequisites & Setup

### Environment Requirements
* **OS:** Linux or Windows Subsystem for Linux (WSL2 with Ubuntu)
* **Compiler:** Go (Linux `amd64` toolchain)
* **Privileges:** Root privileges (`sudo`) are required to execute namespace and mount system calls.

### Preparing the Container Root Filesystem
Before running the container, download and extract the minimal Alpine Linux root filesystem in the project root directory:

```bash
mkdir -p alpine-fs
wget https://dl-cdn.alpinelinux.org/alpine/v3.18/releases/x86_64/alpine-minirootfs-3.18.4-x86_64.tar.gz
tar -C alpine-fs -xf alpine-minirootfs-3.18.4-x86_64.tar.gz
rm alpine-minirootfs-3.18.4-x86_64.tar.gz

```

## Usage & Verification
Start an interactive shell inside the isolated container:

```Bash
sudo go run main.go run /bin/sh
```

### Verification Checks

1. Verify Process Isolation:

Run `ps` inside the container shell to confirm the container init process is running as PID 1 and host processes are hidden:

```Plaintext
/ # ps
PID   USER     TIME  COMMAND
    1 root      0:00 /proc/self/exe child /bin/sh
    7 root      0:00 /bin/sh
    8 root      0:00 ps
```

2. Verify Filesystem Jail:

Run `ls /` to confirm visibility is restricted to the alpine-fs root directory.

3. Verify Mount Cleanup:

Type `exit` to terminate the container, then run `mount | grep proc` on the host machine to verify that alpine-fs/proc was cleanly unmounted.