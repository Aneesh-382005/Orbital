# Orbital: How It's Built, What Breaks, and What's Missing

The deep-dive: internals, failure modes, and production readiness. For setup, see the [README](../README.md).

The failure scenarios below were reproduced by hand against a running Kind cluster at commit [`86f8479`](https://github.com/Aneesh-382005/Orbital/commit/86f8479fa93599f60d13018efa672ea89e185712), the last commit that changed code, manifests, or the Dockerfile. This is a point-in-time check, not a continuous one, and no CI verifies these on future code changes yet.

## How the pieces fit together

Orbital is one program, split into packages by responsibility:

- **`cmd/api`** starts the process. Reads a few optional settings (database path, listen address, reconcile interval), wires the pieces together, shuts down cleanly on SIGINT/SIGTERM.
- **`internal/models`** defines the `Workspace` type: name, owner, status, port. One shared shape, so every package agrees on it.
- **`internal/db`** and **`internal/store`** own the database. Two tables: workspaces and ports. A deleted workspace's name can be reused immediately. Instead of checking this in application code, a database constraint enforces it, so two requests can't race into a duplicate.
- **`internal/auth`** issues and checks login tokens (JWTs). Tokens are signed with a private key and checked with the matching public key. The check step also verifies *how* the token was signed, not just that it has a valid signature. Instead of trusting whatever signing method a token claims to use, Orbital only accepts one method (RSA). Skipping this check is a common way JWT systems get exploited.
- **`internal/provisioner`** is the only package that talks to Docker. It pulls the code-server image, picks a free port, generates a random password, and starts the container with a hard memory and CPU limit. One workspace cannot slow down or crash the others.
- **`internal/reconciler`** runs on a 30-second loop. It compares what the database thinks is running against what Docker actually has running, in both directions. If a container died silently, the workspace gets marked as errored. If a container exists that the database no longer claims, it gets removed.
- **`internal/metrics`** exposes counters and timers (requests handled, workspace start time, workspace count by status, how stale the last reconcile check is) in a format Prometheus can read.
- **`internal/api`** is the HTTP layer: routes, the auth check that runs before protected routes, and a timer on every request.

## What happens when things go wrong

**The login keys are missing or broken.**
Orbital refuses to start. Deploying the placeholder key file that ships in this repo, without replacing it, produces an immediate crash with a clear error about the key being unreadable. Kubernetes keeps restarting the container until working keys are provided. Fix: generate your own keys and replace the secret, as shown in the Quick Start.

**Orbital can't reach Docker.**
Startup succeeds, since Orbital doesn't check for Docker until it's needed. The failure shows up on the first workspace creation call, which returns an error, and the 30-second health check starts failing silently in the logs. On a fresh Kind cluster, Docker refuses the connection because Orbital's user isn't in the right permission group. Fix: grant that access explicitly. Documented in the Kubernetes deployment file.

**Two Orbital instances write to the same database at once.**
SQLite handles one writer safely. Two processes writing the same file can produce errors, or in the worst case, damage the database. This is why Orbital runs as a single instance today. Instead of a config change, fixing this needs a networked database.

**A workspace is marked running, but its container is gone.**
Orbital notices within 30 seconds and marks it errored. Nobody gets notified proactively, the status just changes on the next check. This is deliberate: automatically recreating it without asking could restart something a user stopped on purpose. Fix today is manual: delete it and create a new one.

**A container exists that the database doesn't know about.**
Orbital removes it on the next check, with no confirmation. This mostly cleans up containers that should never have existed, but there is a timing gap: if a workspace was just created and the database write hasn't landed yet, a check landing in that exact window can delete a container seconds old and working correctly. This is a known, unresolved race.

**A port is reserved, but the container fails to start.**
Ports are marked taken before the container is created, to stop two requests from grabbing the same one. If the container then fails to start, that reservation is not released. Enough failures exhaust the port range and require manual cleanup.

**Orbital's pod restarts or moves to a different node.**
State lives in the database file, on a disk separate from the running process, so a restart doesn't lose data. The one gap: if the pod restarts on a different node than the one holding its disk, it can get stuck waiting to reconnect. This is a limitation of a single-node local Kind cluster more than the design itself.

**Docker runs low on memory or CPU.**
Each workspace has a hard resource ceiling, so it can't take over the host on its own. But there's no limit on how many creation requests can happen at once. Enough concurrent requests can still exhaust the host, and further requests fail until it recovers.

## Production readiness, plainly

**Secrets.** Login keys sit in a Kubernetes Secret, which is base64-encoded, not encrypted, unless the cluster is separately configured for that. There's no key rotation without a restart. Production needs a dedicated secrets manager.

**Building and shipping the image.** Today it's a manual `docker build` with no vulnerability scan and no signature. Production needs a pipeline that builds it automatically, scans it, and signs it before deployment.

**Service-to-service trust.** Doesn't apply yet, since only one instance runs. If that changes, there is currently no way for services to verify each other beyond the end-user login token.

**Handling load.** No rate limiting, no queue. A burst of requests becomes a burst of simultaneous Docker calls, bounded only by what the host can handle before failing.

**Multiple regions or zones.** Not supported. Everything depends on one machine with direct Docker access, which doesn't translate to a spread-out setup without significant changes.

**Cost.** Each workspace has a fixed resource ceiling, so the maximum cost per workspace is known. Nothing shuts down an idle workspace, so cost tracks forgotten workspaces, not actual use.

**Data location.** Wherever the disk is, with no replication and no way to guarantee data stays in a region.

**Running this in production.** Partially there. The exposed metrics (requests, workspace counts, reconcile staleness) are the right signals. At commit `86f8479`, the staleness metric climbed correctly for the entire time the Docker connection was broken, confirming it behaves as intended during an outage. Missing: alerts, dashboards, and runbooks built on top of those metrics.

## What Orbital defends against, and what it doesn't

**Defends against:**

- **Forged login tokens.** The check step verifies the signing method, not just the signature. This closes a known JWT attack where a token is signed with the wrong key type and a lenient verifier accepts it anyway.
- **One user reaching another user's workspace.** Every request touching a specific workspace checks ownership fresh, every time.
- **One workspace overwhelming the host.** Hard resource limits mean a runaway process in one workspace only affects that workspace.
- **A leaked password affecting more than one workspace.** Every workspace gets its own random password, not a shared one.

**Does not defend against, on purpose, for now:**

- **Anyone getting a token for any user.** There's no identity check behind token issuance. It's a placeholder for a proper login system. This is the biggest gap and the first item on the roadmap.
- **The trust Orbital places in the host.** To manage containers, Orbital needs deep Docker access, which is more power than is safe to hand to something running in a shared environment. This is a known trade-off for local development, not something acceptable long-term.
- **What a user does inside their own workspace.** Once inside code-server, it's a full dev environment with a terminal and open network access. Orbital's job ends at getting the right user the right, resource-limited workspace.
- **Error messages that leak internal detail.** Failures can return details that shouldn't be user-facing. Not a security defense, just a known rough edge.
