package content

import (
	"context"
	"errors"
	"testing"

	appdb "github.com/kareltilcer/ws-tilcer-karel/backend/internal/platform/db"
	"github.com/kareltilcer/ws-tilcer-karel/backend/internal/platform/httpx"
	"github.com/kareltilcer/ws-tilcer-karel/backend/internal/platform/registry"
	"github.com/kareltilcer/ws-tilcer-karel/backend/internal/platform/testsupport"
)

func newTestSvc(t *testing.T) *Service {
	t.Helper()
	db := testsupport.NewDB(t,
		registry.MigrationSource{Name: "platform", FS: appdb.MigrationsFS},
		registry.MigrationSource{Name: "content", FS: MigrationsFS},
	)
	return NewService(db)
}

// apiStatus extracts the HTTP status from a service error (APIError), or 0.
func apiStatus(err error) int {
	var ae *httpx.APIError
	if errors.As(err, &ae) {
		return ae.Status
	}
	return 0
}

func mustCategory(t *testing.T, s *Service, slug string) int64 {
	t.Helper()
	c, err := s.CreateCategory(context.Background(), CategoryCreate{Slug: slug, NameCS: "CS " + slug, NameEN: "EN " + slug})
	if err != nil {
		t.Fatalf("CreateCategory(%q): %v", slug, err)
	}
	return c.ID
}

func TestCategoryDuplicateSlugAndDeleteInUse(t *testing.T) {
	s := newTestSvc(t)
	ctx := context.Background()
	catID := mustCategory(t, s, "software")

	// duplicate slug → 409
	if _, err := s.CreateCategory(ctx, CategoryCreate{Slug: "software", NameCS: "x", NameEN: "x"}); apiStatus(err) != 409 {
		t.Fatalf("dup slug: want 409, got err=%v", err)
	}
	// bad slug → 422
	if _, err := s.CreateCategory(ctx, CategoryCreate{Slug: "Bad Slug", NameCS: "x", NameEN: "x"}); apiStatus(err) != 422 {
		t.Fatalf("bad slug: want 422, got err=%v", err)
	}
	// referencing project blocks delete → 409
	if _, err := s.CreateProject(ctx, ProjectCreate{Slug: "p1", CategoryID: catID, TitleCS: "a", TitleEN: "b"}); err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	if err := s.DeleteCategory(ctx, catID); apiStatus(err) != 409 {
		t.Fatalf("delete in-use: want 409, got err=%v", err)
	}
}

func TestPublicReadsHideDraftsAndReturnBothLanguages(t *testing.T) {
	s := newTestSvc(t)
	ctx := context.Background()
	catID := mustCategory(t, s, "software")

	pub := "published"
	if _, err := s.CreateProject(ctx, ProjectCreate{
		Slug: "live", CategoryID: catID, TitleCS: "Živá", TitleEN: "Live",
		SummaryCS: "Popis", SummaryEN: "Summary", Status: &pub,
	}); err != nil {
		t.Fatalf("create published: %v", err)
	}
	if _, err := s.CreateProject(ctx, ProjectCreate{
		Slug: "wip", CategoryID: catID, TitleCS: "Koncept", TitleEN: "Draft",
	}); err != nil {
		t.Fatalf("create draft: %v", err)
	}

	list, err := s.ListPublicProjects(ctx, "", nil)
	if err != nil {
		t.Fatalf("ListPublicProjects: %v", err)
	}
	if len(list) != 1 || list[0].Slug != "live" {
		t.Fatalf("public list = %+v, want only 'live'", list)
	}
	// both languages present on the published project
	if list[0].TitleCS != "Živá" || list[0].TitleEN != "Live" {
		t.Fatalf("bilingual titles missing: %+v", list[0])
	}

	// admin list returns both
	page, err := s.ListAdminProjects(ctx, "", "", 50)
	if err != nil {
		t.Fatalf("ListAdminProjects: %v", err)
	}
	if len(page.Items) != 2 {
		t.Fatalf("admin list = %d items, want 2", len(page.Items))
	}

	// unknown public slug → 404
	if _, err := s.GetPublicProject(ctx, "does-not-exist"); apiStatus(err) != 404 {
		t.Fatalf("unknown slug: want 404, got err=%v", err)
	}
	// draft is not publicly reachable → 404
	if _, err := s.GetPublicProject(ctx, "wip"); apiStatus(err) != 404 {
		t.Fatalf("draft public read: want 404, got err=%v", err)
	}
}

