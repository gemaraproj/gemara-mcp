# Gemara

> GRC Engineering Model for Automated Risk Assessment. Gemara provides a logical model describing categories of compliance activities, how they interact, and schemas enabling automated interoperability between GRC tools. Part of the OpenSSF ecosystem.

Gemara delivers three components:

- **The Model** — a seven-layer conceptual framework for GRC activities
- **The Schemas** — CUE format schemas for validation and interoperability
- **The SDKs** — language-specific libraries for working with Gemara documents

The index of all pages is at [llms.txt](https://gemara.openssf.org/llms.txt).

---

## Scope

The purpose of the Gemara model is to provide a common basis for approaching activities and topics related to Governance, Risk, and Compliance (GRC), while leaving room for specific implementation details to grow and mature over time.

---

## Key Terms

- **Assessment** — (1) the process of determining whether an outcome meets the actor's intent; or (2) an atomic process within an Evaluation used to determine a resource's compliance with an Assessment Requirement
- **Assessment Requirement** — a tightly scoped, verifiable condition that must be satisfied and confirmed by an evaluator
- **Audit** — a formal, opinionated review of an organization's policies and posture, conducted at a specific point in time to verify that established requirements are met
- **Behavior Evaluation** — an opinionated observation of simulated or real-world activities
- **Capability** — a feature or function of a system; the primary component comprising an attack surface
- **Catalog** — a structured set of related prose and relevant metadata
- **Compliance** — adherence to a Rule or set of Rules
- **Control** — (1) an organization's ability to fully assert desired state on a system, resource, or state; or (2) a mechanism that asserts desired state; or (3) prose describing the Objective and Assessment Requirements associated with a desired state
- **Continuous Monitoring** — a multi-system process designed to collect Evaluation and operational data on an ongoing basis to detect non-compliance, enable Remediative Enforcement, and observe trends over time
- **Enforcement** — an action taken in response to non-compliance findings and their causes
- **Evaluation** — the manual or automated process of forming an opinion on the state of compliance, guided by a set of Assessment Requirements
- **Evaluation Finding** — the evidence and opinionated result of an Assessment
- **Governance** — strategic oversight of an organization and its activities
- **GRC** — (1) the Governance, Risk, and Compliance domain within cybersecurity; or (2) a coordinated program dedicated to these elements within a business unit
- **Guidance** — prose intended to help bring about a desired outcome for a topic or generalized scenario, based on knowledge of relevant Vectors
- **Guideline** — atomic element of a Guidance Catalog; often includes explanatory context and recommendations for designing optimal implementations
- **Intent Evaluation** — an Evaluation ensuring that a resource is prepared in alignment with Policy, such as through proper training, configuration, or code
- **Objective** — a unified statement of intent, which may encompass multiple situationally applicable statements or requirements
- **Opinion** — a firmly held approximation of reality formed within the constraints of an evaluator's philosophy, perspective, and capabilities
- **Organization** — any logical grouping of human, physical, virtual, and information resources such as a company, business unit, or team
- **Policy** — a clearly scoped set of rules based on an organization's Risk Appetite
- **Preventive Enforcement** — any action that interrupts another process which would otherwise cause non-compliance
- **Remediative Enforcement** — corrective action in response to non-compliance in a deployed activity
- **Residual Risk** — the Risk remaining after Risk Mitigation and Enforcement actions have been implemented
- **Risk** — the potential for loss or damage when a Threat is actualized, determined by calculating the impact of an event to an organization and the likelihood of its occurrence
- **Risk Acceptance** — a clearly documented decision to accept an unmitigated Risk as necessary or unavoidable
- **Risk Appetite** — the level of Risk an organization is willing to accept in pursuit of its objectives
- **Risk Assessment** — the process of identifying the potential or actual Risks introduced by a system
- **Risk Catalog** — a group of related Risks relevant to an organization; used to determine when and how Policies are created for the organization
- **Risk Mitigation** — the process of developing actions to prevent Threats or reduce their impact on organization objectives
- **Rule** — an active, enforceable Policy, regulation, or law
- **Sensitive Activity** — a type of action that introduces Risk to an organization
- **Threat** — a circumstance or event where the concepts of a Vector are applied to a Capability in a specific context, resulting in the potential for negative impact
- **Vector** — (1) an opportunity for an attacker to exploit a vulnerability; or (2) a path by which neglect could result in unintentional negative outcomes
- **Vulnerability** — (1) a weakness in a system inherent in or associated with a Capability that can be exploited when used in unintended ways; or (2) a lack of Control or gap in defense that can be leveraged to cause harm

---

## Foundational Concepts

The Gemara model's structure was directly inspired by ISO 7498 (the OSI Model), applying the same layered separation of concerns to GRC. Different teams and tools focus on specific activities — from high-level Guidance to low-level Enforcement — while interoperating within a cohesive system.

Risk Assessment takes many forms: some activities look forward, anticipating where Risk may occur; others look backward, inspecting active or deployed activities to determine whether Risk has materialized. The Gemara model categorizes these into seven cumulative layers, each calling upon the outputs of those below it.

Layer 4 is the pivot point: it captures Sensitive Activities — the actions that require governance — bridging the Definition layers (1–3) and the Measurement layers (5–7).

---

## The Seven-Layer Model

### Overview

The model is organized into two primary categories:

- **Definition layers (1–3):** produce document assets that may be referenced by higher layers. Each layer builds upon the previous.
- **Measurement layers (5–7):** produce timestamped logs as outputs. These look back at Sensitive Activities to determine compliance.
- **Layer 4:** the pivot point, representing the Sensitive Activities themselves.

A fundamental principle: each layer builds upon the one below it. An Evaluation at Layer 5 tests for conformance with a Policy defined at Layer 3, which is informed by Controls at Layer 2.

### Layer 1 — Vectors & Guidance

Layer 1 covers generic, high-level Risk Assessment. The need for this type of assessment is typically surfaced by factors far removed from the scope of the specific activity being assessed — legislation, emerging technology categories, or industry-wide concerns.

**Vectors** document opportunities for attack or neglect. They do not require understanding of every technological detail; the focus is on the opportunity for mistake or malice. Vectors may be published standalone or within a catalog. Examples can be found in MITRE ATT&CK as *techniques*.

**Guidance** provides prose intended to help bring about desired outcomes. Its constituent parts (Guidelines) do not typically stand alone and are most often published as longstanding Guidance Catalogs. Each Guideline often includes explanatory context and recommendations for designing optimal outcomes without foreknowledge of implementation details.

Guidance is often developed by industry groups, government agencies, or international standards bodies. Examples include the OWASP Top 10, NIST Cybersecurity Framework, HIPAA, GDPR, CRA, PCI standards, and ISO standards.

Vector artifacts can be referenced by both Guidance and Threats to accelerate authoring and increase fidelity. Guidance artifacts can be referenced by Controls to demonstrate how a particular Control applies the respective Guideline.

### Layer 2 — Threats & Controls

Layer 2 builds upon Layer 1 to produce technology-specific, threat-informed Controls. These remain in the hypothetical realm — not yet tied to a specific organization's Risk Catalog, but providing Controls with Assessment Requirements that can be fully operationalized later.

**Threats** draw from known Vectors and are mapped to a specific Capability, location, or action, making clear precisely when or where the Threat can be expected to manifest. Threats may be associated with a specific scenario, business unit, or technology.

**Controls** build upon established Guidance and contain Assessment Requirements that can be directly implemented by evaluators. Controls are typically developed by an organization for internal use, or for general purpose by industry groups, government agencies, or standards bodies. Examples include CIS Benchmarks, FINOS Common Cloud Controls (CCC), and the OpenSSF OSPS Baseline.

Controls are traditionally delivered as Control Catalogs. An emerging trend toward composability defines reusable, agnostic Controls for information systems — nearly like Guidance in their generic prose, but with Assessment Requirements detailed enough to serve as Layer 2 assets.

The recommended process for creating a Control Catalog: assess the technology's capabilities, identify Threats to those Capabilities, then develop Controls to mitigate those Threats.

### Layer 3 — Risk & Policy

Layer 3 is where Risk Assessments exit the conceptual phase and meet organizational reality. Controls are appropriately selected based on the most relevant details of a specific organization.

**Risk** requires a firm understanding of the organization's landscape — technical, operational, and geopolitical details. It is calculated from a Threat Assessment (situational likelihood of a negative outcome) combined with the potential impact on the organization in the context of a particular Sensitive Activity.

**Risk Appetite** is the level of Risk an organization is willing to accept in pursuit of its objectives. A Risk Catalog documents this appetite and is used to determine when and how Policies are created.

**Policy** is a clearly scoped set of rules based on an organization's Risk Appetite. It provides governance rules tailored to the organization — based on best practices and industry standards, but specific to the organization's context and risk profile. Policies inevitably introduce some level of Risk Acceptance and cannot be properly developed without consideration of organization-specific Risk Appetite.

A complete Policy document will be time-bound, reference Threat-informed Controls with Assessment Requirements, and include a plan for rolling out the Policy to impacted parties. Policy documents may reference other Policy documents, creating a functional inheritance model. If created during planning, a Policy can serve as a functional design requirement to ensure security is built into the Sensitive Activity.

### Layer 4 — Sensitive Activities (Pivot Point)

Sensitive Activities are any type of action that might introduce Risk to an organization, creating the need for governance. They are the heart of the Gemara Model and the reason GRC has existed throughout history.

Sensitive Activities take many forms: the building code regulating how a bank branch is built, the rules a bank teller must follow to stay within the law, the processes a development team must follow when creating a new mobile app, or the requirements placed on software code itself.

The Definition layers (1–3) point toward Sensitive Activities, defining what is acceptable and what is not. The Measurement layers (5–7) look back at Sensitive Activities or their outcomes to determine compliance.

Example application — software supply chain security: Policies inform both development workflows and evaluation tools such as Software Composition Analysis (SCA). SCA findings are then integrated into enforcement mechanisms and compliance audits. The model helps leaders verify that all essential activities across the full lifecycle are being performed.

### Layer 5 — Evaluation

Layer 5 produces Opinions through structured inspection of Policy compliance. Evaluations consist of either Intent Evaluation or Behavior Evaluation.

**Intent Evaluation** ensures that a resource is prepared in alignment with Policy — through proper training, configuration, or code. Examples include automated or manual examination of configurations, software composition analysis, cloud resource configuration scanning, dashboard reviews, and code analysis.

**Behavior Evaluation** observes or simulates user behavior to ensure expected outcomes are achieved. Simulating bad behavior identifies security gaps or Policy non-compliance; simulating good behavior verifies the system operates as expected. Digital examples include penetration testing; analog examples include secret shoppers.

An **Opinion** is a firmly held approximation of reality formed within the constraints of the evaluator's philosophy, perspective, and capabilities. This is a post hoc Risk Assessment establishing an understanding of what Risks have been introduced or proposed.

The evaluation process involves comparing Evaluation Findings to the organization's expectations as captured in Policies through Controls and Assessment Requirements. Proper maintenance of mappings between each artifact allows Evaluations to demonstrate a system's compliance state across multiple relevant artifacts simultaneously.

### Layer 6 — Enforcement

When an Evaluation produces a non-compliance finding, the response is an Enforcement action. Choosing when and how to enforce requires forming an Opinion on the best course of action — this remains a Risk Assessment activity.

**Preventive Enforcement** interrupts a process to prevent non-compliance. This is primarily in response to malformed intent (such as a bad configuration) but may also include complex behavioral evaluation processes executed before deployment. In software development, this may take the form of a deployment gate — controlling admission similarly to a physical perimeter gate.

**Remediative Enforcement** describes corrective action in response to non-compliance in a deployed activity. This may include isolating, replacing, retraining, or redeploying the noncompliant resource. It requires observability to detect target state and communication with the latest applicable Policies to ensure changes to the environment or Policy result in Policy-aligned remediation. Proper logging and alerting should always accompany remediation tools.

Enforcement activities find heightened success when properly mapped to Evaluation results, allowing every person or system involved to understand the complete justification.

### Layer 7 — Audit & Continuous Monitoring

**Audit** is a formal, opinionated review of an organization's Policies and posture, assessing the quality of past Risk Assessments as well as Residual Risk. An Audit is an approximation of compliance posture at a point in time; its outcome is a firm assertion based on observations of available evidence or gathered facts.

Audits may have a variety of scopes and may involve forming opinions on the Guidance an organization follows, the Controls it writes, the Policies it implements, its Evaluation methods, and its Enforcement status. The opinion is typically informed by a combination of geopolitical regulatory requirements and known best practices for the industry and resources involved. Traditionally, Audits have favored manual Intent Evaluations, but cybersecurity Audits are increasingly integrating automated tools for both Intent and Behavioral Evaluation.

**Continuous Monitoring (CCM)** aggregates results from Evaluations and operational metrics in real time. This process identifies Threats and Control failures as they occur to support Remediative Enforcement. Mature CCM operations establish a persistent, Policy-driven process that harnesses multiple systems to ensure maximum visibility of all deployed assets at all times.

---

## Schemas

Gemara schemas are written in CUE format and standardize the expression of elements in the model. They enable automated validation and interoperability between GRC tools.

Schemas have three lifecycle states: **Experimental** → **Stable** → **Deprecated**.

Definition layer collections (1–3) are called **Catalogs**. Measurement layer collections (5–7) are called **Logs**.

| Layer | Artifact | Description |
|-------|----------|-------------|
| 1 | Guidance Catalog | High-level guidance from industry groups and standards bodies |
| 1 | Vector Catalog | Attack vectors and techniques used to compromise information systems |
| 1 | Principle Catalog | Foundational values guiding governance, design, and operational decisions |
| 2 | Control Catalog | Technology-specific, threat-informed security controls |
| 2 | Capability Catalog | Technology capabilities that can be leveraged to implement controls |
| 2 | Threat Catalog | Threats mapped to technology capabilities and security controls |
| 3 | Risk Catalog | Organizational risk categories, severity levels, and risk appetite definitions |
| 3 | Policy | Risk-informed rules tailored to an organization's specific needs and risk appetite |
| 5 | Evaluation Log | Inspection of code, configurations, and deployments against policies and controls |
| 6 | Enforcement Log | Prevention or remediation based on assessment findings |
| 7 | Audit Log | Review of organizational policy and conformance |
| — | Mapping Document | Cross-layer artifact mappings between frameworks |
| — | Lexicon | Controlled vocabulary for Gemara terms |
| — | Base Types | Shared type definitions used across schemas |
| — | Mapping Primitives | Inline mapping structures for cross-referencing |

Validate a document using CUE:

```bash
go install cuelang.org/go/cmd/cue@latest
cue vet -c -d '#ControlCatalog' github.com/gemaraproj/gemara@latest your-controls.yaml
```

Full schema reference: https://gemara.openssf.org/schema/
CUE Central Registry: https://registry.cue.works/docs/github.com/gemaraproj/gemara@latest

---

## SDKs

Language-specific SDKs provide programmatic access to Gemara documents. SDK types are generated from CUE schemas, ensuring consistency between validation and programmatic access.

Each SDK maintains its own independent SemVer lifecycle. SDKs must explicitly document which core specification release version they support. When a new core specification release is published, SDKs regenerate their types from the updated schemas.

**Available SDKs:**

- **Go SDK** — type-safe Go APIs for reading, writing, and manipulating Gemara documents: https://gemara.openssf.org/sdk/go-sdk

---

## Real-World Usage

Gemara is in use in production environments:

- **FINOS Common Cloud Controls** — Layer 2 controls for cloud environments
- **OpenSSF Open Source Project Security Baseline (OSPS Baseline)** — Layer 2 security baseline for open source projects
- **Privateer** — Layer 5 evaluation framework with plugins (e.g., the OSPS Baseline Plugin)

---

## Optional

- **Community:** OpenSSF Slack #gemara channel, biweekly meetings — https://gemara.openssf.org/community/
- **Architecture Decision Records:** https://gemara.openssf.org/adrs/
