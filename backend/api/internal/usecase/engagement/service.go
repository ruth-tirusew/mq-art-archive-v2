package engagement

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/apperrors"
	domain "github.com/mq/api/internal/domain/content"
	"github.com/mq/api/internal/domain/engagement"
	"github.com/mq/api/internal/port/inbound"
	"github.com/mq/api/internal/port/outbound"
)

// minPopularCount is how many distinct readers must have overlapping highlights before a
// passage counts as "popular" — a single highlighter's own selection isn't a trend.
const minPopularCount = 2

type Service struct {
	favorites  outbound.FavoriteRepository
	highlights outbound.HighlightRepository
	comments   outbound.CommentRepository
	content    inbound.ContentService
	identity   inbound.IdentityService
}

func NewService(
	favorites outbound.FavoriteRepository,
	highlights outbound.HighlightRepository,
	comments outbound.CommentRepository,
	content inbound.ContentService,
	identity inbound.IdentityService,
) inbound.EngagementService {
	return &Service{favorites: favorites, highlights: highlights, comments: comments, content: content, identity: identity}
}

func (s *Service) ToggleFavorite(ctx context.Context, userID, articleID uuid.UUID) (bool, int, error) {
	if _, err := s.content.AdminGet(ctx, articleID); err != nil {
		return false, 0, err
	}

	already, err := s.favorites.IsFavorited(ctx, userID, articleID)
	if err != nil {
		return false, 0, err
	}

	if already {
		if err := s.favorites.Remove(ctx, userID, articleID); err != nil {
			return false, 0, err
		}
	} else if err := s.favorites.Add(ctx, userID, articleID); err != nil {
		return false, 0, err
	}

	count, err := s.favorites.CountForArticle(ctx, articleID)
	if err != nil {
		return false, 0, err
	}
	return !already, count, nil
}

func (s *Service) FavoriteStatus(ctx context.Context, viewerID, articleID uuid.UUID) (bool, int, error) {
	count, err := s.favorites.CountForArticle(ctx, articleID)
	if err != nil {
		return false, 0, err
	}
	if viewerID == uuid.Nil {
		return false, count, nil
	}
	favorited, err := s.favorites.IsFavorited(ctx, viewerID, articleID)
	if err != nil {
		return false, 0, err
	}
	return favorited, count, nil
}

func (s *Service) ListMyFavorites(ctx context.Context, userID uuid.UUID) ([]domain.Article, error) {
	ids, err := s.favorites.ListArticleIDsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	articles := make([]domain.Article, 0, len(ids))
	for _, id := range ids {
		article, err := s.content.AdminGet(ctx, id)
		if err != nil {
			continue // favorited article was since deleted — skip rather than fail the list
		}
		articles = append(articles, *article)
	}
	return articles, nil
}

// validateSpan checks that quotedText/start/end describe a real, in-bounds substring of
// body — the one piece of input validation both CreateHighlight and an anchored
// CreateComment need before trusting a client-computed span. start/end are rune offsets,
// matching what the frontend computes and what Anchor stores — see engagement.Anchor's
// doc comment for why not byte offsets.
func validateSpan(body, quotedText string, start, end int) error {
	if quotedText == "" {
		return fmt.Errorf("%w: invalid anchor span", apperrors.ErrValidation)
	}
	slice, ok := engagement.Slice(body, start, end)
	if !ok {
		return fmt.Errorf("%w: invalid anchor span", apperrors.ErrValidation)
	}
	if slice != quotedText {
		return fmt.Errorf("%w: quoted text does not match the article body at that span", apperrors.ErrValidation)
	}
	return nil
}

func (s *Service) CreateHighlight(ctx context.Context, userID, articleID uuid.UUID, quotedText string, start, end int) (*inbound.ResolvedHighlight, error) {
	article, err := s.content.AdminGet(ctx, articleID)
	if err != nil {
		return nil, err
	}
	if err := validateSpan(article.Body, quotedText, start, end); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	h, err := s.highlights.Create(ctx, engagement.Highlight{
		ID:        uuid.New(),
		ArticleID: articleID,
		UserID:    userID,
		Anchor: engagement.Anchor{
			QuotedText: quotedText, TextOffsetStart: start, TextOffsetEnd: end,
			ArticleVersionAtAnchor: article.Version,
		},
		CreatedAt: now,
	})
	if err != nil {
		return nil, err
	}
	return &inbound.ResolvedHighlight{
		ID: h.ID, UserID: h.UserID, QuotedText: h.QuotedText,
		Matched: true, Start: start, End: end, CreatedAt: h.CreatedAt,
	}, nil
}

func (s *Service) DeleteHighlight(ctx context.Context, userID, highlightID uuid.UUID) error {
	return s.highlights.Delete(ctx, highlightID, userID)
}

func (s *Service) ListMyHighlights(ctx context.Context, userID, articleID uuid.UUID) ([]inbound.ResolvedHighlight, error) {
	article, err := s.content.AdminGet(ctx, articleID)
	if err != nil {
		return nil, err
	}
	mine, err := s.highlights.ListMineByArticle(ctx, articleID, userID)
	if err != nil {
		return nil, err
	}

	out := make([]inbound.ResolvedHighlight, len(mine))
	for i, h := range mine {
		res := h.Anchor.Resolve(article.Body, article.Version)
		out[i] = inbound.ResolvedHighlight{
			ID: h.ID, UserID: h.UserID, QuotedText: h.QuotedText,
			Matched: res.Matched, Start: res.Start, End: res.End, CreatedAt: h.CreatedAt,
		}
	}
	return out, nil
}

