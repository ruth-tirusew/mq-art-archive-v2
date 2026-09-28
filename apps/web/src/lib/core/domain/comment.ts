export interface Comment {
	id: string;
	user_id: string;
	author_name: string;
	body: string;
	is_general: boolean;
	quoted_text: string;
	start: number;
	end: number;
	matched: boolean;
	created_at: string;
}
