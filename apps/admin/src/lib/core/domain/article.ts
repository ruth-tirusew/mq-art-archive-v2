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
  version: number;
  created_at: string;
  updated_at: string;
}

export interface ArticleRevision {
  id: string;
  article_id: string;
  version: number;
  editor_id: string;
  title: string;
  body: string;
  slug: string;
  category: string;
  excerpt?: string;
  reading_time: number;
  difficulty: string;
  verified: boolean;
  status: ArticleStatus;
  created_at: string;
}

export interface ArticleWrite {
  title: string;
  body?: string;
  category?: string;
  excerpt?: string;
  difficulty?: string;
  verified?: boolean;
  status?: ArticleStatus;
}

export interface ArticlePatch {
  status?: ArticleStatus;
  verified?: boolean;
}
