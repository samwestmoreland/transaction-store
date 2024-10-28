SHELL := /bin/bash -o pipefail
KUBECTL := kubectl --context k3d-cluster

.PHONY: create-k3d-cluster
.PHONY: delete-local-kube-cluster
.PHONY: build-server
.PHONY: import-images
.PHONY: generate-db-secrets
.PHONY: deploy-application

create-k3d-cluster: delete-local-kube-cluster
	@which k3d >> /dev/null || echo "k3d must be installed to create local Kubernetes cluster\n==> visit https://k3d.io/ \n\n wget -q -O - https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash" \
	&& k3d cluster create cluster --k3s-arg '--disable=servicelb@server:0' --k3s-arg '--disable=traefik@server:0' --agents 2

delete-local-kube-cluster:
	@echo "Deleting existing Kubernetes cluster..." && k3d cluster delete cluster

build-server:
	docker build -t server:latest .

import-images: build-server create-k3d-cluster
	k3d image import server:latest --cluster cluster

generate-db-secrets: create-k3d-cluster
    ${KUBECTL} create secret generic db-secrets \
        --from-literal=POSTGRES_DB=transaction_store \
        --from-literal=POSTGRES_USER=transaction_store_user \
        --from-literal=POSTGRES_PASSWORD=transaction_store_password

deploy-application: import-images generate-db-secrets
	${KUBECTL} create namespace observability \
	&& ${KUBECTL} create \
	  -f manifests/server \
      -f manifests/postgres \
	  -f observability/manifests \
	&& echo "Transaction server deployed to Kubernetes cluster"
