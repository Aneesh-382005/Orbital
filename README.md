# Orbital

[Repo](https://github.com/Aneesh-382005/Orbital) · [Architecture, failure scenarios, threat model](docs/ARCHITECTURE.md)

Orbital creates on-demand, per-user cloud IDEs and tracks their state so nothing runs unaccounted for.

Quick Start and the failure scenarios in [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) were checked by hand against a running Kind cluster at commit [`86f8479`](https://github.com/Aneesh-382005/Orbital/commit/86f8479fa93599f60d13018efa672ea89e185712), the last commit that changed code, manifests, or the Dockerfile. Nothing since then has changed what those checks cover. No CI runs this yet, so treat future code changes as unverified until that's in place. See the Roadmap.

## What it does

You call one endpoint with a name. Orbital starts a [code-server](https://github.com/coder/code-server) container (VS Code in the browser), assigns it a port and a password, and returns a link. You stop or delete it with separate calls.

Starting the container is one Docker API call. Staying accurate about what exists takes more work: which container belongs to which user, which ports are free, and what to do when the database and Docker disagree. That happens when a container dies without telling anyone, or when a container exists that nothing claims anymore.

## How it's built

One Go binary, four parts:

- **API**: handles create, list, stop, delete.
- **Database**: SQLite file. Tracks every workspace and every port in use.
- **Provisioner**: the only part that talks to Docker. Creates and removes containers.
- **Reconciler**: checks every 30 seconds that the database and Docker agree, and fixes it when they don't.

Instead of a queue or a scheduler, one process runs the loop directly.

## Quick Start

Needs Docker, a running [Kind](https://kind.sigs.k8s.io/) cluster, `kubectl`, and Go 1.25+. Tested start to finish in under 5 minutes.

```bash
git clone https://github.com/Aneesh-382005/Orbital orbital && cd orbital

# 1. Generate the keys Orbital uses to sign login tokens
mkdir -p secrets
openssl genpkey -algorithm RSA -out secrets/private.pem -pkeyopt rsa_keygen_bits:2048
openssl rsa -pubout -in secrets/private.pem -out secrets/public.pem

# 2. Build the image
docker build -t orbital:local .

# 3. Load it into Kind
kind load docker-image orbital:local --name orbital

# 4. Replace the placeholder secret with your own keys
kubectl create secret generic orbital-jwt-keys -n orbital \
  --from-file=private.pem=secrets/private.pem \
  --from-file=public.pem=secrets/public.pem \
  --dry-run=client -o yaml > deploy/k8s/02-secret.yaml

# 5. Deploy
kubectl apply -f deploy/k8s/
kubectl rollout status deployment/orbital -n orbital --timeout=90s

# 6. Verify
kubectl port-forward -n orbital svc/orbital 8080:80 &
curl http://localhost:8080/health

TOKEN=$(curl -s -X POST http://localhost:8080/auth/token \
  -H 'Content-Type: application/json' -d '{"user_id":"demo"}' | jq -r .token)

curl -s -X POST http://localhost:8080/api/v1/workspaces \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"name":"my-first-workspace"}'
```

The last call needs Orbital to reach Docker on the node. On a fresh Kind cluster that usually needs one extra permission step, covered in [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md).

## What's Not Finished

- **Runs as one instance only.** SQLite does not handle two writers safely. Orbital cannot scale to multiple replicas until the database changes.
- **Login has no identity check.** Anyone can request a token for any user ID. There's no password or verification behind it.
- **Workspace links use plain HTTP.** Fine for local testing, not for the open internet.
- **Orbital needs deep Docker access to work**, which gives it more power over the host than a shared environment should allow.
- **No cap on workspaces per user.**
- **No automated tests.** Checked by hand against a running Kind cluster, not by CI.

More detail, and how each of these could be fixed, is in [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md).

## Roadmap

- Postgres instead of SQLite, so Orbital can run more than one replica
- OIDC login, instead of the current open token endpoint
- Restrict what workspace containers can reach on the network
- A Helm chart, instead of hardcoded namespace and image values
- Auto-stop idle workspaces
- Measure cold-start vs warm-start time
- CI that runs the Quick Start and failure scenarios on every push, instead of relying on a manually verified commit

Issues and PRs welcome.
