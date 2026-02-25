-- Project Hub: seed data
-- Populates the 15 current ecosystem projects (7 flagship + 7 lab + 2 absorbed).

-- Flagship
INSERT OR IGNORE INTO projects (name, slug, tagline, status, category, version, stack, icon, color, repo_url, web_url, hosting, notes, sort_order) VALUES
    ('Quedamos', 'quedamos', 'Coordinar quedadas entre amigos', 'development', 'flagship', 'v0.2', 'React, Ionic, Capacitor, NestJS, Supabase', 'users', '#8B5CF6', 'https://github.com/alvarotorresc/quedamos-app', 'https://quedamos.app', 'Vercel + Railway + Supabase', 'Monorepo. v0.2 in development. Chat and map features planned.', 1),
    ('Sinherencia', 'sinherencia', 'Blog politico semanal', 'active', 'flagship', NULL, 'Astro 5', 'newspaper', '#DC2626', 'https://github.com/alvarotorresc/sinherencia', 'https://sinherencia.com', 'Vercel', 'Weekly political blog. Active and maintained.', 2),
    ('Guitar App', 'guitar-app', 'Herramientas para guitarra electrica', 'concept', 'flagship', NULL, 'SolidJS, Web Audio API', 'music', '#F59E0B', NULL, NULL, 'Vercel', 'Passion project. Tuner, scales, chord library, metronome.', 3),
    ('Cortex', 'cortex', 'Hub personal self-hosted con plugins', 'development', 'flagship', NULL, 'Go, SvelteKit, gRPC, SQLite', 'brain', '#10B981', 'https://github.com/alvarotorresc/cortex', NULL, 'Self-hosted (Docker Compose)', 'MVP in progress. Finance Tracker and Quick Notes plugins done.', 4),
    ('Fogon', 'fogon', 'App cocina y compra colaborativa', 'design', 'flagship', NULL, 'React Native (Expo), NestJS, Supabase', 'flame', '#C2410C', NULL, 'https://fogon.app', 'Vercel + Railway + Supabase', 'Design approved. Absorbs Price Tracker.', 5),
    ('Huellas', 'huellas', 'App animales perdidos', 'design', 'flagship', NULL, 'Next.js 15, React Native (Expo), NestJS, Supabase, PostGIS', 'paw-print', '#F97316', NULL, 'https://huellas.app', 'Vercel + Railway + Supabase', 'Design approved. Public product.', 6),
    ('Libroteca', 'libroteca', 'El gran producto', 'concept', 'flagship', NULL, 'TBD', 'book-open', '#0070F3', NULL, NULL, NULL, 'Everything builds toward this. Stack TBD.', 7);

-- Lab
INSERT OR IGNORE INTO projects (name, slug, tagline, status, category, version, stack, icon, color, repo_url, web_url, hosting, notes, sort_order) VALUES
    ('create-astro-blog', 'create-astro-blog', 'CLI scaffolder para blogs Astro', 'active', 'lab', NULL, 'Node.js CLI', 'terminal', '#0070F3', 'https://github.com/alvarotorresc/create-astro-blog', 'https://www.npmjs.com/package/create-astro-blog', 'npm', 'Published on npm. Done.', 1),
    ('PokeUtils', 'pokeutils', 'SPA Pokemon con utilidades', 'active', 'lab', NULL, 'HTML, CSS, JavaScript', 'gamepad-2', '#EF4444', 'https://github.com/alvarotorresc/pokeutils', 'https://pokeutils.alvarotc.com', 'Netlify', 'Vanilla JS. Active.', 2),
    ('DevTools', 'devtools', 'Suite herramientas para desarrolladores', 'active', 'lab', NULL, 'Astro 5, TypeScript', 'wrench', '#06B6D4', 'https://github.com/alvarotorresc/devtools', 'https://devtools.alvarotc.com', 'Netlify', 'Active. Multiple utilities.', 3),
    ('Clipboard Manager', 'clipboard-manager', 'Clipboard compartido en red local', 'concept', 'lab', NULL, 'Go, WebSocket', 'clipboard', '#0070F3', NULL, NULL, 'Self-hosted', 'Local network clipboard sharing.', 4),
    ('System Config Manager', 'system-config-manager', 'Config Linux portable', 'concept', 'lab', NULL, 'Rust CLI', 'settings', '#0070F3', NULL, NULL, NULL, 'Portable Linux configuration manager.', 5),
    ('Swiss Knife', 'swiss-knife', '14 herramientas random para Android', 'active', 'lab', 'v0.3', 'Kotlin, Jetpack Compose, Material 3', 'smartphone', '#0070F3', 'https://github.com/alvarotorresc/swiss-knife', NULL, 'APK', 'v0.3 done. 14 tools.', 6),
    ('IronLog', 'ironlog', 'Cuaderno de gym personal', 'design', 'lab', NULL, 'React Native (Expo), expo-sqlite', 'dumbbell', '#64748B', NULL, NULL, NULL, 'Personal gym log. Zero cloud. Design approved.', 7);

-- Absorbed
INSERT OR IGNORE INTO projects (name, slug, tagline, status, category, version, stack, icon, color, repo_url, hosting, notes, sort_order) VALUES
    ('Finance App', 'finance-app', 'App finanzas personales', 'absorbed', 'lab', NULL, 'Tauri 2, Svelte 5, SQLite', 'wallet', '#16A34A', NULL, NULL, 'Absorbed into Cortex as Finance Tracker plugin.', 8),
    ('Price Tracker', 'price-tracker', 'Tracker precios supermercado', 'absorbed', 'lab', NULL, 'Python, FastAPI, Turso', 'tag', '#EA580C', NULL, NULL, 'Absorbed into Fogon.', 9);
