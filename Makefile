SHELL := /bin/bash -o pipefail
KUBECTL := kubectl --context k3d-cluster
CLUSTER_NAME := cluster

# Track state using hidden files
CLUSTER_CREATED := .cluster-created
DB_SECRETS_CREATED := .db-secrets-created
IMAGES_IMPORTED := .images-imported

.PHONY: all
.PHONY: clean
.PHONY: create-k3d-cluster
.PHONY: delete-local-kube-cluster
.PHONY: build-server
.PHONY: import-images
.PHONY: generate-db-secrets
.PHONY: deploy-application

all: deploy-application

clean:
	@rm -f $(CLUSTER_CREATED) $(DB_SECRETS_CREATED) $(IMAGES_IMPORTED)
	@echo "Cleaned up state files"

$(CLUSTER_CREATED):
	@which k3d > /dev/null || (echo "k3d must be installed to create local Kubernetes cluster\n==> visit https://k3d.io/ \n\n wget -q -O - https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash" && exit 1)
	@echo "Creating Kubernetes cluster..."
	@k3d cluster create $(CLUSTER_NAME) \
		--k3s-arg '--disable=servicelb@server:0' \
		--k3s-arg '--disable=traefik@server:0' \
		--agents 2
	@touch $(CLUSTER_CREATED)

create-k3d-cluster: delete-local-kube-cluster $(CLUSTER_CREATED)

delete-local-kube-cluster:
	@echo "Deleting existing Kubernetes cluster..."
	@k3d cluster delete $(CLUSTER_NAME) 2>/dev/null || true
	@rm -f $(CLUSTER_CREATED) $(DB_SECRETS_CREATED) $(IMAGES_IMPORTED)

build-server:
	docker build -t server:latest .

$(IMAGES_IMPORTED): build-server $(CLUSTER_CREATED)
	k3d image import server:latest --cluster $(CLUSTER_NAME)
	@touch $(IMAGES_IMPORTED)

import-images: $(IMAGES_IMPORTED)

$(DB_SECRETS_CREATED): $(IMAGES_IMPORTED)
	$(KUBECTL) delete secret db-secrets --ignore-not-found
	$(KUBECTL) create secret generic db-secrets \
		--from-literal=POSTGRES_DB=transaction_store \
		--from-literal=POSTGRES_USER=transaction_store_user \
		--from-literal=POSTGRES_PASSWORD=transaction_store_password
	@touch $(DB_SECRETS_CREATED)

generate-db-secrets: $(DB_SECRETS_CREATED)

deploy-application: $(DB_SECRETS_CREATED)
	$(KUBECTL) create namespace observability --dry-run=client -o yaml | $(KUBECTL) apply -f -
	$(KUBECTL) apply \
		-f manifests/server \
		-f manifests/postgres \
		-f observability/manifests
	@echo "Transaction server deployed to Kubernetes cluster"
