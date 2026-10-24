ARG BASE_IMAGE=local/l4d2-manager-next:custom-v1.12.7-20260826.1
FROM ${BASE_IMAGE}

COPY backend/static /static
