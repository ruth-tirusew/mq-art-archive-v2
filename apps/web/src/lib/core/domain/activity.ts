export interface ActivityHighlight {
	id: string;
	article_id: string;
	article_slug: string;
	article_title: string;
	quoted_text: string;
	created_at: string;
}

export interface ActivityComment {
	id: string;
	article_id: string;
	article_slug: string;
	article_title: string;
	body: string;
	is_general: boolean;
	created_at: string;
}
