SHELL := /bin/bash -o pipefail
CLUSTER_NAME := peaceful-beaver
KUBECTL := kubectl --context k3d-$(CLUSTER_NAME)

# Track state using hidden files
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
	@rm -f $(DB_SECRETS_CREATED) $(IMAGES_IMPORTED)
	@echo "Cleaned up state files"

build-server:
	docker build -t server:latest .

$(IMAGES_IMPORTED): build-server
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
		-f k8s/manifests/server \
		-f k8s/manifests/postgres \
		-f k8s/manifests/observability
	@echo "Transaction server deployed to Kubernetes cluster"
