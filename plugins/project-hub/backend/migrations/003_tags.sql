-- Project Hub: tags system
-- Replaces the free-text stack field with a proper tags system.
-- Each tag has a name and color. Projects are linked to tags via project_tags.

CREATE TABLE IF NOT EXISTS tags (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE COLLATE NOCASE,
    color TEXT NOT NULL DEFAULT '#6B7280'
);

CREATE TABLE IF NOT EXISTS project_tags (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    tag_id INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    UNIQUE(project_id, tag_id)
);

CREATE INDEX IF NOT EXISTS idx_project_tags_project_id ON project_tags(project_id);
CREATE INDEX IF NOT EXISTS idx_project_tags_tag_id ON project_tags(tag_id);

-- Seed common technology tags with brand colors.
INSERT OR IGNORE INTO tags (name, color) VALUES
    ('React', '#61DAFB'),
    ('React Native', '#61DAFB'),
    ('Ionic', '#3880FF'),
    ('Capacitor', '#119EFF'),
    ('NestJS', '#E0234E'),
    ('Supabase', '#3ECF8E'),
    ('Astro', '#FF5D01'),
    ('SolidJS', '#2C4F7C'),
    ('Web Audio API', '#4A90D9'),
    ('Go', '#00ADD8'),
    ('SvelteKit', '#FF3E00'),
    ('gRPC', '#244C5A'),
    ('SQLite', '#003B57'),
    ('Expo', '#000020'),
    ('Next.js', '#000000'),
    ('PostGIS', '#336791'),
    ('Node.js', '#339933'),
    ('HTML', '#E34F26'),
    ('CSS', '#1572B6'),
    ('JavaScript', '#F7DF1E'),
    ('TypeScript', '#3178C6'),
    ('Kotlin', '#7F52FF'),
    ('Jetpack Compose', '#4285F4'),
    ('Material 3', '#6750A4'),
    ('Tauri', '#FFC131'),
    ('Svelte', '#FF3E00'),
    ('Python', '#3776AB'),
    ('FastAPI', '#009688'),
    ('Turso', '#4FF8D2'),
    ('Rust', '#CE422B'),
    ('WebSocket', '#010101'),
    ('Docker', '#2496ED'),
    ('TBD', '#6B7280');

-- Migrate existing stack data into tags.
-- For each project, split the stack field by comma and link to matching tags.
-- This uses a recursive CTE to split comma-separated values.

-- Quedamos: React, Ionic, Capacitor, NestJS, Supabase
INSERT OR IGNORE INTO project_tags (project_id, tag_id)
SELECT p.id, t.id FROM projects p, tags t
WHERE p.slug = 'quedamos' AND t.name IN ('React', 'Ionic', 'Capacitor', 'NestJS', 'Supabase');

-- Sinherencia: Astro 5 -> Astro
INSERT OR IGNORE INTO project_tags (project_id, tag_id)
SELECT p.id, t.id FROM projects p, tags t
WHERE p.slug = 'sinherencia' AND t.name IN ('Astro');

-- Guitar App: SolidJS, Web Audio API
INSERT OR IGNORE INTO project_tags (project_id, tag_id)
SELECT p.id, t.id FROM projects p, tags t
WHERE p.slug = 'guitar-app' AND t.name IN ('SolidJS', 'Web Audio API');

-- Cortex: Go, SvelteKit, gRPC, SQLite
INSERT OR IGNORE INTO project_tags (project_id, tag_id)
SELECT p.id, t.id FROM projects p, tags t
WHERE p.slug = 'cortex' AND t.name IN ('Go', 'SvelteKit', 'gRPC', 'SQLite');

-- Fogon: React Native (Expo), NestJS, Supabase -> React Native, Expo, NestJS, Supabase
INSERT OR IGNORE INTO project_tags (project_id, tag_id)
SELECT p.id, t.id FROM projects p, tags t
WHERE p.slug = 'fogon' AND t.name IN ('React Native', 'Expo', 'NestJS', 'Supabase');

-- Huellas: Next.js 15, React Native (Expo), NestJS, Supabase, PostGIS
INSERT OR IGNORE INTO project_tags (project_id, tag_id)
SELECT p.id, t.id FROM projects p, tags t
WHERE p.slug = 'huellas' AND t.name IN ('Next.js', 'React Native', 'Expo', 'NestJS', 'Supabase', 'PostGIS');

-- Libroteca: TBD
INSERT OR IGNORE INTO project_tags (project_id, tag_id)
SELECT p.id, t.id FROM projects p, tags t
WHERE p.slug = 'libroteca' AND t.name IN ('TBD');

-- create-astro-blog: Node.js CLI -> Node.js
INSERT OR IGNORE INTO project_tags (project_id, tag_id)
SELECT p.id, t.id FROM projects p, tags t
WHERE p.slug = 'create-astro-blog' AND t.name IN ('Node.js');

-- PokeUtils: HTML, CSS, JavaScript
INSERT OR IGNORE INTO project_tags (project_id, tag_id)
SELECT p.id, t.id FROM projects p, tags t
WHERE p.slug = 'pokeutils' AND t.name IN ('HTML', 'CSS', 'JavaScript');

-- DevTools: Astro 5, TypeScript -> Astro, TypeScript
INSERT OR IGNORE INTO project_tags (project_id, tag_id)
SELECT p.id, t.id FROM projects p, tags t
WHERE p.slug = 'devtools' AND t.name IN ('Astro', 'TypeScript');

-- Clipboard Manager: Go, WebSocket
INSERT OR IGNORE INTO project_tags (project_id, tag_id)
SELECT p.id, t.id FROM projects p, tags t
WHERE p.slug = 'clipboard-manager' AND t.name IN ('Go', 'WebSocket');

-- System Config Manager: Rust CLI -> Rust
INSERT OR IGNORE INTO project_tags (project_id, tag_id)
SELECT p.id, t.id FROM projects p, tags t
WHERE p.slug = 'system-config-manager' AND t.name IN ('Rust');

-- Swiss Knife: Kotlin, Jetpack Compose, Material 3
INSERT OR IGNORE INTO project_tags (project_id, tag_id)
SELECT p.id, t.id FROM projects p, tags t
WHERE p.slug = 'swiss-knife' AND t.name IN ('Kotlin', 'Jetpack Compose', 'Material 3');

-- IronLog: React Native (Expo), expo-sqlite -> React Native, Expo, SQLite
INSERT OR IGNORE INTO project_tags (project_id, tag_id)
SELECT p.id, t.id FROM projects p, tags t
WHERE p.slug = 'ironlog' AND t.name IN ('React Native', 'Expo', 'SQLite');

-- Finance App: Tauri 2, Svelte 5, SQLite -> Tauri, Svelte, SQLite
INSERT OR IGNORE INTO project_tags (project_id, tag_id)
SELECT p.id, t.id FROM projects p, tags t
WHERE p.slug = 'finance-app' AND t.name IN ('Tauri', 'Svelte', 'SQLite');

-- Price Tracker: Python, FastAPI, Turso
INSERT OR IGNORE INTO project_tags (project_id, tag_id)
SELECT p.id, t.id FROM projects p, tags t
WHERE p.slug = 'price-tracker' AND t.name IN ('Python', 'FastAPI', 'Turso');