func (s *Service) PopularHighlight(ctx context.Context, articleID uuid.UUID) (*inbound.PopularHighlight, error) {
	article, err := s.content.AdminGet(ctx, articleID)
	if err != nil {
		return nil, err
	}
	all, err := s.highlights.ListByArticle(ctx, articleID)
	if err != nil {
		return nil, err
	}

	anchors := make([]engagement.Anchor, len(all))
	for i, h := range all {
		anchors[i] = h.Anchor
	}
	resolved := engagement.ResolveAll(anchors, article.Body, article.Version)
	counts := engagement.Coverage(engagement.RuneLen(article.Body), resolved)
	start, end, count := engagement.MostHighlighted(counts, minPopularCount)
	if count == 0 {
		return nil, nil
	}
	return &inbound.PopularHighlight{Start: start, End: end, Count: count}, nil
}

func (s *Service) CreateComment(ctx context.Context, userID, articleID uuid.UUID, body, quotedText string, start, end int, isGeneral bool) (*inbound.ResolvedComment, error) {
	if strings.TrimSpace(body) == "" {
		return nil, fmt.Errorf("%w: comment body is required", apperrors.ErrValidation)
	}
	article, err := s.content.AdminGet(ctx, articleID)
	if err != nil {
		return nil, err
	}

	anchor := engagement.Anchor{}
	if !isGeneral {
		if err := validateSpan(article.Body, quotedText, start, end); err != nil {
			return nil, err
		}
		anchor = engagement.Anchor{QuotedText: quotedText, TextOffsetStart: start, TextOffsetEnd: end, ArticleVersionAtAnchor: article.Version}
	}

	now := time.Now().UTC()
	c, err := s.comments.Create(ctx, engagement.Comment{
		ID: uuid.New(), ArticleID: articleID, UserID: userID, Body: strings.TrimSpace(body),
		IsGeneral: isGeneral, Anchor: anchor, CreatedAt: now,
	})
	if err != nil {
		return nil, err
	}
	return s.toResolvedComment(ctx, *c, article.Body, article.Version)
}

func (s *Service) DeleteComment(ctx context.Context, userID, commentID uuid.UUID) error {
	return s.comments.Delete(ctx, commentID, userID)
}

func (s *Service) ListComments(ctx context.Context, articleID uuid.UUID) ([]inbound.ResolvedComment, error) {
	article, err := s.content.AdminGet(ctx, articleID)
	if err != nil {
		return nil, err
	}
	all, err := s.comments.ListByArticle(ctx, articleID)
	if err != nil {
		return nil, err
	}

	out := make([]inbound.ResolvedComment, 0, len(all))
	for _, c := range all {
		resolved, err := s.toResolvedComment(ctx, c, article.Body, article.Version)
		if err != nil {
			return nil, err
		}
		out = append(out, *resolved)
	}
	return out, nil
}

func (s *Service) toResolvedComment(ctx context.Context, c engagement.Comment, body string, version int) (*inbound.ResolvedComment, error) {
	res := c.Anchor.Resolve(body, version)
	if c.IsGeneral {
		res.Matched = false
	}

	authorName := "Reader"
	if user, err := s.identity.GetUser(ctx, c.UserID); err == nil {
		if user.DisplayName != "" {
			authorName = user.DisplayName
		} else if at := strings.Index(user.Email, "@"); at > 0 {
			authorName = user.Email[:at]
		}
	}

	return &inbound.ResolvedComment{
		ID: c.ID, UserID: c.UserID, AuthorName: authorName, Body: c.Body, IsGeneral: c.IsGeneral,
		QuotedText: c.QuotedText, Matched: res.Matched, Start: res.Start, End: res.End, CreatedAt: c.CreatedAt,
	}, nil
}

func (s *Service) ListMyHighlightsAll(ctx context.Context, userID uuid.UUID) ([]inbound.ActivityHighlight, error) {
	items, err := s.highlights.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	out := make([]inbound.ActivityHighlight, 0, len(items))
	for _, h := range items {
		article, err := s.content.AdminGet(ctx, h.ArticleID)
		if err != nil {
			continue // article was since deleted — skip rather than fail the whole feed
		}
		out = append(out, inbound.ActivityHighlight{
			ID: h.ID, ArticleID: h.ArticleID, ArticleSlug: article.Slug, ArticleTitle: article.Title,
			QuotedText: h.QuotedText, CreatedAt: h.CreatedAt,
		})
	}
	return out, nil
}

func (s *Service) ListMyCommentsAll(ctx context.Context, userID uuid.UUID) ([]inbound.ActivityComment, error) {
	items, err := s.comments.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	out := make([]inbound.ActivityComment, 0, len(items))
	for _, c := range items {
		article, err := s.content.AdminGet(ctx, c.ArticleID)
		if err != nil {
			continue // article was since deleted — skip rather than fail the whole feed
		}
		out = append(out, inbound.ActivityComment{
			ID: c.ID, ArticleID: c.ArticleID, ArticleSlug: article.Slug, ArticleTitle: article.Title,
			Body: c.Body, IsGeneral: c.IsGeneral, CreatedAt: c.CreatedAt,
		})
	}
	return out, nil
}
