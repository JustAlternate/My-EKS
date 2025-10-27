# My AWS Infra

**/!\ This summary was writen by an AI based on notes I wrote at 4am that are here : [notes.md](./notes.md)**

> Apprentice Platform Engineer @ iAdvize  
> Passionate about GitOps, Infrastructure as Code, and SRE principles  

---

## 🌍 Project Overview — “Production-Ready AWS Infra”

This project is my personal journey to build a **production-grade infrastructure on AWS**,  
combining everything I’ve learned from:

- My work at **iAdvize** (Nomad-based internal deployment systems)
- My SRE/DevOPS books 
- Experimentation, research, and AI-assisted learning

### 🎯 Objectives

- Learn and apply **SRE / DevOps / Platform Engineering** best practices  
- Build a **GitOps-driven, observable, scalable infrastructure**
- Showcase end-to-end CI/CD, IaC, Observability, and SLO-driven operations  

### 💬 Guiding Philosophy

> “No overscoping. No overengineering.  
> Just building things like they *should* scale.”

This project exists to:
- Help me **learn deeply**
- Help me **embrace SRE**

---

## 🏗️ Architecture Overview

**Core Components:**
- **AWS:** VPC, EKS, RDS, ECR, CloudWatch, Route53
- **IaC:** OpenTofu (Terraform fork), S3 backend, GitHub Actions + OIDC auth
- **Services:** 2 Go microservices (API + Web)
- **CI/CD:** GitHub Actions pipelines for IaC & services
- **Observability:** CloudWatch (base), Prometheus/Grafana planned
- **Deployment philosophy:** Manual YAML now, moving toward ArgoCD GitOps

---

## ⚙️ Infrastructure Features & Roadmap

### ✅ IaC Setup
- S3 tfstate backend with DynamoDB locking
- EKS cluster (multi-AZ, arm64 nodes)
- Simple VPC + security groups
- Managed node group (tg4.small Amazon Linux)
- CloudWatch integration

### ✅ Release Engineering
- ECR (immutable, auto-cleanup for untagged images)
- CI for OpenTofu (lint, format, plan, apply)
- CI for Golang microservices (lint, test, build, tag, push)
- Automated release tagging (major/minor/patch via PR labels)
- Multi-arch build on arm64 runners

### ✅ Microservices & Networking
- `api` service → updates counter in RDS
- `web` service → serves frontend + calls `api`
- Private RDS with internal Route53 DNS
- Proper health checks (liveness, readiness, startup)
- LoadBalancer service for `web` public access

### 🔄 In Progress
- Managed Prometheus & Grafana (AMP/AMG)
- Service metrics instrumentation (Prometheus Go client)
- Grafana dashboards for:
  - EKS
  - RDS (CloudWatch)
  - Microservices latency and errors
- HPA & PDB configuration
- ArgoCD deployment + Argo Rollouts for safe deploys
- Define SLI/SLO and visualize in Grafana

### 🧩 Planned / Bonus
- RDS Backup & Restore (Velero)
- Helm charts for services
- Karpenter for automatic spot provisioning
- Chaos testing (k6, Litmus)
- Postmortems & Runbooks
- Simple Golang eBPF tracing (Pixie)
- Cilium for networking and circuit breaking
- Tempo + OpenTelemetry for tracing

---

## 🧱 Lessons & Reflections (REX)

### ☁️ First Steps in AWS
- Learning VPCs, IAM, and EKS provisioning was easier than expected.
- OpenTofu made infra creation seamless — love the open-source mindset.
- Multi-AZ setup helped me internalize **resilience design** early.

### 🔒 IaC CI/CD
- Using OIDC for GitHub → AWS was a big win: **no credentials stored**, fully secure auth flow.
- First time wiring “PR comment with tf plan” — felt like magic ✨
- Enforced branch labeling to control release versions (Major/Minor/Patch).

### 🐘 RDS Adventure
- Took hours to debug why my API couldn’t connect to RDS 😅
- Learned about AWS default security groups — EKS creates its own SG that must be whitelisted in RDS inbound rules.
- Set up a private Route53 record (`postgresql.justalternate-eks-cluster.internal`) for stable DB URLs.

### 🐳 Docker Struggles
- Cross-compiled arm64 images with **Chainguard base images** for secure, minimal builds.
- Struggled with Dockerfiles for Amazon Linux nodes — learned to test multi-arch locally before ECR.
- Realized thin images = faster deploys + cheaper ECR storage.

### ⚙️ Kubernetes Insights
- Manual YAML deploys were a great intro to K8s primitives.
- Using probes (liveness/readiness/startup) taught me how K8s treats unhealthy pods.
- LoadBalancer type for `web` made public access simple — good validation of routing.

---

## 📊 Observability Goals

- Integrate AMP + AMG for metrics collection and visualization
- Instrument custom app metrics:
  - HTTP latency
  - Request count
  - DB query time
- Create Grafana dashboards for:
  - Microservices
  - RDS
  - EKS nodes
- Implement SLI/SLO dashboards for golden signals:
  - Latency
  - Traffic
  - Errors
  - Saturation

---

## 🔍 SRE Practices Roadmap

| Goal | Description | Status |
|------|--------------|--------|
| **Define SLI/SLO** | Based on golden signals | 🧠 Learning |
| **Dashboards** | For services, infra, and RDS | 🧩 In progress |
| **Error Budgets** | Tie to alerting rules | 🔜 Planned |
| **Postmortems** | Template & incident simulations | 🔜 Planned |
| **Runbooks** | For recurring issues | 🔜 Planned |
| **Chaos Experiments** | k6 + pod kills | 🔜 Planned |

---

## 🧩 Tech Stack Summary

| Category | Stack |
|-----------|--------|
| **Cloud** | AWS (EKS, RDS, ECR, Route53, CloudWatch) |
| **IaC** | OpenTofu (Terraform fork) |
| **CI/CD** | GitHub Actions + OIDC |
| **Languages** | Golang (microservices), YAML (K8s), HCL (IaC) |
| **Observability** | CloudWatch, Prometheus, Grafana (planned) |
| **K8s Components** | Deployment, Service, HPA, PDB |
| **Bonus Tools** | Karpenter, ArgoCD, Helm, Chaos Mesh, Tempo |

---

## 🪄 Philosophy of the Project

I’m intentionally “overengineering” to:
- Learn to think like an **SRE** in a massive production environment
- Build confidence with real AWS/K8s tooling

---

## 💡 Next Steps

- [ ] Finish observability stack (AMP/AMG)
- [ ] Define first SLIs and Grafana alerts
- [ ] Implement ArgoCD + Rollouts
- [ ] Automate postmortem creation

---

## 📘 References & Inspiration

- **Becoming SRE** — David N. Blank-Edelman  
- **Google SRE Workbook**
- **Fundamentals of DevOps and Software Delivery** — Yevgeniy Brikman

---
