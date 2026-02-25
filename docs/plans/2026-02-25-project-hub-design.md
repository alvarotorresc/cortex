# Project Hub -- Cortex Plugin Design

> Design document for the Project Hub plugin.
> Author: product-agent
> Date: 2026-02-25
> Status: Proposal

---

## 1. Problem Statement

Alvaro manages ~15 projects across two categories (Flagship and Lab), each with its own repo, stack, domain, status, and version. Today this information lives scattered across:

- `ecosystem-plan.md` (markdown, manual, hard to scan)
- GitHub repos (fragmented, no unified view)
- Mental model (undocumented connections, priorities, absorptions)

**The pain:** There is no single place to look at the full picture and answer basic questions instantly: "What state is Guitar App in?", "What version is Swiss Knife on?", "Which projects use Supabase?", "What did I absorb into what?"

**Who it is for:** Alvaro TC, sole user. This is a personal tool inside Cortex.

**Value proposition in one sentence:** See the state of your entire project ecosystem in one glance, from one screen, without opening a single repo.

---

## 2. Competitive Landscape

| Alternative | Strengths | Weaknesses | Differentiator for us |
|---|---|---|---|
| GitHub profile + pinned repos | Already exists, zero effort | Max 6 pins, no custom metadata, no status tracking, no private view | Full CRUD, custom fields, all projects visible |
| Notion / spreadsheet | Flexible, collaborative | External tool, no integration with Cortex, manual sync | Native to Cortex, same UI, same local data philosophy |
| `ecosystem-plan.md` | Already works, text-searchable | No visual overview, hard to scan 15+ projects, no filtering/sorting, no quick edits | Visual cards, filters, color-coded status, widget |

**Verdict:** None of these give a fast, visual, local-first dashboard of the full ecosystem. Building it as a Cortex plugin makes it the first plugin that serves as "meta" infrastructure -- it tracks the ecosystem that Cortex itself belongs to.

---

## 3. Data Model

### Table: `projects`

| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | INTEGER | PRIMARY KEY AUTOINCREMENT | Internal ID |
| `name` | TEXT | NOT NULL, UNIQUE | Display name ("Swiss Knife", "Cortex") |
| `slug` | TEXT | NOT NULL, UNIQUE | URL-safe identifier ("swiss-knife", "cortex") |
| `tagline` | TEXT | NOT NULL | One-sentence description |
| `status` | TEXT | NOT NULL, CHECK enum | Current status (see enum below) |
| `category` | TEXT | NOT NULL, CHECK enum | "flagship" or "lab" |
| `version` | TEXT | | Current version ("v0.3", "v1.0.0") or NULL if pre-release |
| `stack` | TEXT | NOT NULL | Comma-separated tech stack ("Kotlin, Jetpack Compose, Material 3") |
| `icon` | TEXT | NOT NULL DEFAULT 'folder' | Lucide icon name |
| `color` | TEXT | NOT NULL DEFAULT '#0070F3' | Hex color for theming |
| `repo_url` | TEXT | | GitHub repo URL |
| `web_url` | TEXT | | Production/deploy URL |
| `docs_url` | TEXT | | Documentation URL |
| `hosting` | TEXT | | Where it runs ("Vercel", "Netlify", "Railway", "Self-hosted", "APK") |
| `notes` | TEXT | | Free-text notes (markdown supported) |
| `sort_order` | INTEGER | NOT NULL DEFAULT 0 | Manual sort within category |
| `created_at` | TEXT | NOT NULL DEFAULT datetime('now') | When the record was created |
| `updated_at` | TEXT | NOT NULL DEFAULT datetime('now') | Last modification timestamp |

### Status enum

| Value | Meaning | Color suggestion |
|---|---|---|
| `concept` | Idea stage, no code yet | `text-tertiary` (gray) |
| `design` | Design approved, pre-implementation | `#6366F1` (indigo) |
| `development` | Actively being built | `#0070F3` (brand-blue) |
| `active` | Deployed and maintained | `#16A34A` (success green) |
| `maintenance` | Stable, only bug fixes | `#D97706` (warning amber) |
| `archived` | No longer maintained | `text-tertiary` (gray, dimmed) |
| `absorbed` | Absorbed into another project | `text-tertiary` (gray, strikethrough) |

### Table: `project_links`

For projects that need more than the 3 standard URLs (repo, web, docs).