func TestUnknownCategoryRejected(t *testing.T) {
	s := newTestSvc(t)
	if _, err := s.CreateProject(context.Background(), ProjectCreate{
		Slug: "x", CategoryID: 999, TitleCS: "a", TitleEN: "b",
	}); apiStatus(err) != 422 {
		t.Fatalf("unknown category: want 422, got err=%v", err)
	}
}

func TestReorderRewritesSortOrder(t *testing.T) {
	s := newTestSvc(t)
	ctx := context.Background()
	a := mustCategory(t, s, "aaa")
	b := mustCategory(t, s, "bbb")
	c := mustCategory(t, s, "ccc")

	if err := s.ReorderCategories(ctx, []int64{c, a, b}); err != nil {
		t.Fatalf("reorder: %v", err)
	}
	cats, err := s.ListAllCategories(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	// ListCategories orders by sort_order — expect c, a, b.
	got := []int64{cats[0].ID, cats[1].ID, cats[2].ID}
	want := []int64{c, a, b}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("order = %v, want %v", got, want)
		}
	}
}

func TestLinkClickIncrements(t *testing.T) {
	s := newTestSvc(t)
	ctx := context.Background()
	l, err := s.CreateLink(ctx, LinkCreate{LabelCS: "GH", LabelEN: "GH", URL: "https://github.com/x"})
	if err != nil {
		t.Fatalf("CreateLink: %v", err)
	}
	if err := s.ClickLink(ctx, l.ID); err != nil {
		t.Fatalf("ClickLink: %v", err)
	}
	if err := s.ClickLink(ctx, l.ID); err != nil {
		t.Fatalf("ClickLink 2: %v", err)
	}
	// unknown id → 404
	if err := s.ClickLink(ctx, 9999); apiStatus(err) != 404 {
		t.Fatalf("unknown click: want 404, got err=%v", err)
	}
	links, err := s.ListAllLinks(ctx)
	if err != nil {
		t.Fatalf("list links: %v", err)
	}
	if links[0].ClickCount != 2 {
		t.Fatalf("click_count = %d, want 2", links[0].ClickCount)
	}
}

func TestBadLinkURLRejected(t *testing.T) {
	s := newTestSvc(t)
	if _, err := s.CreateLink(context.Background(), LinkCreate{LabelCS: "x", LabelEN: "x", URL: "javascript:alert(1)"}); apiStatus(err) != 422 {
		t.Fatalf("bad url: want 422, got err=%v", err)
	}
}

func TestSkillBilingualGroupRoundTrip(t *testing.T) {
	s := newTestSvc(t)
	ctx := context.Background()

	// Both group labels are required, mirroring the bilingual name.
	if _, err := s.CreateSkill(ctx, SkillCreate{NameCS: "Go", NameEN: "Go", CategoryCS: "jazyky"}); apiStatus(err) != 422 {
		t.Fatalf("missing category_en: want 422, got err=%v", err)
	}

	sk, err := s.CreateSkill(ctx, SkillCreate{NameCS: "Go", NameEN: "Go", CategoryCS: "jazyky", CategoryEN: "languages"})
	if err != nil {
		t.Fatalf("CreateSkill: %v", err)
	}
	if sk.CategoryCS != "jazyky" || sk.CategoryEN != "languages" {
		t.Fatalf("group = %q/%q, want jazyky/languages", sk.CategoryCS, sk.CategoryEN)
	}

	// Retitling the group in both languages survives a round-trip through the store.
	cs, en := "nástroje", "tools"
	if _, err := s.UpdateSkill(ctx, sk.ID, SkillUpdate{CategoryCS: &cs, CategoryEN: &en}); err != nil {
		t.Fatalf("UpdateSkill: %v", err)
	}
	skills, err := s.ListAllSkills(ctx)
	if err != nil {
		t.Fatalf("list skills: %v", err)
	}
	if len(skills) != 1 || skills[0].CategoryCS != "nástroje" || skills[0].CategoryEN != "tools" {
		t.Fatalf("after update: %+v", skills)
	}
}
