# Skiff 🚢

**Skiff** is a lightweight, extensible Command-Line Interface (CLI) designed for managing Terraform and Terragrunt infrastructure deployments at scale across multiple cloud environments. It utilizes a declarative YAML manifest for configuration. Skiff aims to abstract the complexities of directory structure organization, enabling platform and DevOps teams to focus on Terraform module development rather than structural management overhead.

---

## 🚀 Features

- ⚙️ **Service Types Catalog** – Define reusable services (terraform modules ) like VPC, RDS, IAM
- 🧾 **Declarative Manifest** – Describe services and dependencies
- 🛠️ **Template-based Terragrunt File Generation**
- 🧭 **Strategy-Driven Folder Layouts** – Flexible, configurable directory structures
- 🏃 **Command Runner** – Run `terragrunt` commands filtered by labels
- 🔁 **Dry-Run Support** – Preview changes without executing them

---

## 🚀 Quickstart

### Install

```console
go install github.com/nyambati/skiff@latest
```

### Initialize a New Project

```console
skiff init
```

This creates the project folder structure and a default .skiff config file.

```console
.
├── .skiff              # Project config (includes strategy)
├── manifests/          # Service types + service definitions manifest
│   ├── catalog.yaml
│   └── 123456789012.yaml
├── templates/          # Terragrunt templates
│   └── terragrunt.default.tmpl
└── terragrunt/         # Output folder (rendered files)
```

### Add/Edit service catalog

```console
skiff edit  catalog --type vpc \
  --values source="github.com/terraform-aws-modules/terraform-aws-vpc",version="4.16.0"
```

### Add/Edit manifest

```console
skiff edit manifest \
  --name my-manifest \
  --metadata account_name=my_account,account_id=123456789012,env=production
```

### Add/Edit service

```console
skiff edit service \
  --manifest my-manifest \
  --service simple-vpc
```

### Generate Terragrunt files

```console
skiff generate --manifest my-manifest --labels env=prod,region=us-west-2
```

### Run terragrunt

```console
skiff [plan,apply,destroy] --labels env=prod,region=us-west-2
```

📚 Full Documentation

See docs/ for a complete guide including design philosophy, layout strategies, and customization options.

## 🤝 Contributing

We’re building Skiff to be a reliable, developer-first IaC companion. PRs, issues, and discussions welcome!

---

## 📄 License

MIT License
