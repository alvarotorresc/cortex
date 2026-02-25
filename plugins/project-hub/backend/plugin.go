package main

import (
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	_ "modernc.org/sqlite"

	"github.com/alvarotorresc/cortex/pkg/sdk"
)

//go:embed migrations/*.sql
var migrations embed.FS

// ProjectHubPlugin implements sdk.CortexPlugin for project ecosystem tracking.
type ProjectHubPlugin struct {
	db *sql.DB
}

// GetManifest returns the plugin's metadata.
func (p *ProjectHubPlugin) GetManifest() (*sdk.Manifest, error) {
	return &sdk.Manifest{
		ID:          "project-hub",
		Name:        "Project Hub",
		Version:     "0.1.0",
		Description: "Track the state of your entire project ecosystem",
		Icon:        "folder-git-2",
		Color:       "#8B5CF6",
		Permissions: []string{"db:read", "db:write"},
	}, nil
}

// Migrate opens the SQLite database and runs embedded SQL migrations.
func (p *ProjectHubPlugin) Migrate(databasePath string) error {
	database, err := sql.Open("sqlite", databasePath)
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	p.db = database

	// Enable WAL mode for better concurrent read performance.
	if _, err := p.db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return fmt.Errorf("enabling WAL mode: %w", err)
	}

	// Enable foreign keys for CASCADE deletes.
	if _, err := p.db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		return fmt.Errorf("enabling foreign keys: %w", err)
	}

	schemaSQL, err := migrations.ReadFile("migrations/001_schema.sql")
	if err != nil {
		return fmt.Errorf("reading schema migration: %w", err)
	}

	if _, err := p.db.Exec(string(schemaSQL)); err != nil {
		return fmt.Errorf("running schema migration: %w", err)
	}

	seedSQL, err := migrations.ReadFile("migrations/002_seed.sql")
	if err != nil {
		return fmt.Errorf("reading seed migration: %w", err)
	}

	if _, err := p.db.Exec(string(seedSQL)); err != nil {
		return fmt.Errorf("running seed migration: %w", err)
	}

	tagsSQL, err := migrations.ReadFile("migrations/003_tags.sql")
	if err != nil {
		return fmt.Errorf("reading tags migration: %w", err)
	}

	if _, err := p.db.Exec(string(tagsSQL)); err != nil {
		return fmt.Errorf("running tags migration: %w", err)
	}

	return nil
}

// HandleAPI routes incoming API requests to the appropriate handler.
func (p *ProjectHubPlugin) HandleAPI(req *sdk.APIRequest) (*sdk.APIResponse, error) {
	switch {
	// Projects
	case req.Method == "GET" && req.Path == "/projects":
		return p.listProjects(req)
	case req.Method == "POST" && req.Path == "/projects":
		return p.createProject(req)
	case req.Method == "GET" && strings.HasPrefix(req.Path, "/projects/") && !strings.Contains(req.Path[len("/projects/"):], "/"):
		return p.getProject(req)
	case req.Method == "PUT" && strings.HasPrefix(req.Path, "/projects/") && !strings.Contains(req.Path[len("/projects/"):], "/"):
		return p.updateProject(req)
	case req.Method == "DELETE" && strings.HasPrefix(req.Path, "/projects/") && !strings.Contains(req.Path[len("/projects/"):], "/"):
		return p.deleteProject(req)

	// Project links
	case req.Method == "POST" && strings.HasPrefix(req.Path, "/projects/") && strings.HasSuffix(req.Path, "/links"):
		return p.createLink(req)
	case req.Method == "PUT" && strings.HasPrefix(req.Path, "/links/"):
		return p.updateLink(req)
	case req.Method == "DELETE" && strings.HasPrefix(req.Path, "/links/"):
		return p.deleteLink(req)

	// Tags
	case req.Method == "GET" && req.Path == "/tags":
		return p.listTags(req)
	case req.Method == "POST" && req.Path == "/tags":
		return p.createTag(req)
	case req.Method == "DELETE" && strings.HasPrefix(req.Path, "/tags/"):
		return p.deleteTag(req)

	// Project tags
	case req.Method == "POST" && strings.HasPrefix(req.Path, "/projects/") && strings.HasSuffix(req.Path, "/tags"):
		return p.setProjectTags(req)

	default:
		return jsonError(404, "NOT_FOUND", "route not found")
	}
}

