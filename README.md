# Skiff 🚢

**Skiff** is a lightweight, extensible CLI for managing Terraform + Terragrunt infrastructure at scale — across multiple AWS accounts, regions, and services.

Inspired by Kubernetes-style configuration and GitOps practices, Skiff provides structure, reusability, and full customization through templated strategies.

---

## 🚀 Features

- ⚙️ **Modular Service Types** – Define reusable services like VPC, RDS, IAM
- 🧾 **Account Manifests** – Describe AWS accounts and services declaratively
- 🛠️ **Template-based Terragrunt File Generation**
- 🧭 **Strategy-Driven Folder Layouts** – Flexible, configurable directory structures
- 🏃 **Command Runner** – Run `terragrunt` commands filtered by labels
- 🔁 **Dry-Run Support** – Preview changes without executing them
- 🧼 **Path Sanitization** – Clean, consistent folder output

---

## 📁 Folder Structure

```console
.
├── .skiff              # Project config (includes strategy)
├── manifests/          # Service types + account definitions
│   ├── service-types.yaml
│   └── 123456789012.yaml
├── templates/          # Terragrunt templates
│   └── terragrunt.default.tmpl
└── terragrunt/         # Output folder (rendered files)
```

---

## 🧠 Strategy-Driven Layout

Skiff uses Go templates defined inside `.skiff` to determine where to place `terragrunt.hcl` files.

```yaml
strategy:
  name: terragrunt
  description: Account-based layout with global and regional separation
  template: |
    {{- var.env }}/{{ var.account_id }}/
    {{ if eq var.scope "global" }}
      global/{{ var.service }}
    {{ else }}
      regions/{{ .region }}/{{ var.group }}/{{ var.service }}
    {{ end }}
```

Variables provided per service:

- `account_id`
- `account_name` Name of the account
- `service` (name of the service)
- `scope` (global or regional)
- `region`,
- `group`,
- all keys in metadata and labels

---

## 📌 Example Manifests

### service-types.yaml

```yaml
apiVersion: v1
types:
  network:
    source: github.com/terraform-aws-modules/terraform-aws-vpc
    group: networks
    version: v1.0.0
  database:
    source: github.com/terraform-aws-modules/terraform-aws-rds
    group: databases
    version: v1.0.0
```

### account_id.yaml

```yaml
apiVersion: v1
account:
  name: dev
  id: "123456789012"
metadata:
  app: user
  env: prod
services:
  - name: iam-root
    type: iam
    scope: global
    labels:
      app: shared
      env: prod
    inputs:
      account_id: "123456789012"

```

---

## 🗺️ Diagram: Flow Overview

```text
+-------------+           +----------------------+         +-------------------------+
| .skiff file |  ----->   |  Load Strategy + Vars |  --->  | Render Folder Structure |
+-------------+           +----------------------+         +-------------------------+
                                                                 |
                                                                 v
                                                          terragrunt.hcl files
```

---

## 🧪 Example Workflow

```bash
# Initialize a new Skiff project
skiff init

# Add a service type
skiff add service-type --name network --type vpc --source github.com/terraform-aws-modules/terraform-aws-vpc

# Add an account
skiff add account --name dev --id 123456789012

# Add a service to the account
skiff add service --account 123456789012 --name user-vpc --type network --region us-east-1 --labels app=user

# Generate folders and terragrunt.hcl files
skiff generate

# Plan infrastructure only for app=user
skiff plan --labels app=user
```

---

## 🗺️ Roadmap (Backlog)

- [ ] Strategy overrides via CLI (`--strategy`)
- [ ] Required variable validation for strategies
- [ ] Versioned strategies for upgrade support
- [ ] Multi-layout support (grouped, flat, regional)
- [ ] Pluggable pre/post hooks
- [ ] Docs + examples + starter templates

---

## 🤝 Contributing

We’re building Skiff to be a reliable, developer-first IaC companion. PRs, issues, and discussions welcome!

---

## 📄 License

MIT License
