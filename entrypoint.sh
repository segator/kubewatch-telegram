#!/bin/sh
KUBECONFIG=""
if [ ! "$K8S_SERVICE" = true ]; then
  KUBECONFIG="--kubeconfig=/root/.kube/config"
fi

kubewatch --telegramapi=${TELEGRAM_API} --telegramgroup=${TELEGRAM_GROUPID} ${KUBECONFIG} ${RESOURCES}