// GetWidgetData returns dashboard widget data for the requested slot.
func (p *ProjectHubPlugin) GetWidgetData(slot string) ([]byte, error) {
	if slot != "dashboard-widget" {
		return json.Marshal(map[string]interface{}{"data": nil})
	}

	rows, err := p.db.Query("SELECT status, COUNT(*) FROM projects GROUP BY status")
	if err != nil {
		return nil, fmt.Errorf("querying project counts: %w", err)
	}
	defer rows.Close()

	byStatus := make(map[string]int)
	total := 0
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("scanning status count: %w", err)
		}
		byStatus[status] = count
		total += count
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating status counts: %w", err)
	}

	return json.Marshal(map[string]interface{}{
		"data": map[string]interface{}{
			"total":     total,
			"by_status": byStatus,
		},
	})
}

// Teardown closes the database connection when the plugin is unloaded.
func (p *ProjectHubPlugin) Teardown() error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}

// --- Data types ---

// Project represents a project record.
type Project struct {
	ID        int64   `json:"id"`
	Name      string  `json:"name"`
	Slug      string  `json:"slug"`
	Tagline   string  `json:"tagline"`
	Status    string  `json:"status"`
	Category  string  `json:"category"`
	Version   *string `json:"version"`
	Stack     string  `json:"stack"`
	Icon      string  `json:"icon"`
	Color     string  `json:"color"`
	RepoURL   *string `json:"repo_url"`
	WebURL    *string `json:"web_url"`
	DocsURL   *string `json:"docs_url"`
	Hosting   *string `json:"hosting"`
	Notes     *string `json:"notes"`
	SortOrder int     `json:"sort_order"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

// Tag represents a technology tag with a display color.
type Tag struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

// ProjectLink represents a custom link attached to a project.
type ProjectLink struct {
	ID        int64  `json:"id"`
	ProjectID int64  `json:"project_id"`
	Label     string `json:"label"`
	URL       string `json:"url"`
	SortOrder int    `json:"sort_order"`
}

// ProjectWithTags is a project with its associated tags.
type ProjectWithTags struct {
	Project
	Tags []Tag `json:"tags"`
}

// ProjectWithLinksAndTags is a project with its associated links and tags.
type ProjectWithLinksAndTags struct {
	Project
	Links []ProjectLink `json:"links"`
	Tags  []Tag         `json:"tags"`
}

// --- Handlers ---

func (p *ProjectHubPlugin) listProjects(req *sdk.APIRequest) (*sdk.APIResponse, error) {
	query := "SELECT DISTINCT p.id, p.name, p.slug, p.tagline, p.status, p.category, p.version, p.stack, p.icon, p.color, p.repo_url, p.web_url, p.docs_url, p.hosting, p.notes, p.sort_order, p.created_at, p.updated_at FROM projects p"
	args := make([]interface{}, 0)
	joins := ""
	wheres := make([]string, 0)

	if tag := req.Query["tag"]; tag != "" {
		joins += " JOIN project_tags pt ON pt.project_id = p.id JOIN tags t ON t.id = pt.tag_id"
		wheres = append(wheres, "t.name = ?")
		args = append(args, tag)
	}

	if status := req.Query["status"]; status != "" {
		wheres = append(wheres, "p.status = ?")
		args = append(args, status)
	}

	if category := req.Query["category"]; category != "" {
		wheres = append(wheres, "p.category = ?")
		args = append(args, category)
	}

	if search := req.Query["search"]; search != "" {
		wheres = append(wheres, "p.name LIKE ? ESCAPE '\\'")
		escaped := escapeLike(search)
		args = append(args, "%"+escaped+"%")
	}

	query += joins
	if len(wheres) > 0 {
		query += " WHERE " + strings.Join(wheres, " AND ")
	}
	query += " ORDER BY p.category, p.sort_order, p.name"

	rows, err := p.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying projects: %w", err)
	}
	defer rows.Close()

	projects := make([]ProjectWithTags, 0)
	for rows.Next() {
		var proj ProjectWithTags
		if err := rows.Scan(
			&proj.ID, &proj.Name, &proj.Slug, &proj.Tagline, &proj.Status, &proj.Category,
			&proj.Version, &proj.Stack, &proj.Icon, &proj.Color, &proj.RepoURL, &proj.WebURL,
			&proj.DocsURL, &proj.Hosting, &proj.Notes, &proj.SortOrder, &proj.CreatedAt, &proj.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning project: %w", err)
		}
		projects = append(projects, proj)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating projects: %w", err)
	}

	// Fetch tags for all projects in a single query.
	if len(projects) > 0 {
		ids := make([]interface{}, len(projects))
		placeholders := make([]string, len(projects))
		idToIdx := make(map[int64]int)
		for i, proj := range projects {
			ids[i] = proj.ID
			placeholders[i] = "?"
			idToIdx[proj.ID] = i
			projects[i].Tags = make([]Tag, 0)
		}

		tagQuery := fmt.Sprintf(
			"SELECT pt.project_id, t.id, t.name, t.color FROM project_tags pt JOIN tags t ON t.id = pt.tag_id WHERE pt.project_id IN (%s) ORDER BY t.name",
			strings.Join(placeholders, ","),
		)

		tagRows, err := p.db.Query(tagQuery, ids...)
		if err != nil {
			return nil, fmt.Errorf("querying project tags: %w", err)
		}
		defer tagRows.Close()

		for tagRows.Next() {
			var projectID int64
			var tag Tag
			if err := tagRows.Scan(&projectID, &tag.ID, &tag.Name, &tag.Color); err != nil {
				return nil, fmt.Errorf("scanning project tag: %w", err)
			}
			if idx, ok := idToIdx[projectID]; ok {
				projects[idx].Tags = append(projects[idx].Tags, tag)
			}
		}

		if err := tagRows.Err(); err != nil {
			return nil, fmt.Errorf("iterating project tags: %w", err)
		}
	}

	return jsonSuccess(200, projects)
}

func (p *ProjectHubPlugin) getProject(req *sdk.APIRequest) (*sdk.APIResponse, error) {
	slug := extractPathParam(req.Path, "/projects/")

	var proj Project
	err := p.db.QueryRow(
		`SELECT id, name, slug, tagline, status, category, version, stack, icon, color,
		        repo_url, web_url, docs_url, hosting, notes, sort_order, created_at, updated_at
		 FROM projects WHERE slug = ?`, slug,
	).Scan(
		&proj.ID, &proj.Name, &proj.Slug, &proj.Tagline, &proj.Status, &proj.Category,
		&proj.Version, &proj.Stack, &proj.Icon, &proj.Color, &proj.RepoURL, &proj.WebURL,
		&proj.DocsURL, &proj.Hosting, &proj.Notes, &proj.SortOrder, &proj.CreatedAt, &proj.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return jsonError(404, "NOT_FOUND", "project not found")
	}
	if err != nil {
		return nil, fmt.Errorf("querying project: %w", err)
	}

	// Fetch links for this project.
	linkRows, err := p.db.Query(
		"SELECT id, project_id, label, url, sort_order FROM project_links WHERE project_id = ? ORDER BY sort_order, id",
		proj.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("querying links: %w", err)
	}
	defer linkRows.Close()

	links := make([]ProjectLink, 0)
	for linkRows.Next() {
		var link ProjectLink
		if err := linkRows.Scan(&link.ID, &link.ProjectID, &link.Label, &link.URL, &link.SortOrder); err != nil {
			return nil, fmt.Errorf("scanning link: %w", err)
		}
		links = append(links, link)
	}

	if err := linkRows.Err(); err != nil {
		return nil, fmt.Errorf("iterating links: %w", err)
	}

	// Fetch tags for this project.
	tagRows, err := p.db.Query(
		"SELECT t.id, t.name, t.color FROM tags t JOIN project_tags pt ON pt.tag_id = t.id WHERE pt.project_id = ? ORDER BY t.name",
		proj.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("querying tags: %w", err)
	}
	defer tagRows.Close()

	tags := make([]Tag, 0)
	for tagRows.Next() {
		var tag Tag
		if err := tagRows.Scan(&tag.ID, &tag.Name, &tag.Color); err != nil {
			return nil, fmt.Errorf("scanning tag: %w", err)
		}
		tags = append(tags, tag)
	}

	if err := tagRows.Err(); err != nil {
		return nil, fmt.Errorf("iterating tags: %w", err)
	}

	result := ProjectWithLinksAndTags{
		Project: proj,
		Links:   links,
		Tags:    tags,
	}

	return jsonSuccess(200, result)
}

func (p *ProjectHubPlugin) createProject(req *sdk.APIRequest) (*sdk.APIResponse, error) {
	var input struct {
		Name      string  `json:"name"`
		Tagline   string  `json:"tagline"`
		Status    string  `json:"status"`
		Category  string  `json:"category"`
		Version   *string `json:"version"`
		Stack     string  `json:"stack"`
		Icon      string  `json:"icon"`
		Color     string  `json:"color"`
		RepoURL   *string `json:"repo_url"`
		WebURL    *string `json:"web_url"`
		DocsURL   *string `json:"docs_url"`
		Hosting   *string `json:"hosting"`
		Notes     *string `json:"notes"`
		SortOrder int     `json:"sort_order"`
		TagIDs    []int64 `json:"tag_ids"`
	}

	if err := json.Unmarshal(req.Body, &input); err != nil {
		return jsonError(400, "VALIDATION_ERROR", "invalid JSON body")
	}

	// Validate required fields.
	if strings.TrimSpace(input.Name) == "" {
		return jsonError(400, "VALIDATION_ERROR", "name is required")
	}
	if len(input.Name) > 100 {
		return jsonError(400, "VALIDATION_ERROR", "name must be 100 characters or less")
	}
	if strings.TrimSpace(input.Tagline) == "" {
		return jsonError(400, "VALIDATION_ERROR", "tagline is required")
	}
	if len(input.Tagline) > 200 {
		return jsonError(400, "VALIDATION_ERROR", "tagline must be 200 characters or less")
	}
	if !isValidStatus(input.Status) {
		return jsonError(400, "VALIDATION_ERROR", "status must be one of: concept, design, development, active, maintenance, archived, absorbed")
	}
	if input.Category != "flagship" && input.Category != "lab" {
		return jsonError(400, "VALIDATION_ERROR", "category must be 'flagship' or 'lab'")
	}
	if strings.TrimSpace(input.Stack) == "" {
		return jsonError(400, "VALIDATION_ERROR", "stack is required")
	}
	if input.Color != "" && !isValidHexColor(input.Color) {
		return jsonError(400, "VALIDATION_ERROR", "color must be a valid hex color (e.g. #0070F3)")
	}

	// Validate URL fields (prevent javascript: XSS).
	for _, u := range []*string{input.RepoURL, input.WebURL, input.DocsURL} {
		if u != nil && *u != "" && !isValidURL(*u) {
			return jsonError(400, "VALIDATION_ERROR", "URLs must use http:// or https://")
		}
	}

	// Generate slug from name.
	slug := toSlug(input.Name)

	// Defaults.
	if input.Icon == "" {
		input.Icon = "folder"
	}
	if input.Color == "" {
		input.Color = "#0070F3"
	}

	result, err := p.db.Exec(
		`INSERT INTO projects (name, slug, tagline, status, category, version, stack, icon, color, repo_url, web_url, docs_url, hosting, notes, sort_order)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		input.Name, slug, input.Tagline, input.Status, input.Category, input.Version,
		input.Stack, input.Icon, input.Color, input.RepoURL, input.WebURL, input.DocsURL,
		input.Hosting, input.Notes, input.SortOrder,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return jsonError(409, "CONFLICT", "a project with this name or slug already exists")
		}
		return nil, fmt.Errorf("inserting project: %w", err)
	}

	id, _ := result.LastInsertId()

	// Assign tags if provided.
	for _, tagID := range input.TagIDs {
		if _, err := p.db.Exec("INSERT OR IGNORE INTO project_tags (project_id, tag_id) VALUES (?, ?)", id, tagID); err != nil {
			return nil, fmt.Errorf("assigning tag: %w", err)
		}
	}

	return jsonSuccess(201, map[string]interface{}{"id": id, "slug": slug})
}

