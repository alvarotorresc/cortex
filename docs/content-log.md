# Content Log

Moments, decisions, and milestones worth capturing for blog posts, social media, or retrospectives.

---

## 2026-02-23 -- Project kickoff: Cortex
**Phase:** Ideation + Setup
**Type:** decision
**Potential:** blog | social media
**Context:** Starting Cortex, a self-hosted personal hub with a plugin system. The motivation: consolidating personal productivity tools (finance tracking, notes, bookmarks) into a single, local-first platform with zero cloud dependencies. Chose Go for the backend (new language to learn, great for systems programming), HashiCorp go-plugin + gRPC for plugin isolation (battle-tested in Terraform and Vault), SvelteKit for the frontend (lightweight, reactive, Svelte 5 runes), and SQLite per plugin for data isolation. The architectural reference is Grafana's plugin system. This is a flagship product designed to absorb the Finance App and serve as the foundation for future personal tooling.
