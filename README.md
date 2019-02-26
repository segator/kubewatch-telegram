# kubewatch-telegram

[![Version Widget]][Version] [![License Widget]][License] [![GoReportCard Widget]][GoReportCard] [![Travis Widget]][Travis] [![DockerHub Widget]][DockerHub]

[Version]: https://github.com/prg3/kubewatch-telegram/releases
[Version Widget]: https://img.shields.io/github/release/prg3/kubewatch-telegram.svg?maxAge=60
[License]: http://www.apache.org/licenses/LICENSE-2.0.txt
[License Widget]: https://img.shields.io/badge/license-APACHE2-1eb0fc.svg
[GoReportCard]: https://goreportcard.com/report/prg3/kubewatch-telegram
[GoReportCard Widget]: https://goreportcard.com/badge/prg3/kubewatch-telegram
[Travis]: https://travis-ci.org/prg3/kubewatch-telegram
[Travis Widget]: https://travis-ci.org/prg3/kubewatch-telegram.svg?branch=master
[DockerHub]: https://hub.docker.com/r/majestik/kubewatch-telegram
[DockerHub Widget]: https://img.shields.io/docker/pulls/majestik/kubewatch-telegram.svg

Kubernetes API event watcher with output to Telegram based on code from https://github.com/softonic/kubewatch.

##### Install

```
go get -u github.com/prg3/kubewatch-telegram
```

##### Shell completion

```
eval "$(kubewatch --completion-script-${0#-})"
```

##### Help

```
kubewatch --help
usage: kubewatch --telegramapi=TELEGRAMAPI --telegramgroup=TELEGRAMGROUP [<flags>] <resources>...

Watches Kubernetes resources via its API.

Flags:
  -h, --help                     Show context-sensitive help (also try --help-long and --help-man).
      --kubeconfig=/Users/prg3/.kube/config
                                 Absolute path to the kubeconfig file.
      --namespace=""             Set the namespace to be watched.
      --flatten                  Whether to produce flatten JSON output or not.
      --telegramapi=TELEGRAMAPI  API Key for Telegram bot
      --telegramgroup=TELEGRAMGROUP
                                 Group that the bot should post to. Note that
                                 Telegram groups are negative values, but drop
                                 the - here. If you wish to message an
                                 individual, you will need to add a negative on
                                 the command line
      --version                  Show application version.

Args:
  <resources>  Space delimited list of resources to be watched.
```

##### Out-of-cluster examples:

Not including required --telegramapi and --telegramgroup in examples for cleanliness.

Watch for `pods` and `events` in all `namespaces`:
```
kubewatch pods events | jq '.'
```

Same thing with docker:
```
docker run -it --rm \
-v ~/.kube/config:/root/.kube/config \
prg3/kubewatch pods events | jq '.'
```

Watch for `services` events in namespace `foo`:
```
kubewatch --namespace foo services | jq '.'
```

Same thing with docker:
```
docker run -it --rm \
-v ~/.kube/config:/root/.kube/config \
prg3/kubewatch --namespace foo services | jq '.'
```

##### In-cluster examples:

Run `kubewatch` in the `monitoring` namespace and watch for `pods` in all namespaces:
```
kubectl --namespace monitoring run kubewatch --image prg3/kubewatch -- pods
```

Run `kubewatch` in the `monitoring` namespace and watch for `pods`, `deployments` and `events` objects in all namespaces. Also flatten the `json` output:
```
kubectl --namespace monitoring \
run kubewatch --image prg3/kubewatch \
-- --flatten pods deployments events
```
