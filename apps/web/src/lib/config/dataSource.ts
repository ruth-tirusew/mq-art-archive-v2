import { PUBLIC_API_URL } from '$env/static/public';

/** When true, route loaders fetch from the mq Go API instead of static seed data. */
export const useApi = Boolean(PUBLIC_API_URL);