func (p *ProjectHubPlugin) updateProject(req *sdk.APIRequest) (*sdk.APIResponse, error) {
	slug := extractPathParam(req.Path, "/projects/")

	// Check project exists.
	var projectID int64
	err := p.db.QueryRow("SELECT id FROM projects WHERE slug = ?", slug).Scan(&projectID)
	if err == sql.ErrNoRows {
		return jsonError(404, "NOT_FOUND", "project not found")
	}
	if err != nil {
		return nil, fmt.Errorf("querying project: %w", err)
	}

	var input struct {
		Name      *string `json:"name"`
		Tagline   *string `json:"tagline"`
		Status    *string `json:"status"`
		Category  *string `json:"category"`
		Version   *string `json:"version"`
		Stack     *string `json:"stack"`
		Icon      *string `json:"icon"`
		Color     *string `json:"color"`
		RepoURL   *string `json:"repo_url"`
		WebURL    *string `json:"web_url"`
		DocsURL   *string `json:"docs_url"`
		Hosting   *string `json:"hosting"`
		Notes     *string `json:"notes"`
		SortOrder *int    `json:"sort_order"`
	}

	if err := json.Unmarshal(req.Body, &input); err != nil {
		return jsonError(400, "VALIDATION_ERROR", "invalid JSON body")
	}

	// Build dynamic update query.
	setClauses := make([]string, 0)
	args := make([]interface{}, 0)

	if input.Name != nil {
		if strings.TrimSpace(*input.Name) == "" {
			return jsonError(400, "VALIDATION_ERROR", "name is required")
		}
		if len(*input.Name) > 100 {
			return jsonError(400, "VALIDATION_ERROR", "name must be 100 characters or less")
		}
		setClauses = append(setClauses, "name = ?")
		args = append(args, *input.Name)
	}
	if input.Tagline != nil {
		if strings.TrimSpace(*input.Tagline) == "" {
			return jsonError(400, "VALIDATION_ERROR", "tagline is required")
		}
		if len(*input.Tagline) > 200 {
			return jsonError(400, "VALIDATION_ERROR", "tagline must be 200 characters or less")
		}
		setClauses = append(setClauses, "tagline = ?")
		args = append(args, *input.Tagline)
	}
	if input.Status != nil {
		if !isValidStatus(*input.Status) {
			return jsonError(400, "VALIDATION_ERROR", "status must be one of: concept, design, development, active, maintenance, archived, absorbed")
		}
		setClauses = append(setClauses, "status = ?")
		args = append(args, *input.Status)
	}
	if input.Category != nil {
		if *input.Category != "flagship" && *input.Category != "lab" {
			return jsonError(400, "VALIDATION_ERROR", "category must be 'flagship' or 'lab'")
		}
		setClauses = append(setClauses, "category = ?")
		args = append(args, *input.Category)
	}
	if input.Version != nil {
		setClauses = append(setClauses, "version = ?")
		args = append(args, *input.Version)
	}
	if input.Stack != nil {
		if strings.TrimSpace(*input.Stack) == "" {
			return jsonError(400, "VALIDATION_ERROR", "stack is required")
		}
		setClauses = append(setClauses, "stack = ?")
		args = append(args, *input.Stack)
	}
	if input.Icon != nil {
		setClauses = append(setClauses, "icon = ?")
		args = append(args, *input.Icon)
	}
	if input.Color != nil {
		if *input.Color != "" && !isValidHexColor(*input.Color) {
			return jsonError(400, "VALIDATION_ERROR", "color must be a valid hex color (e.g. #0070F3)")
		}
		setClauses = append(setClauses, "color = ?")
		args = append(args, *input.Color)
	}
	if input.RepoURL != nil {
		if *input.RepoURL != "" && !isValidURL(*input.RepoURL) {
			return jsonError(400, "VALIDATION_ERROR", "URLs must use http:// or https://")
		}
		setClauses = append(setClauses, "repo_url = ?")
		args = append(args, *input.RepoURL)
	}
	if input.WebURL != nil {
		if *input.WebURL != "" && !isValidURL(*input.WebURL) {
			return jsonError(400, "VALIDATION_ERROR", "URLs must use http:// or https://")
		}
		setClauses = append(setClauses, "web_url = ?")
		args = append(args, *input.WebURL)
	}
	if input.DocsURL != nil {
		if *input.DocsURL != "" && !isValidURL(*input.DocsURL) {
			return jsonError(400, "VALIDATION_ERROR", "URLs must use http:// or https://")
		}
		setClauses = append(setClauses, "docs_url = ?")
		args = append(args, *input.DocsURL)
	}
	if input.Hosting != nil {
		setClauses = append(setClauses, "hosting = ?")
		args = append(args, *input.Hosting)
	}
	if input.Notes != nil {
		setClauses = append(setClauses, "notes = ?")
		args = append(args, *input.Notes)
	}
	if input.SortOrder != nil {
		setClauses = append(setClauses, "sort_order = ?")
		args = append(args, *input.SortOrder)
	}

	if len(setClauses) == 0 {
		return jsonError(400, "VALIDATION_ERROR", "no fields to update")
	}

	// Always update updated_at.
	setClauses = append(setClauses, "updated_at = datetime('now')")

	query := fmt.Sprintf("UPDATE projects SET %s WHERE id = ?", strings.Join(setClauses, ", "))
	args = append(args, projectID)

	if _, err := p.db.Exec(query, args...); err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return jsonError(409, "CONFLICT", "a project with this name already exists")
		}
		return nil, fmt.Errorf("updating project: %w", err)
	}

	return jsonSuccess(200, map[string]interface{}{"updated": slug})
}

