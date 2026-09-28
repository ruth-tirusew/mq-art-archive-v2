export interface Highlight {
	id: string;
	quoted_text: string;
	start: number;
	end: number;
	matched: boolean;
	created_at: string;
}

export interface PopularHighlight {
	start: number;
	end: number;
	count: number;
}
