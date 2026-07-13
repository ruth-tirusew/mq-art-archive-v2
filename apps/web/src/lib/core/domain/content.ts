export type ArticleStatus = 'draft' | 'published' | 'archived';

export interface Article {
  id: string;
  slug: string;
  title: string;
  body: string;
  category?: string;
  excerpt?: string;
  reading_time?: number;
  difficulty?: string;
  verified?: boolean;
  contributors?: number;
  author_id: string;
  status: ArticleStatus;
  created_at: string;
  updated_at: string;
}

export interface ArticleListParams {
  category?: string;
  q?: string;
  limit?: number;
  offset?: number;
}
