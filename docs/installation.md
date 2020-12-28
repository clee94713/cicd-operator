# Installation Guide

This guides to install CI/CD operator. The contents are as follows.

* [Prerequisites](#prerequisites)
* [Installing CI/CD Operator](#installing-cicd-operator)
* [Enabling email feature](#enabling-email-feature)

## Prerequisites
- [Install Tekton Pipelines](https://github.com/tektoncd/pipeline/blob/master/docs/install.md) (at least v0.19.0)

## Installing CI/CD Operator
1. Run the following command to install CI/CD operator  
   ```bash
   VERSION=v0.1.0
   kubectl apply -f https://raw.githubusercontent.com/cqbqdd11519/cicd-operator/$VERSION/config/release.yaml
   ```
2. Enable `CustomTask` feature, disable `Affinity Assistant`
   ```bash
   kubectl -n tekton-pipelines patch configmap feature-flags \
   --type merge \
   -p '{"data": {"enable-custom-tasks": "true", "disable-affinity-assistant": "true"}}'
   ```

## Enabling email feature
**You need an external SMTP server**
1. Run the following command to create basic-auth secret for SMTP server
   ```bash
   SMTP_USERNAME=<SMTP Username>
   SMTP_PASSWORD=<SMTP Password>
   kubectl -n cicd-system create secret generic cicd-smtp \
   --type='kubernetes.io/basic-auth' \
   --from-literal=username=$SMTP_USERNAME \
   --from-literal=password=$SMTP_PASSWORD
   ```
2. Run the following command to enable email-feature and configure SMTP server information
   ```bash
   SMTP_HOST=<SMTP server HOST:PORT>
   k -n cicd-system patch configmaps cicd-config \
   --type merge \
   -p "{\"data\":{\"enableMail\":\"true\",\"smtpHost\":\"$SMTP_HOST\",\"smtpUserSecret\":\"cicd-smtp\"}}"
   ```