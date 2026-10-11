# API flows (Sleuth)

End-to-end API flows run by [Sleuth](https://github.com/ruth-tirusew/sleuth). Every call is checked against
`contracts/openapi.yaml`; CI runs them on every pull request (`.github/workflows/sleuth-api.yml`) and publishes JUnit
XML and Sleuth's evidence files as the `sleuth-results` artifact.

| Flow | What it covers |
|---|---|
| `public-browse.flow.yaml` | health, articles, artists, posts, events and search, opening one of each, and a 404 |
| `auth.flow.yaml` | register, sign in, read the current user with the session, and the refusals (no session, wrong password, duplicate address) |

Run them against a local API (`make up && make api`):

    export MQ_BASE_URL=http://localhost:8080
    export MQ_TEST_EMAIL="me-$(date +%s)@example.test"   # the auth flow registers it, so it must be new
    export MQ_TEST_PASSWORD="$(openssl rand -hex 16)"
    sleuth run -env env.yaml public-browse.flow.yaml
    sleuth run -env env.yaml -parallel 4 auth.flow.yaml

Exit codes: 0 passed, 1 an assertion or contract failed, 2 a flow or setup error, 3 the API could not be reached.