func (p *ProjectHubPlugin) deleteProject(req *sdk.APIRequest) (*sdk.APIResponse, error) {
	slug := extractPathParam(req.Path, "/projects/")

	result, err := p.db.Exec("DELETE FROM projects WHERE slug = ?", slug)
	if err != nil {
		return nil, fmt.Errorf("deleting project: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return jsonError(404, "NOT_FOUND", "project not found")
	}

	return jsonSuccess(200, map[string]interface{}{"deleted": slug})
}

// --- Link handlers ---

func (p *ProjectHubPlugin) createLink(req *sdk.APIRequest) (*sdk.APIResponse, error) {
	// Extract slug from /projects/{slug}/links
	pathParts := strings.Split(strings.TrimPrefix(req.Path, "/"), "/")
	if len(pathParts) < 3 {
		return jsonError(400, "VALIDATION_ERROR", "invalid path")
	}
	slug := pathParts[1]

	// Find project ID.
	var projectID int64
	err := p.db.QueryRow("SELECT id FROM projects WHERE slug = ?", slug).Scan(&projectID)
	if err == sql.ErrNoRows {
		return jsonError(404, "NOT_FOUND", "project not found")
	}
	if err != nil {
		return nil, fmt.Errorf("querying project: %w", err)
	}

	var input struct {
		Label     string `json:"label"`
		URL       string `json:"url"`
		SortOrder int    `json:"sort_order"`
	}

	if err := json.Unmarshal(req.Body, &input); err != nil {
		return jsonError(400, "VALIDATION_ERROR", "invalid JSON body")
	}

	if strings.TrimSpace(input.Label) == "" {
		return jsonError(400, "VALIDATION_ERROR", "label is required")
	}
	if strings.TrimSpace(input.URL) == "" {
		return jsonError(400, "VALIDATION_ERROR", "url is required")
	}
	if !isValidURL(input.URL) {
		return jsonError(400, "VALIDATION_ERROR", "URLs must use http:// or https://")
	}

	result, err := p.db.Exec(
		"INSERT INTO project_links (project_id, label, url, sort_order) VALUES (?, ?, ?, ?)",
		projectID, input.Label, input.URL, input.SortOrder,
	)
	if err != nil {
		return nil, fmt.Errorf("inserting link: %w", err)
	}

	id, _ := result.LastInsertId()
	return jsonSuccess(201, map[string]interface{}{"id": id})
}

func (p *ProjectHubPlugin) updateLink(req *sdk.APIRequest) (*sdk.APIResponse, error) {
	id := extractPathParam(req.Path, "/links/")

	var input struct {
		Label     *string `json:"label"`
		URL       *string `json:"url"`
		SortOrder *int    `json:"sort_order"`
	}

	if err := json.Unmarshal(req.Body, &input); err != nil {
		return jsonError(400, "VALIDATION_ERROR", "invalid JSON body")
	}

	setClauses := make([]string, 0)
	args := make([]interface{}, 0)

	if input.Label != nil {
		if strings.TrimSpace(*input.Label) == "" {
			return jsonError(400, "VALIDATION_ERROR", "label is required")
		}
		setClauses = append(setClauses, "label = ?")
		args = append(args, *input.Label)
	}
	if input.URL != nil {
		if strings.TrimSpace(*input.URL) == "" {
			return jsonError(400, "VALIDATION_ERROR", "url is required")
		}
		if !isValidURL(*input.URL) {
			return jsonError(400, "VALIDATION_ERROR", "URLs must use http:// or https://")
		}
		setClauses = append(setClauses, "url = ?")
		args = append(args, *input.URL)
	}
	if input.SortOrder != nil {
		setClauses = append(setClauses, "sort_order = ?")
		args = append(args, *input.SortOrder)
	}

	if len(setClauses) == 0 {
		return jsonError(400, "VALIDATION_ERROR", "no fields to update")
	}

	query := fmt.Sprintf("UPDATE project_links SET %s WHERE id = ?", strings.Join(setClauses, ", "))
	args = append(args, id)

	result, err := p.db.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("updating link: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return jsonError(404, "NOT_FOUND", "link not found")
	}

	return jsonSuccess(200, map[string]interface{}{"updated": id})
}

func (p *ProjectHubPlugin) deleteLink(req *sdk.APIRequest) (*sdk.APIResponse, error) {
	id := extractPathParam(req.Path, "/links/")

	result, err := p.db.Exec("DELETE FROM project_links WHERE id = ?", id)
	if err != nil {
		return nil, fmt.Errorf("deleting link: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return jsonError(404, "NOT_FOUND", "link not found")
	}

	return jsonSuccess(200, map[string]interface{}{"deleted": id})
}

// --- Tag handlers ---

func (p *ProjectHubPlugin) listTags(req *sdk.APIRequest) (*sdk.APIResponse, error) {
	rows, err := p.db.Query("SELECT id, name, color FROM tags ORDER BY name")
	if err != nil {
		return nil, fmt.Errorf("querying tags: %w", err)
	}
	defer rows.Close()

	tags := make([]Tag, 0)
	for rows.Next() {
		var tag Tag
		if err := rows.Scan(&tag.ID, &tag.Name, &tag.Color); err != nil {
			return nil, fmt.Errorf("scanning tag: %w", err)
		}
		tags = append(tags, tag)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating tags: %w", err)
	}

	return jsonSuccess(200, tags)
}

func (p *ProjectHubPlugin) createTag(req *sdk.APIRequest) (*sdk.APIResponse, error) {
	var input struct {
		Name  string `json:"name"`
		Color string `json:"color"`
	}

	if err := json.Unmarshal(req.Body, &input); err != nil {
		return jsonError(400, "VALIDATION_ERROR", "invalid JSON body")
	}

	if strings.TrimSpace(input.Name) == "" {
		return jsonError(400, "VALIDATION_ERROR", "name is required")
	}
	if len(input.Name) > 50 {
		return jsonError(400, "VALIDATION_ERROR", "name must be 50 characters or less")
	}
	if input.Color == "" {
		input.Color = "#6B7280"
	}
	if !isValidHexColor(input.Color) {
		return jsonError(400, "VALIDATION_ERROR", "color must be a valid hex color (e.g. #0070F3)")
	}

	result, err := p.db.Exec("INSERT INTO tags (name, color) VALUES (?, ?)", input.Name, input.Color)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return jsonError(409, "CONFLICT", "a tag with this name already exists")
		}
		return nil, fmt.Errorf("inserting tag: %w", err)
	}

	id, _ := result.LastInsertId()
	return jsonSuccess(201, map[string]interface{}{"id": id, "name": input.Name, "color": input.Color})
}

