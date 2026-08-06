# ⏳ TimeCourt

A rule-based decision engine that evaluates authenticated requests against versioned rules and facts to produce deterministic decisions.

---

![Introduction](/assest/image.png)

---
## 🏛️ Core Pillars

The decision engine is built on immutable facts, versioned rules, and deterministic decisions. Together, they enable accurate historical reasoning, reproducible outcomes, and efficient evaluation without recomputation.

---
![Core Pillars](/assest/image2.png)

---
## 🎯 Requirements

TimeCourt is designed to manage immutable facts and versioned rules while providing deterministic, timestamp-aware decisions. The system emphasizes correctness, scalability, and historical reproducibility.

---
![Requirements](/assest/image3.png)

---
## 🗄️ Data Model & APIs

TimeCourt exposes a simple API surface built on immutable, versioned data models. Every decision is persisted and can be reproduced or explained at any point in time.

---
![Data Model & APIs](/assest/image4.png)

---
## 📈 Rule Versioning

TimeCourt models rules as immutable, versioned entities. Each rule version maintains a reference to its predecessor, enabling complete historical traceability and deterministic evaluation for any point in time.

---

![Rule Versioning](/assest/image5.png)

---
## 💾 Fact Storage

TimeCourt stores facts as immutable, versioned records to guarantee consistency and prevent data loss. Every fact is preserved exactly as received, ensuring that future decisions remain deterministic and historically reproducible.

---
![Fact Storage](/assest/image6.png)