| Column | Type | Constraints | Description |
|---|---|---|---|
| `id` | INTEGER | PRIMARY KEY AUTOINCREMENT | Internal ID |
| `project_id` | INTEGER | NOT NULL, FK projects(id) ON DELETE CASCADE | Parent project |
| `label` | TEXT | NOT NULL | Display label ("Play Store", "npm", "Figma", "API docs") |
| `url` | TEXT | NOT NULL | The URL |
| `sort_order` | INTEGER | NOT NULL DEFAULT 0 | Display order |

### Indexes

```sql
CREATE INDEX idx_projects_status ON projects(status);
CREATE INDEX idx_projects_category ON projects(category);
CREATE INDEX idx_project_links_project_id ON project_links(project_id);
```

### Why this model

- **Flat `stack` as comma-separated text** instead of a join table: There are only ~15 projects. A normalized `project_technologies` join table adds complexity for zero practical benefit at this scale. Filtering by tech can use `LIKE '%SolidJS%'` which is fine for <50 rows. If the ecosystem grows to 100+ projects (it will not), reconsider.
- **`project_links` as separate table** instead of JSON blob: Links are variable-length per project (some have 1, some might have 5). A separate table allows clean CRUD without parsing JSON. It also makes the schema more honest about the relationship.
- **`sort_order` on both tables:** Manual ordering is important. Alvaro thinks about projects in a specific mental order (Flagship #1 is Quedamos, not alphabetical). Respect that.
- **`notes` as free text:** Not structured. These are quick thoughts like "waiting for Play Store verification" or "need to migrate to Lefthook". Markdown rendering is a nice-to-have but not required for MVP.

---

## 4. Views and Screens

### 4.1 Dashboard Widget (Cortex home page)

**Slot:** `dashboard-widget`

The widget shows a compact summary of the ecosystem.

```
+------------------------------------------+
| [FolderGit2]  Project Hub                |
|                                          |
|  15 projects                             |
|                                          |
|  [green dot] 3 Active                    |
|  [blue dot]  2 In Development            |
|  [indigo dot]3 Design Approved           |
|  [gray dot]  4 Concept                   |
|  [amber dot] 1 Maintenance               |
|  [gray dim]  2 Absorbed                  |
|                                          |
+------------------------------------------+
```

**Data returned by `GetWidgetData`:**

```json
{
  "data": {
    "total": 15,
    "by_status": {
      "active": 3,
      "development": 2,
      "design": 3,
      "concept": 4,
      "maintenance": 1,
      "archived": 0,
      "absorbed": 2
    }
  }
}
```

**Interaction:** Clicking the widget navigates to the plugin's full page.

### 4.2 Full Page -- Project List (default view)

**Route:** `/plugins/project-hub`

This is the main screen. A filterable, sortable grid/list of all projects.

```
+--------------------------------------------------------------+
| [FolderGit2]  Project Hub                    [+ New Project]  |
|                                                               |
| Filters: [All Status v] [All Categories v] [Search...      ] |
|                                                               |
| -- FLAGSHIP (7) -------------------------------------------- |
|                                                               |
| [violet dot] Quedamos          v0.2     In Development        |
| React+Ionic+NestJS             quedamos.app                   |
|                                                               |
| [red dot] Sinherencia          active   Active                |
| Astro 5                        sinherencia.com                |
|                                                               |
| [amber dot] Guitar App         --       Concept               |
| SolidJS, Web Audio API         --                             |
|                                                               |
| [emerald dot] Cortex           MVP      In Development        |
| Go, SvelteKit, gRPC            cortex.alvarotc.com            |
|                                                               |
| ... more ...                                                  |
|                                                               |
| -- LAB (7) ------------------------------------------------- |
|                                                               |
| [dice icon] Swiss Knife        v0.3     Active                |
| Kotlin, Jetpack Compose        Android APK                    |
|                                                               |
| ... more ...                                                  |
+--------------------------------------------------------------+
```

**Layout:**
- Header with title and "New Project" button
- Filter bar: status dropdown, category dropdown, text search
- Projects grouped by category (Flagship first, then Lab)
- Each project is a card/row showing: icon+color, name, version, status badge, stack summary, primary URL
- Clicking a project opens the detail view

**Sorting:**
- Within each category, projects sort by `sort_order` (manual)
- The filter bar allows overriding with: alphabetical, by status, by last updated

### 4.3 Full Page -- Project Detail

**Route:** `/plugins/project-hub?project={slug}` (or side panel, see approaches)

Displays all information about one project with edit capabilities.

```
+--------------------------------------------------------------+
| [<- Back]                                    [Edit] [Delete]  |
|                                                               |
| [violet dot large]                                            |
| Quedamos                                           v0.2       |
| Coordinar quedadas entre amigos                               |
|                                                               |
| Status: [In Development]    Category: [Flagship]              |
|                                                               |
| Stack                                                         |
| React, Ionic, Capacitor, NestJS, Supabase                    |
|                                                               |
| Hosting: Vercel (web) + Railway (API) + Supabase             |
|                                                               |
| Links                                                         |
| [GitHub] github.com/alvarotorresc/quedamos-app               |
| [Globe]  quedamos.app                                         |
| [Book]   docs/plans/quedamos-design.md                        |
|                                                               |
| Notes                                                         |
| v0.2 in development. Chat and map features planned.           |
| Monorepo structure. Absorbs nothing.                          |
|                                                               |
| ---                                                           |
| Created: 2026-02-13  |  Last updated: 2026-02-25             |
+--------------------------------------------------------------+
```

### 4.4 Create/Edit Project Form

**Trigger:** "New Project" button or "Edit" on detail view.

**Implementation:** Modal dialog (consistent with Cortex UI patterns -- Quick Notes uses modals for create/edit).

```
+--------------------------------------------------------------+
|                    New Project                          [X]    |
|                                                               |
| Name*          [________________________]                     |
| Tagline*       [________________________]                     |
| Status*        [Concept           v]                          |
| Category*      [Lab               v]                          |
| Version        [________________________]                     |
| Stack*         [________________________]                     |
| Icon           [folder            v]  (dropdown with preview) |
| Color          [#0070F3] [color swatch]                       |
| Hosting        [________________________]                     |
|                                                               |
| -- Links --                                                   |
| Repository URL [________________________]                     |
| Website URL    [________________________]                     |
| Docs URL       [________________________]                     |
| [+ Add custom link]                                           |
|                                                               |
| Notes                                                         |
| [                                                   ]         |
| [                                                   ]         |
| [                                                   ]         |
|                                                               |
|                              [Cancel]  [Save Project]         |
+--------------------------------------------------------------+
```

**Field behavior:**
- `Name` and `Tagline`: free text, required
- `Status`: dropdown with the 7 status options
- `Category`: dropdown with "flagship" / "lab"
- `Version`: free text, optional (empty for pre-release projects)
- `Stack`: free text, comma-separated (considered tag input but YAGNI for MVP)
- `Icon`: dropdown of common Lucide icons (curated list of ~20-30 relevant ones)
- `Color`: hex color input with a swatch preview
- `Hosting`: free text
- Links: 3 standard fields (repo, web, docs) plus dynamic "add custom link" with label+url pairs
- `Notes`: textarea, optional

---

## 5. Interactions

### Adding a project

1. User clicks "New Project" on the list page
2. Modal opens with empty form, status defaults to "concept", category defaults to "lab"
3. User fills fields, clicks "Save Project"
4. Toast confirmation: "Project created"
5. List refreshes, new project appears at the bottom of its category

### Editing a project

1. User clicks on a project to open detail view
2. User clicks "Edit" button
3. Same modal opens, pre-filled with current data
4. User modifies fields, clicks "Save Project"
5. Toast confirmation: "Project updated"
6. Detail view refreshes with new data

### Archiving / changing status

This is just an edit. Change the status dropdown to "archived" or "absorbed". No special workflow needed for MVP.

### Deleting a project

1. User clicks "Delete" on detail view
2. Confirmation dialog: "Delete {name}? This cannot be undone."
3. User confirms
4. Project and its links are deleted (CASCADE)
5. Toast: "Project deleted"
6. Redirect to list view

### Reordering projects

Not in MVP. Projects sort by `sort_order` which defaults to creation order. Manual drag-and-drop reordering is a post-MVP feature.

---

## 6. User Stories

### US-1: View all projects at a glance

```
As Alvaro
I want to see all my projects in a single list grouped by category
So that I know the current state of my ecosystem without checking multiple sources

Acceptance criteria:
- [ ] List page shows all projects grouped by Flagship and Lab
- [ ] Each project shows: name, version, status badge, stack, primary URL
- [ ] Status badges are color-coded
- [ ] Each project's icon and theme color are visible
```

### US-2: Add a new project

```
As Alvaro
I want to add a new project with its metadata
So that I keep my ecosystem catalog up to date

Acceptance criteria:
- [ ] "New Project" button opens a modal form
- [ ] Required fields: name, tagline, status, category, stack
- [ ] Optional fields: version, icon, color, hosting, links, notes
- [ ] Slug is auto-generated from name (kebab-case)
- [ ] Duplicate name/slug shows validation error
- [ ] After save, project appears in the list
```

### US-3: Edit project details

```
As Alvaro
I want to edit any field of an existing project
So that I can update versions, status, and notes as projects evolve

Acceptance criteria:
- [ ] Edit button on detail view opens pre-filled modal
- [ ] All fields are editable
- [ ] updated_at is set automatically on save
- [ ] Changes reflect immediately in list and detail views
```

### US-4: Filter and search projects

```
As Alvaro
I want to filter projects by status and category, and search by name
So that I can quickly find what I am looking for in a growing list

Acceptance criteria:
- [ ] Status filter dropdown (All, Concept, Design, Development, Active, Maintenance, Archived, Absorbed)
- [ ] Category filter dropdown (All, Flagship, Lab)
- [ ] Text search filters by project name (case-insensitive, instant)
- [ ] Filters combine (AND logic)
- [ ] Showing "X of Y projects" count
```

### US-5: Dashboard widget

```
As Alvaro
I want to see a summary of my projects on the Cortex dashboard
So that I get a quick pulse of the ecosystem without navigating to the plugin

Acceptance criteria:
- [ ] Widget shows total project count
- [ ] Widget shows count per status with color-coded dots
- [ ] Clicking widget navigates to the Project Hub full page
```

### US-6: Delete a project

```
As Alvaro
I want to delete a project that no longer exists
So that the catalog stays clean

Acceptance criteria:
- [ ] Delete button on detail view
- [ ] Confirmation dialog before deletion
- [ ] Deletion removes project and all its custom links
- [ ] Redirects to list after deletion
```

---

## 7. Approaches

### Approach A: List + Modal (Recommended)

**Description:** The main view is a vertical list of project cards grouped by category. Clicking a card opens a detail view (same page, scroll-to or replace content). Create/edit uses a modal overlay.

**Pros:**
- Consistent with Quick Notes plugin UX (list + modal for CRUD)
- Simple to implement. One page component with conditional rendering
- Works well on both large screens and smaller windows
- Fast navigation. No page transitions needed

**Cons:**
- Detail view replaces the list, requiring a "back" button
- Modal forms can feel constrained for many fields

**Effort:** Low-medium. ~1 week for backend, ~1.5 weeks for frontend. Total: ~2.5 weeks.

### Approach B: List + Side Panel

**Description:** The main view is a list on the left (~60% width). Clicking a project opens a detail panel on the right (~40% width) without leaving the list. Create/edit uses the same side panel.

**Pros:**
- Context preserved. Can see the list while viewing details
- Feels more like a "dashboard" / Linear-style experience
- No back button needed

**Cons:**
- More complex layout logic (responsive breakpoints, panel sizing)
- On smaller screens the panel must either overlay or replace the list anyway
- More frontend work for the same functionality

**Effort:** Medium. ~1 week for backend (same), ~2.5 weeks for frontend. Total: ~3.5 weeks.

### Approach C: Card Grid + Drawer

**Description:** Projects displayed as a visual grid of cards (like a portfolio/gallery). Clicking opens a full-height drawer from the right with details. Create/edit in the drawer.

**Pros:**
- Most visually striking. Each card shows the project's color and icon prominently
- Gallery feel matches "portfolio overview" mental model
- Drawer pattern is common and well-understood

**Cons:**
- Cards waste more vertical space than list rows
- Grid layout requires more responsive tuning
- Harder to scan quickly compared to a dense list (information density is lower)
- Drawer needs its own scroll context

**Effort:** Medium-high. ~1 week for backend (same), ~3 weeks for frontend. Total: ~4 weeks.

### Recommendation: Approach A (List + Modal)

**Rationale:**
1. **Time-box compliance.** The MVP framework says 2-4 weeks. Approach A fits in ~2.5 weeks. Approach B is borderline at 3.5. Approach C exceeds at 4.
2. **Consistency.** Quick Notes and Finance Tracker both use list + modal. Users (Alvaro) build muscle memory with consistent patterns.
3. **Information density.** A list is faster to scan than a grid when you have 15 items and care about status/version/stack. This is a utility tool, not a portfolio showcase.
4. **Simplicity principle.** Pilar 1. The simplest solution that works is the correct one. A side panel or drawer adds complexity that does not solve any problem the modal approach fails to solve.
5. **Upgrade path.** Starting with A does not prevent migrating to B later if the list grows or if Alvaro wants the side-panel feel. The data model and API are identical across all three approaches.

---

## 8. MVP Definition (v0.1)

### What is IN

- SQLite schema: `projects` table + `project_links` table + migrations
- Backend Go plugin: full CRUD for projects and links + dashboard widget data
- Frontend: list view grouped by category, detail view, create/edit modal, delete with confirmation
- Filters: by status, by category, text search by name
- Dashboard widget: total count + count by status
- i18n: EN + ES
- Seed data: the 15 current ecosystem projects pre-loaded in the migration
- Manifest: `id: "project-hub"`, `icon: "folder-git-2"`, `color: "#8B5CF6"` (violet -- distinct from other plugins)

### What is NOT in v0.1

- **Drag-and-drop reordering.** Sort order is set via a number field in the edit form. Good enough.
- **GitHub API integration.** No scraping, no API calls. All data is manually entered. This was an explicit decision by the user.
- **Markdown rendering in notes.** Notes are plain text in v0.1. Rendering markdown adds a dependency.
- **Tags / technology taxonomy.** Stack remains a free-text comma field. No tag system, no autocomplete.
- **Project relationships.** "Absorbed by" connections (e.g., Finance App absorbed by Cortex, Price Tracker absorbed by Fogon) are tracked via notes text, not a formal relation. A `parent_id` or `absorbed_by_id` column is overkill for 2 cases.
- **Activity timeline / changelog per project.** No history of status changes. If needed later, add an `events` table.
- **Bulk actions.** No multi-select, no bulk status change.
- **Export/import.** No CSV or JSON export.

### Success Metrics

Since this is a personal tool with a single user, traditional metrics (DAU, retention) do not apply. Instead:

| Metric | Target | How to measure |
|---|---|---|
| Data completeness | All 15 ecosystem projects entered within 1 day of first use | Manual check |
| Replaces ecosystem-plan reference | Alvaro checks Project Hub instead of ecosystem-plan.md for project status | Self-reported |
| Time to answer "what state is X in?" | Under 5 seconds (open Cortex, glance at widget or list) | Qualitative |
| Data freshness | Projects updated within 24h of status changes | Manual check |

---

## 9. API Design

All routes are prefixed by the Cortex plugin router: `/api/plugins/project-hub/...`

### Projects

| Method | Path | Description |
|---|---|---|
| `GET` | `/projects` | List all projects. Query params: `status`, `category`, `search` |
| `GET` | `/projects/:slug` | Get one project with its links |
| `POST` | `/projects` | Create a project |
| `PUT` | `/projects/:slug` | Update a project |
| `DELETE` | `/projects/:slug` | Delete a project and its links |

### Links

| Method | Path | Description |
|---|---|---|
| `POST` | `/projects/:slug/links` | Add a custom link to a project |
| `PUT` | `/links/:id` | Update a custom link |
| `DELETE` | `/links/:id` | Delete a custom link |

### Widget

Handled by `GetWidgetData("dashboard-widget")` -- returns the counts JSON described in section 4.1.

### Response format

Per PATTERNS.md:

```json
// Success
{ "data": { ... } }

// Error
{ "error": { "code": "VALIDATION_ERROR", "message": "name is required" } }
```

---

## 10. Seed Data

The migration includes an INSERT with all 15 current ecosystem projects so the plugin is useful from the first load. This avoids a cold empty state.

**Flagship:**
1. Quedamos -- v0.2, development, React+Ionic+NestJS+Supabase, #8B5CF6
2. Sinherencia -- active, Astro 5, #DC2626
3. Guitar App -- concept, SolidJS+Web Audio API, #F59E0B
4. Cortex -- development, Go+SvelteKit+gRPC+SQLite, #10B981
5. Fogon -- design, React Native (Expo)+NestJS+Supabase, #C2410C
6. Huellas -- design, Next.js 15+RN/Expo+NestJS+Supabase/PostGIS, #F97316
7. Libroteca -- concept, TBD, #0070F3

**Lab:**
1. create-astro-blog -- active, Node.js CLI, #0070F3
2. PokeUtils -- active, HTML/CSS/JS, #EF4444
3. DevTools -- active, Astro 5+TypeScript, #06B6D4
4. Clipboard Manager -- concept, Go+WebSocket, #0070F3
5. System Config Manager -- concept, Rust CLI, #0070F3
6. Swiss Knife -- active (v0.3), Kotlin+Jetpack Compose, #0070F3
7. IronLog -- design, React Native (Expo)+expo-sqlite, #64748B

**Absorbed (included for completeness):**
- Finance App -- absorbed (by Cortex), Tauri 2+Svelte 5+SQLite, #16A34A
- Price Tracker -- absorbed (by Fogon), Python+FastAPI+Turso, #EA580C

---

## 11. Plugin Manifest

```json
{
  "id": "project-hub",
  "name": "Project Hub",
  "version": "0.1.0",
  "author": "alvarotorresc",
  "description": "Track the state of your entire project ecosystem",
  "icon": "folder-git-2",
  "color": "#8B5CF6",
  "permissions": ["db:read", "db:write"],
  "slots": {
    "dashboard-widget": true,
    "full-page": true
  }
}
```

**Why `folder-git-2`:** It is a Lucide icon that represents "projects in a folder with version control" -- exactly what this plugin tracks. It is visually distinct from Finance Tracker's `wallet` and Quick Notes' `notebook-pen`.

**Why `#8B5CF6` (violet):** This is the same violet used for Quedamos as a thematic color, but more importantly it is distinct from Finance Tracker's emerald (#10B981) and Quick Notes' indigo (#6366F1). The three plugin colors create a visually balanced set: emerald, indigo, violet.

---

## 12. Post-MVP Roadmap

Ordered by estimated impact vs effort:

### v0.2 -- Polish

| Feature | Impact | Effort |
|---|---|---|
| Drag-and-drop reordering within categories | Medium | Medium |
| Markdown rendering in notes field | Low | Low |
| "Absorbed by" formal field (links to another project) | Low | Low |
| Bulk status change (multi-select) | Low | Medium |

### v0.3 -- Insights

| Feature | Impact | Effort |
|---|---|---|
| Tech stack tag system with autocomplete | Medium | Medium |
| "Filter by tech" (show all projects using React) | Medium | Low (if tags exist) |
| Project timeline (log of status changes with dates) | Medium | Medium |
| Stats view: tech distribution chart, status distribution chart | Low | Medium |

### v0.4 -- Integration

| Feature | Impact | Effort |
|---|---|---|
| GitHub API: auto-fetch latest version tag, last commit date | High | High |
| Uptime check: ping web_url and show up/down badge | Medium | Medium |
| Export ecosystem as JSON/markdown | Low | Low |

### Future / Maybe

- Link to Cortex's other plugins (e.g., Finance Tracker transactions tagged by project)
- RSS feed of project updates
- "Ecosystem map" visual (node graph of project relationships and absorptions)

---

## 13. Open Questions

1. **Should absorbed projects be hidden by default?** Recommendation: No. Show them dimmed/strikethrough but visible. They are part of the history. The status filter allows hiding them explicitly.

2. **Should the seed migration be a separate SQL file?** Recommendation: Yes. Keep `001_schema.sql` for structure and `002_seed.sql` for initial data. This makes it easy to reset data without losing schema.

3. **Icon picker scope?** The form needs a curated list of Lucide icons. Suggestion: expose ~30 icons relevant to software projects (folder, code, globe, smartphone, terminal, database, server, rocket, wrench, gamepad, music, book, shopping-cart, heart, shield, etc.) rather than the full 1000+ Lucide catalog. Less choice paradox, faster to pick.

---

## 14. Implementation Notes for Backend/Frontend Agents

### Backend (Go)

- Follow the exact same structure as `plugins/finance-tracker/backend/`: `main.go` (calls `sdk.Serve`), `plugin.go` (implements `CortexPlugin`), `migrations/` folder with embedded SQL.
- Use path-based routing in `HandleAPI` with the same switch pattern.
- Input validation at the handler level (same pattern as Finance Tracker).
- Slug auto-generation: lowercase the name, replace spaces and special chars with hyphens, trim trailing hyphens. Validate uniqueness against DB.
- `updated_at` must be set on every UPDATE using `datetime('now')`.

### Frontend (SvelteKit)

- Create `ProjectHubPage.svelte` in `frontend/src/lib/components/plugins/`.
- Use `pluginApi('project-hub')` for all API calls.
- Reuse `WidgetCard.svelte` for the dashboard widget.
- Status badge component: small colored dot + text label. Reusable.
- Modal component: check if Cortex has a shared modal already. If not, create one following IDENTITY.md patterns (radius-lg, border, shadow-lg, backdrop).
- i18n keys under `projectHub.*` namespace in both `en.json` and `es.json`.
- Filter state in URL query params so filters persist on page refresh.