func (p *ProjectHubPlugin) deleteTag(req *sdk.APIRequest) (*sdk.APIResponse, error) {
	id := extractPathParam(req.Path, "/tags/")

	result, err := p.db.Exec("DELETE FROM tags WHERE id = ?", id)
	if err != nil {
		return nil, fmt.Errorf("deleting tag: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return jsonError(404, "NOT_FOUND", "tag not found")
	}

	return jsonSuccess(200, map[string]interface{}{"deleted": id})
}

// setProjectTags replaces all tags for a project with the given tag IDs.
func (p *ProjectHubPlugin) setProjectTags(req *sdk.APIRequest) (*sdk.APIResponse, error) {
	// Extract slug from /projects/{slug}/tags
	pathParts := strings.Split(strings.TrimPrefix(req.Path, "/"), "/")
	if len(pathParts) < 3 {
		return jsonError(400, "VALIDATION_ERROR", "invalid path")
	}
	slug := pathParts[1]

	// Find project ID.
	var projectID int64
	err := p.db.QueryRow("SELECT id FROM projects WHERE slug = ?", slug).Scan(&projectID)
	if err == sql.ErrNoRows {
		return jsonError(404, "NOT_FOUND", "project not found")
	}
	if err != nil {
		return nil, fmt.Errorf("querying project: %w", err)
	}

	var input struct {
		TagIDs []int64 `json:"tag_ids"`
	}

	if err := json.Unmarshal(req.Body, &input); err != nil {
		return jsonError(400, "VALIDATION_ERROR", "invalid JSON body")
	}

	// Replace all tags in a transaction.
	tx, err := p.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("beginning transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Remove existing tags.
	if _, err := tx.Exec("DELETE FROM project_tags WHERE project_id = ?", projectID); err != nil {
		return nil, fmt.Errorf("removing existing tags: %w", err)
	}

	// Insert new tags.
	for _, tagID := range input.TagIDs {
		if _, err := tx.Exec("INSERT INTO project_tags (project_id, tag_id) VALUES (?, ?)", projectID, tagID); err != nil {
			return nil, fmt.Errorf("inserting project tag: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("committing transaction: %w", err)
	}

	// Update updated_at.
	if _, err := p.db.Exec("UPDATE projects SET updated_at = datetime('now') WHERE id = ?", projectID); err != nil {
		return nil, fmt.Errorf("updating project timestamp: %w", err)
	}

	return jsonSuccess(200, map[string]interface{}{"project": slug, "tag_count": len(input.TagIDs)})
}

// --- Helpers ---

var validStatuses = map[string]bool{
	"concept": true, "design": true, "development": true, "active": true,
	"maintenance": true, "archived": true, "absorbed": true,
}

func isValidStatus(s string) bool {
	return validStatuses[s]
}

var hexColorRegex = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

func isValidHexColor(c string) bool {
	return hexColorRegex.MatchString(c)
}

// isValidURL checks that a URL uses http or https scheme (prevents javascript: XSS).
func isValidURL(u string) bool {
	return strings.HasPrefix(u, "https://") || strings.HasPrefix(u, "http://")
}

var slugRegex = regexp.MustCompile(`[^a-z0-9]+`)

func toSlug(name string) string {
	slug := strings.ToLower(name)
	slug = slugRegex.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	return slug
}

func extractPathParam(path, prefix string) string {
	return strings.TrimPrefix(path, prefix)
}

// escapeLike escapes LIKE wildcard characters (%, _) in a search string.
func escapeLike(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}

// --- JSON response helpers ---

// jsonSuccess wraps data in `{ "data": ... }` format per PATTERNS.md.
func jsonSuccess(status int, data interface{}) (*sdk.APIResponse, error) {
	body, err := json.Marshal(map[string]interface{}{"data": data})
	if err != nil {
		return nil, fmt.Errorf("marshaling response: %w", err)
	}
	return &sdk.APIResponse{
		StatusCode:  status,
		Body:        body,
		ContentType: "application/json",
	}, nil
}

// jsonError wraps errors in `{ "error": { "code": ..., "message": ... } }` format per PATTERNS.md.
func jsonError(status int, code string, message string) (*sdk.APIResponse, error) {
	body, _ := json.Marshal(map[string]interface{}{
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
		},
	})
	return &sdk.APIResponse{
		StatusCode:  status,
		Body:        body,
		ContentType: "application/json",
	}, nil
}
