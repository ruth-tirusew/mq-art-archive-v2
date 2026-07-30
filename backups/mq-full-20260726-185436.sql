--
-- PostgreSQL database dump
--

\restrict azZsFcvp4tMvust95Hhi3lHsThfqdUjBDDdxqXlLGXLqquTFrUKEY4cKw6HQCAq

-- Dumped from database version 16.14
-- Dumped by pg_dump version 16.14

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: art_post_media; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.art_post_media (
    id uuid NOT NULL,
    art_post_id uuid NOT NULL,
    url text NOT NULL,
    mime_type text DEFAULT ''::text NOT NULL,
    width integer DEFAULT 0 NOT NULL,
    height integer DEFAULT 0 NOT NULL,
    sort_order integer DEFAULT 0 NOT NULL,
    media_asset_id uuid
);


--
-- Name: art_posts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.art_posts (
    id uuid NOT NULL,
    artist_id uuid NOT NULL,
    title text NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    medium text DEFAULT ''::text NOT NULL,
    status text NOT NULL,
    published_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    year integer,
    dimensions text DEFAULT ''::text NOT NULL,
    city text DEFAULT ''::text NOT NULL,
    style text DEFAULT ''::text NOT NULL,
    featured_acquisition boolean DEFAULT false NOT NULL,
    palette text[] DEFAULT '{}'::text[] NOT NULL,
    residency text,
    exhibition text,
    CONSTRAINT art_posts_status_check CHECK ((status = ANY (ARRAY['draft'::text, 'published'::text, 'archived'::text])))
);


--
-- Name: article_revisions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.article_revisions (
    id uuid NOT NULL,
    article_id uuid NOT NULL,
    version integer NOT NULL,
    editor_id uuid NOT NULL,
    title text NOT NULL,
    body text DEFAULT ''::text NOT NULL,
    slug text NOT NULL,
    category text DEFAULT 'General'::text NOT NULL,
    excerpt text DEFAULT ''::text NOT NULL,
    reading_time integer DEFAULT 1 NOT NULL,
    difficulty text DEFAULT 'Beginner'::text NOT NULL,
    verified boolean DEFAULT false NOT NULL,
    status text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: article_submissions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.article_submissions (
    id uuid NOT NULL,
    submitter_id uuid NOT NULL,
    article_id uuid,
    title text NOT NULL,
    body text NOT NULL,
    status text DEFAULT 'pending'::text NOT NULL,
    review_notes text DEFAULT ''::text NOT NULL,
    reviewed_by uuid,
    reviewed_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT article_submissions_status_check CHECK ((status = ANY (ARRAY['pending'::text, 'approved'::text, 'rejected'::text])))
);


--
-- Name: articles; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.articles (
    id uuid NOT NULL,
    slug text NOT NULL,
    title text NOT NULL,
    body text DEFAULT ''::text NOT NULL,
    status text NOT NULL,
    author_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    category text DEFAULT 'General'::text NOT NULL,
    excerpt text DEFAULT ''::text NOT NULL,
    reading_time integer DEFAULT 10 NOT NULL,
    difficulty text DEFAULT 'Beginner'::text NOT NULL,
    verified boolean DEFAULT false NOT NULL,
    contributors integer DEFAULT 1 NOT NULL,
    search_vector tsvector GENERATED ALWAYS AS (to_tsvector('english'::regconfig, ((((COALESCE(title, ''::text) || ' '::text) || COALESCE(excerpt, ''::text)) || ' '::text) || COALESCE(body, ''::text)))) STORED,
    version integer DEFAULT 1 NOT NULL,
    CONSTRAINT articles_status_check CHECK ((status = ANY (ARRAY['draft'::text, 'published'::text, 'archived'::text])))
);


--
-- Name: artist_profiles; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.artist_profiles (
    id uuid NOT NULL,
    user_id uuid NOT NULL,
    slug text NOT NULL,
    display_name text NOT NULL,
    bio text DEFAULT ''::text NOT NULL,
    contact_email text DEFAULT ''::text NOT NULL,
    contact_phone text DEFAULT ''::text NOT NULL,
    contact_website text DEFAULT ''::text NOT NULL,
    contact_location text DEFAULT ''::text NOT NULL,
    social_instagram text DEFAULT ''::text NOT NULL,
    social_twitter text DEFAULT ''::text NOT NULL,
    social_telegram text DEFAULT ''::text NOT NULL,
    status text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    handle text,
    born text DEFAULT ''::text NOT NULL,
    discipline text DEFAULT ''::text NOT NULL,
    tagline text DEFAULT ''::text NOT NULL,
    years_active text DEFAULT ''::text NOT NULL,
    featured boolean DEFAULT false NOT NULL,
    portrait_url text DEFAULT ''::text NOT NULL,
    influences text[] DEFAULT '{}'::text[] NOT NULL,
    in_residence boolean DEFAULT false NOT NULL,
    residency_place text,
    open_for_commission boolean DEFAULT false NOT NULL,
    approved_at timestamp with time zone,
    portrait_media_asset_id uuid,
    CONSTRAINT artist_profiles_status_check CHECK ((status = ANY (ARRAY['draft'::text, 'pending'::text, 'approved'::text])))
);


--
-- Name: email_verification_tokens; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.email_verification_tokens (
    id uuid NOT NULL,
    user_id uuid NOT NULL,
    token_hash text NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: event_locations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.event_locations (
    id uuid NOT NULL,
    name text NOT NULL,
    pin_coords double precision[],
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT event_locations_pin_coords_len CHECK (((pin_coords IS NULL) OR (array_length(pin_coords, 1) = 2)))
);


--
-- Name: events; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.events (
    id uuid NOT NULL,
    title text NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    source_url text NOT NULL,
    image_url text,
    location_id uuid,
    starts_at timestamp with time zone NOT NULL,
    ends_at timestamp with time zone,
    scraped_at timestamp with time zone NOT NULL,
    status text DEFAULT 'pending'::text NOT NULL,
    review_notes text DEFAULT ''::text NOT NULL,
    reviewed_by uuid,
    reviewed_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    search_vector tsvector GENERATED ALWAYS AS (to_tsvector('english'::regconfig, ((COALESCE(title, ''::text) || ' '::text) || COALESCE(description, ''::text)))) STORED,
    slug text,
    event_type text DEFAULT 'Opening'::text NOT NULL,
    venue text DEFAULT ''::text NOT NULL,
    city text DEFAULT ''::text NOT NULL,
    CONSTRAINT events_status_check CHECK ((status = ANY (ARRAY['pending'::text, 'approved'::text, 'rejected'::text])))
);


--
-- Name: goose_db_version; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.goose_db_version (
    id integer NOT NULL,
    version_id bigint NOT NULL,
    is_applied boolean NOT NULL,
    tstamp timestamp without time zone DEFAULT now() NOT NULL
);


--
-- Name: goose_db_version_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.goose_db_version ALTER COLUMN id ADD GENERATED BY DEFAULT AS IDENTITY (
    SEQUENCE NAME public.goose_db_version_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: institution_profiles; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.institution_profiles (
    id uuid NOT NULL,
    user_id uuid NOT NULL,
    slug text NOT NULL,
    name text NOT NULL,
    description text DEFAULT ''::text NOT NULL,
    contact_email text DEFAULT ''::text NOT NULL,
    contact_phone text DEFAULT ''::text NOT NULL,
    contact_website text DEFAULT ''::text NOT NULL,
    contact_location text DEFAULT ''::text NOT NULL,
    status text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT institution_profiles_status_check CHECK ((status = ANY (ARRAY['draft'::text, 'pending'::text, 'approved'::text])))
);


--
-- Name: media_assets; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.media_assets (
    id uuid NOT NULL,
    owner_user_id uuid NOT NULL,
    public_id text NOT NULL,
    secure_url text NOT NULL,
    resource_type text DEFAULT 'image'::text NOT NULL,
    width integer,
    height integer,
    bytes integer,
    folder text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT media_assets_resource_type_check CHECK ((resource_type = 'image'::text))
);


--
-- Name: oauth_accounts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.oauth_accounts (
    id uuid NOT NULL,
    user_id uuid NOT NULL,
    provider text NOT NULL,
    provider_user_id text NOT NULL,
    email text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: onboarding_applications; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.onboarding_applications (
    id uuid NOT NULL,
    applicant_id uuid NOT NULL,
    applicant_type text NOT NULL,
    display_name text NOT NULL,
    notes text DEFAULT ''::text NOT NULL,
    status text NOT NULL,
    reviewed_by uuid,
    reviewed_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    requested_handle text,
    CONSTRAINT onboarding_applications_applicant_type_check CHECK ((applicant_type = ANY (ARRAY['artist'::text, 'institution'::text]))),
    CONSTRAINT onboarding_applications_status_check CHECK ((status = ANY (ARRAY['pending'::text, 'approved'::text, 'rejected'::text])))
);


--
-- Name: page_view_daily; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.page_view_daily (
    entity_type text NOT NULL,
    entity_id uuid NOT NULL,
    day date NOT NULL,
    count bigint DEFAULT 0 NOT NULL,
    CONSTRAINT page_view_daily_count_check CHECK ((count >= 0)),
    CONSTRAINT page_view_daily_entity_type_check CHECK ((entity_type = ANY (ARRAY['artist'::text, 'post'::text, 'article'::text])))
);


--
-- Name: page_view_dedupe; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.page_view_dedupe (
    hash text NOT NULL,
    expires_at timestamp with time zone NOT NULL
);


--
-- Name: password_reset_tokens; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.password_reset_tokens (
    id uuid NOT NULL,
    user_id uuid NOT NULL,
    token_hash text NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    used_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: scrape_settings; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.scrape_settings (
    id integer NOT NULL,
    scrape_enabled boolean DEFAULT false NOT NULL,
    scrape_sources text[] DEFAULT '{}'::text[] NOT NULL,
    scrape_user_agent text DEFAULT 'mq-scraper/1.0'::text NOT NULL,
    scrape_timeout_seconds integer DEFAULT 30 NOT NULL,
    scrape_interval_seconds integer DEFAULT 21600 NOT NULL,
    telegram_enabled boolean DEFAULT false NOT NULL,
    telegram_api_id integer DEFAULT 0 NOT NULL,
    telegram_api_hash text DEFAULT ''::text NOT NULL,
    telegram_channels text[] DEFAULT '{}'::text[] NOT NULL,
    telegram_keywords text[] DEFAULT '{}'::text[] NOT NULL,
    telegram_fetch_limit integer DEFAULT 50 NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_by uuid,
    CONSTRAINT scrape_settings_id_check CHECK ((id = 1))
);


--
-- Name: user_notification_preferences; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_notification_preferences (
    user_id uuid NOT NULL,
    email_on_new_application boolean DEFAULT true NOT NULL,
    email_on_event_sync_summary boolean DEFAULT false NOT NULL,
    newsletter_enabled boolean DEFAULT false NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    id uuid NOT NULL,
    email text NOT NULL,
    role text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    password_hash text,
    display_name text,
    avatar_url text,
    email_verified_at timestamp with time zone,
    CONSTRAINT users_role_check CHECK ((role = ANY (ARRAY['public'::text, 'artist'::text, 'institution'::text, 'contributor'::text, 'admin'::text])))
);


--
-- Data for Name: art_post_media; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.art_post_media (id, art_post_id, url, mime_type, width, height, sort_order, media_asset_id) FROM stdin;
66666666-6666-6666-6666-666666666601	55555555-5555-5555-5555-555555555501	https://images.unsplash.com/photo-1579783902614-a3fb3927b6a5?w=1200&q=80	image/jpeg	0	0	0	\N
66666666-6666-6666-6666-666666666602	55555555-5555-5555-5555-555555555502	https://images.unsplash.com/photo-1541961017774-22349e4a1262?w=1200&q=80	image/jpeg	0	0	0	\N
66666666-6666-6666-6666-666666666603	55555555-5555-5555-5555-555555555503	https://images.unsplash.com/photo-1578301978693-85fa9d0ae59c?w=1200&q=80	image/jpeg	0	0	0	\N
66666666-6666-6666-6666-666666666604	55555555-5555-5555-5555-555555555504	https://images.unsplash.com/photo-1515405295579-ba7b45403062?w=1200&q=80	image/jpeg	0	0	0	\N
dddddddd-dddd-dddd-dddd-dddddddd0005	cccccccc-cccc-cccc-cccc-cccccccc0005	https://images.unsplash.com/photo-1541961017774-22349e4a1262?w=1200&q=80	image/jpeg	0	0	0	\N
dddddddd-dddd-dddd-dddd-dddddddd0006	cccccccc-cccc-cccc-cccc-cccccccc0006	https://images.unsplash.com/photo-1579783902614-a3fb3927b6a5?w=1200&q=80	image/jpeg	0	0	0	\N
dddddddd-dddd-dddd-dddd-dddddddd0007	cccccccc-cccc-cccc-cccc-cccccccc0007	https://images.unsplash.com/photo-1541961017774-22349e4a1262?w=1200&q=80	image/jpeg	0	0	0	\N
dddddddd-dddd-dddd-dddd-dddddddd0008	cccccccc-cccc-cccc-cccc-cccccccc0008	https://images.unsplash.com/photo-1541961017774-22349e4a1262?w=1200&q=80	image/jpeg	0	0	0	\N
dddddddd-dddd-dddd-dddd-dddddddd0009	cccccccc-cccc-cccc-cccc-cccccccc0009	https://images.unsplash.com/photo-1578301978693-85fa9d0ae59c?w=1200&q=80	image/jpeg	0	0	0	\N
dddddddd-dddd-dddd-dddd-dddddddd0001	cccccccc-cccc-cccc-cccc-cccccccc0001	https://images.unsplash.com/photo-1547891654-e66ed7ebb968?w=1200&q=80	image/jpeg	0	0	0	\N
dddddddd-dddd-dddd-dddd-dddddddd0002	cccccccc-cccc-cccc-cccc-cccccccc0002	https://images.unsplash.com/photo-1578301978693-85fa9d0ae59c?w=1200&q=80	image/jpeg	0	0	0	\N
dddddddd-dddd-dddd-dddd-dddddddd0003	cccccccc-cccc-cccc-cccc-cccccccc0003	https://images.unsplash.com/photo-1579783902614-a3fb3927b6a5?w=1200&q=80	image/jpeg	0	0	0	\N
dddddddd-dddd-dddd-dddd-dddddddd0004	cccccccc-cccc-cccc-cccc-cccccccc0004	https://images.unsplash.com/photo-1541961017774-22349e4a1262?w=1200&q=80	image/jpeg	0	0	0	\N
\.


--
-- Data for Name: art_posts; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.art_posts (id, artist_id, title, description, medium, status, published_at, created_at, updated_at, year, dimensions, city, style, featured_acquisition, palette, residency, exhibition) FROM stdin;
55555555-5555-5555-5555-555555555501	44444444-4444-4444-4444-444444444444	Blue Hour Market	Evening light over Merkato stalls.	oil on canvas	published	2026-06-07 11:59:42.396691+00	2026-07-07 11:59:42.396691+00	2026-07-07 11:59:42.396691+00	2024	120 × 90 cm	Addis Ababa	Contemporary	t	{#1e3a5f,#c4a574,#8b4513,#f5f0e8}	\N	\N
55555555-5555-5555-5555-555555555502	44444444-4444-4444-4444-444444444444	Coffee Ceremony	Three figures in dialogue around the jebena.	acrylic	published	2026-06-17 11:59:42.396691+00	2026-07-07 11:59:42.396691+00	2026-07-07 11:59:42.396691+00	2024	80 × 60 cm	Addis Ababa	Figurative	f	{#8b2942,#d4a574,#2c1810}	\N	\N
55555555-5555-5555-5555-555555555503	44444444-4444-4444-4444-444444444444	Entoto Mist	Hills dissolving into morning fog.	oil on linen	published	2026-06-27 11:59:42.396691+00	2026-07-07 11:59:42.396691+00	2026-07-07 11:59:42.396691+00	2023	100 × 70 cm	Addis Ababa	Landscape	f	{#4a6741,#87ceeb,#f5f0e8}	\N	\N
55555555-5555-5555-5555-555555555504	44444444-4444-4444-4444-444444444444	Thread & Ash	Textile fragments and charcoal on board.	mixed media	published	2026-07-02 11:59:42.396691+00	2026-07-07 11:59:42.396691+00	2026-07-07 11:59:42.396691+00	2024	60 × 60 cm	Addis Ababa	Mixed media	t	{#2c1810,#c4a574,#8b2942}	\N	\N
cccccccc-cccc-cccc-cccc-cccccccc0001	bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbb0001	Ascension, after the rains	Oil and gold leaf on linen.	Oil on linen	published	2026-07-09 21:55:36.688015+00	2026-07-12 21:28:38.465824+00	2026-07-12 21:55:36.688015+00	2023	180 × 150 cm	Addis Ababa	Abstract	t	{#1a2f6b,#d94e1f,#f5d76e,#f1ead8}	\N	\N
cccccccc-cccc-cccc-cccc-cccccccc0002	bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbb0002	Four Saints in Indigo	Tempera on board.	Tempera on board	published	2026-07-05 21:55:36.688015+00	2026-07-12 21:28:38.465824+00	2026-07-12 21:55:36.688015+00	2024	60 × 45 cm	Berlin	Figurative	t	{#1a2f6b,#c2410c,#f5d76e,#e8d5b0}	\N	\N
cccccccc-cccc-cccc-cccc-cccccccc0003	bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbb0003	Fidäl, dissolving	Mixed media on jute.	Mixed media on jute	published	2026-06-28 21:55:36.688015+00	2026-07-12 21:28:38.465824+00	2026-07-12 21:55:36.688015+00	2024	90 × 70 cm	Addis Ababa	Text-based	f	{#b91c1c,#f59e0b,#f5e6c0,#1c1917}	\N	\N
cccccccc-cccc-cccc-cccc-cccccccc0004	bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbb0004	Conversations in Pink	Acrylic on canvas.	Acrylic on canvas	published	2026-06-21 21:55:36.688015+00	2026-07-12 21:28:38.465824+00	2026-07-12 21:55:36.688015+00	2023	150 × 120 cm	Brooklyn	Abstract	f	{#ec4899,#0d9488,#0a0a0a,#f97316}	\N	\N
cccccccc-cccc-cccc-cccc-cccccccc0005	bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbb0003	Three Squares for Meskel	Acrylic on canvas.	Acrylic on canvas	published	2026-07-02 21:55:36.688015+00	2026-07-12 21:55:36.688015+00	2026-07-12 21:55:36.688015+00	2024	120 × 90 cm	Addis Ababa	Geometric	f	{#d62828,#fcbf49,#0a0a0a,#f1ead8}	\N	\N
cccccccc-cccc-cccc-cccc-cccccccc0006	bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbb0001	Field study, Entoto	Oil and gold leaf.	Oil and gold leaf	published	2026-06-24 21:55:36.688015+00	2026-07-12 21:55:36.688015+00	2026-07-12 21:55:36.688015+00	2022	100 × 80 cm	Addis Ababa	Landscape	f	{#2d4a2b,#c47a3d,#f4c430,#e8e1c6}	\N	\N
cccccccc-cccc-cccc-cccc-cccccccc0007	bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbb0005	Vessel study I	Stoneware vessel study.	Stoneware	published	2026-07-07 21:55:36.688015+00	2026-07-12 21:55:36.688015+00	2026-07-12 21:55:36.688015+00	2024	40 × 30 cm	Addis Ababa	Ceramic	f	{#0d9488,#f97316,#1c1917,#f5e6c0}	\N	\N
cccccccc-cccc-cccc-cccc-cccccccc0008	bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbb0005	Mercato floor tiles	Glazed ceramic installation.	Glazed ceramic	published	2026-06-30 21:55:36.688015+00	2026-07-12 21:55:36.688015+00	2026-07-12 21:55:36.688015+00	2023	Installation	Addis Ababa	Installation	f	{#d62828,#fcbf49,#0a0a0a,#f1ead8}	\N	\N
cccccccc-cccc-cccc-cccc-cccccccc0009	bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbb0005	Hawassa light	Terracotta sculpture.	Terracotta	published	2026-06-17 21:55:36.688015+00	2026-07-12 21:55:36.688015+00	2026-07-12 21:55:36.688015+00	2022	50 × 40 cm	Hawassa	Sculpture	f	{#c2410c,#f5d76e,#1a2f6b,#e8d5b0}	\N	\N
\.


--
-- Data for Name: article_revisions; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.article_revisions (id, article_id, version, editor_id, title, body, slug, category, excerpt, reading_time, difficulty, verified, status, created_at) FROM stdin;
5a8718c8-8c9c-49d9-ab5e-77a575b94bff	eeeeeeee-eeee-eeee-eeee-eeeeeeee0011	1	00000000-0000-4000-8000-000000000001	Registering your work with the EIPA	Step-by-step on filing with the Ethiopian Intellectual Property Authority — fees, forms, and what to bring. Register early to establish a clear record of your work.	eipa-registration	Legal	Step-by-step on filing with the Ethiopian Intellectual Property Authority — fees, forms, and what to bring.	8	Beginner	t	published	2026-07-18 15:33:21.288981+00
66b2d272-7cd5-4c86-9a17-5bc17ebc26a0	eeeeeeee-eeee-eeee-eeee-eeeeeeee0011	2	00000000-0000-4000-8000-000000000001	Registering your work with the EIPA	Step-by-step on filing with the Ethiopian Intellectual Property Authority — fees, forms, and what to bring. Register early to establish a clear record of your work.Step-by-step on filing with the Ethiopian Intellectual Property Authority — fees, forms, and what to bring. Register early to establish a clear record of your work.Step-by-step on filing with the Ethiopian Intellectual Property Authority — fees, forms, and what to bring. Register early to establish a clear record of your work.Step-by-step on filing with the Ethiopian Intellectual Property Authority — fees, forms, and what to bring. Register early to establish a clear record of your work.Step-by-step on filing with the Ethiopian Intellectual Property Authority — fees, forms, and what to bring. Register early to establish a clear record of your work.	eipa-registration	Legal	Step-by-step on filing with the Ethiopian Intellectual Property Authority — fees, forms, and what to bring.	8	Beginner	t	published	2026-07-18 15:33:32.272359+00
c2a3f4aa-e380-4e77-844a-590a21b825d9	eeeeeeee-eeee-eeee-eeee-eeeeeeee0011	3	00000000-0000-4000-8000-000000000001	Registering your work with the EIPA	Step-by-step on filing with the Ethiopian Intellectual Property Authority — fees, forms, and what to bring. Register early to establish a clear record of your work.	eipa-registration	Legal	Step-by-step on filing with the Ethiopian Intellectual Property Authority — fees, forms, and what to bring.	8	Beginner	t	published	2026-07-18 15:33:34.895231+00
0df893d5-85e8-487a-b9bd-8d3450404d05	eeeeeeee-eeee-eeee-eeee-eeeeeeee0011	4	00000000-0000-4000-8000-000000000001	Registering your work with the EIPA	Step-by-step on filing with the Ethiopian Intellectual Property Authority — fees, forms, and what to bring. Register early to establish a clear record of your work.	eipa-registration	Legal	Step-by-step on filing with the Ethiopian Intellectual Property Authority — fees, forms, and what to bring.	8	Beginner	t	published	2026-07-18 15:36:39.529208+00
\.


--
-- Data for Name: article_submissions; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.article_submissions (id, submitter_id, article_id, title, body, status, review_notes, reviewed_by, reviewed_at, created_at, updated_at) FROM stdin;
fe4ec12c-22cc-47c3-b6d1-9bf7e253bf24	33333333-3333-3333-3333-333333333333	ef2a5be9-c247-411c-972d-920f58504e45	Article One	What is Lorem Ipsum?\n\nLorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's standard dummy text ever since 1966, when designers at Letraset and James Mosley, the librarian at St Bride Printing Library in London, took a 1914 Cicero translation and scrambled it to make dummy text for Letraset's Body Type sheets. It has survived not only many decades, but also the leap into electronic typesetting, remaining essentially unchanged. It was popularised thanks to these sheets and more recently with desktop publishing software like Aldus PageMaker and Microsoft Word including versions of Lorem Ipsum.\nWhy do we use it?\n\nIt is a long established fact that a reader will be distracted by the readable content of a page when looking at its layout. The point of using Lorem Ipsum is that it has a more-or-less normal distribution of letters, as opposed to using 'Content here, content here', making it look like readable English. Many desktop publishing packages and web page editors now use Lorem Ipsum as their default model text, and a search for 'lorem ipsum' will uncover many web sites still in their infancy. Various versions have evolved over the years, sometimes by accident, sometimes on purpose (injected humour and the like).	approved	test 	00000000-0000-4000-8000-000000000001	2026-07-26 11:54:40.827723+00	2026-07-26 11:53:53.393328+00	2026-07-26 11:54:40.827723+00
\.


--
-- Data for Name: articles; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.articles (id, slug, title, body, status, author_id, created_at, updated_at, category, excerpt, reading_time, difficulty, verified, contributors, version) FROM stdin;
11111111-1111-1111-1111-111111111111	welcome-to-mq	Welcome to mq	A community wiki and archive for Ethiopian artists.	published	22222222-2222-2222-2222-222222222222	2026-07-06 12:43:20.652934+00	2026-07-06 12:43:20.652934+00	General		10	Beginner	f	1	1
e4c8d8ec-a8ba-42c9-967d-dd3a8e3d46c0	how-to-paint	How to paint	Start with brushes.	draft	33333333-3333-3333-3333-333333333333	2026-07-06 12:44:16.745766+00	2026-07-06 12:44:16.745766+00	General		10	Beginner	f	1	1
eeeeeeee-eeee-eeee-eeee-eeeeeeee0001	copyright-basics	Copyright Basics for Artists	Ethiopian copyright law protects original works from the moment of creation. Register with the Intellectual Property Office for additional protection.	published	33333333-3333-3333-3333-333333333333	2026-07-12 21:28:38.465824+00	2026-07-12 21:28:38.465824+00	Legal	What every Ethiopian artist should know about protecting their work.	8	Beginner	t	12	1
eeeeeeee-eeee-eeee-eeee-eeeeeeee0002	oil-painting-techniques	Oil Painting Techniques	Layering, glazing, and impasto methods used by contemporary Ethiopian painters.	published	33333333-3333-3333-3333-333333333333	2026-07-12 21:28:38.465824+00	2026-07-12 21:28:38.465824+00	Technique	Foundational oil techniques from Addis studios.	15	Intermediate	t	8	1
eeeeeeee-eeee-eeee-eeee-eeeeeeee0012	addis-pigment-sources	Where to buy pigments, linen and stretchers in Addis	A working list of Piassa, Mercato and Bole suppliers — what they stock, who imports, and price ranges. Updated by artists who buy materials weekly.	published	33333333-3333-3333-3333-333333333333	2026-07-12 21:55:36.688015+00	2026-07-12 21:55:36.688015+00	Materials	A working list of Piassa, Mercato and Bole suppliers — what they stock, who imports, and price ranges.	12	Beginner	t	23	1
eeeeeeee-eeee-eeee-eeee-eeeeeeee0013	pricing-local-vs-international	Pricing commissions: local vs. international clients	How to set a base rate in ETB and USD, account for FX volatility, and avoid common underpricing traps when working across borders.	published	33333333-3333-3333-3333-333333333333	2026-07-12 21:55:36.688015+00	2026-07-12 21:55:36.688015+00	Pricing	How to set a base rate in ETB and USD, account for FX volatility, and avoid common underpricing traps.	15	Intermediate	f	7	1
eeeeeeee-eeee-eeee-eeee-eeeeeeee0014	shipping-works-abroad	Shipping works abroad from Addis	Crating, customs paperwork, DHL vs. freight forwarders, and what collectors actually pay for when importing Ethiopian art.	published	33333333-3333-3333-3333-333333333333	2026-07-12 21:55:36.688015+00	2026-07-12 21:55:36.688015+00	Distribution	Crating, customs paperwork, DHL vs. freight forwarders, and what collectors actually pay for.	10	Beginner	f	5	1
eeeeeeee-eeee-eeee-eeee-eeeeeeee0015	fair-use-amharic	Fair use and licensing in an Ethiopian context	What Ethiopian copyright law actually says about derivative works, sampling, and reference photography.	published	33333333-3333-3333-3333-333333333333	2026-07-12 21:55:36.688015+00	2026-07-12 21:55:36.688015+00	Legal	What Ethiopian copyright law actually says about derivative works, sampling, and reference photography.	14	Advanced	t	14	1
eeeeeeee-eeee-eeee-eeee-eeeeeeee0003	gallery-contracts	Reading a gallery consignment contract	Commission splits, insurance, exclusivity windows, and the clauses Ethiopian artists keep getting burned by.	published	33333333-3333-3333-3333-333333333333	2026-07-12 21:28:38.465824+00	2026-07-12 21:55:36.688015+00	Contracts	Commission splits, insurance, exclusivity windows, and the clauses Ethiopian artists keep getting burned by.	18	Advanced	t	9	1
eeeeeeee-eeee-eeee-eeee-eeeeeeee0011	eipa-registration	Registering your work with the EIPA	Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's standard dummy text ever since 1966, when designers at Letraset and James Mosley, the librarian at St Bride Printing Library in London, took a 1914 Cicero translation and scrambled it to make dummy text for Letraset's Body Type sheets. It has survived not only many decades, but also the leap into electronic typesetting, remaining essentially unchanged. It was popularised thanks to these sheets and more recently with desktop publishing software like Aldus PageMaker and Microsoft Word including versions of Lorem Ipsum.\nWhy do we use it?\n\nIt is a long established fact that a reader will be distracted by the readable content of a page when looking at its layout. The point of using Lorem Ipsum is that it has a more-or-less normal distribution of letters, as opposed to using 'Content here, content here', making it look like readable English. Many desktop publishing packages and web page editors now use Lorem Ipsum as their default model text, and a search for 'lorem ipsum' will uncover many web sites still in their infancy. Various versions have evolved over the years, sometimes by accident, sometimes on purpose (injected humour and the like).\n	published	33333333-3333-3333-3333-333333333333	2026-07-12 21:55:36.688015+00	2026-07-18 15:36:39.532811+00	Legal	Step-by-step on filing with the Ethiopian Intellectual Property Authority — fees, forms, and what to bring.	8	Beginner	t	11	5
ef2a5be9-c247-411c-972d-920f58504e45	article-one	Article One	What is Lorem Ipsum?\n\nLorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's standard dummy text ever since 1966, when designers at Letraset and James Mosley, the librarian at St Bride Printing Library in London, took a 1914 Cicero translation and scrambled it to make dummy text for Letraset's Body Type sheets. It has survived not only many decades, but also the leap into electronic typesetting, remaining essentially unchanged. It was popularised thanks to these sheets and more recently with desktop publishing software like Aldus PageMaker and Microsoft Word including versions of Lorem Ipsum.\nWhy do we use it?\n\nIt is a long established fact that a reader will be distracted by the readable content of a page when looking at its layout. The point of using Lorem Ipsum is that it has a more-or-less normal distribution of letters, as opposed to using 'Content here, content here', making it look like readable English. Many desktop publishing packages and web page editors now use Lorem Ipsum as their default model text, and a search for 'lorem ipsum' will uncover many web sites still in their infancy. Various versions have evolved over the years, sometimes by accident, sometimes on purpose (injected humour and the like).	published	00000000-0000-4000-8000-000000000001	2026-07-26 11:54:40.821609+00	2026-07-26 11:54:40.821609+00	General		2	Beginner	f	1	1
\.


--
-- Data for Name: artist_profiles; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.artist_profiles (id, user_id, slug, display_name, bio, contact_email, contact_phone, contact_website, contact_location, social_instagram, social_twitter, social_telegram, status, created_at, updated_at, handle, born, discipline, tagline, years_active, featured, portrait_url, influences, in_residence, residency_place, open_for_commission, approved_at, portrait_media_asset_id) FROM stdin;
bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbb0004	aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaa0004	hewan-tadesse	Hewan Tadesse	Hewan's work is a quiet riot — chromatic blocks layered with woodblock prints of household objects, market scales, and her mother's hands.				Brooklyn	@hewan.tadesse			approved	2026-07-12 21:28:38.465824+00	2026-07-18 11:45:52.306393+00	hewan-tadesse	b. 1995, Dire Dawa	Painting / Print	A quiet riot of color and print.	2017 — present	f	https://images.unsplash.com/photo-1438761681033-6461ffad8d80?w=800&q=80	{}	f		f	2026-07-18 11:45:52.306393+00	\N
bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbb0001	aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaa0001	tewodros-hailu	Tewodros Hailu	Tewodros works in oil and pigment-soaked linen, building gestural fields that recall the bold mark-making of Gebre Kristos Desta. His practice circles around memory, migration, and the residual marks of city walls.				Addis Ababa	@tewodros_studio		https://t.me/tewodros_studio	approved	2026-07-12 21:28:38.465824+00	2026-07-12 21:55:36.688015+00	tewodros-hailu	b. 1987, Addis Ababa	Painting / Mixed media	Painter of residue — walls, weather, and the memory of both.	2009 — present	f	https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?w=800&q=80	{}	f	\N	f	2026-07-12 21:55:36.688015+00	\N
bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbb0002	aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaa0002	meron-alemu	Meron Alemu	Meron's canvases collide Orthodox iconography with abstract chromatic noise. She treats color as language — a system of grammar she learned from her grandmother's church murals.				Berlin / Addis Ababa	@meronalemu		https://t.me/meronalemu	approved	2026-07-12 21:28:38.465824+00	2026-07-12 21:55:36.688015+00	meron-alemu	b. 1991, Bahir Dar	Painting / Collage	Color as language.	2013 — present	f	https://images.unsplash.com/photo-1494790108377-be9c29b29330?w=800&q=80	{}	f	\N	f	2026-07-12 21:55:36.688015+00	\N
bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbb0003	aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaa0003	abel-getachew	Abel Getachew	Working from a small studio above the Mercato, Abel paints monumental color fields punctuated by fidäl — the syllabic script of Amharic — broken apart into pure form.				Addis Ababa			https://t.me/abelgetachew	approved	2026-07-12 21:28:38.465824+00	2026-07-12 21:55:36.688015+00	abel-getachew	b. 1984, Harar	Painting	Fidäl broken into pure form.	2006 — present	f	https://images.unsplash.com/photo-1472099645785-5658abf4ff4e?w=800&q=80	{}	f	\N	f	2026-07-12 21:55:36.688015+00	\N
bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbb0005	aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaa0005	demo-yordanos-kebede	Yordanos Kebede	Placeholder profile for the Mäkdäs shareable handle demo. Swap in your portrait, bio, contact links, and works when you claim your own @handle.				Addis Ababa	@makdas.demo		https://t.me/makdas_demo	approved	2026-07-12 21:55:36.688015+00	2026-07-12 21:55:36.688015+00	demo	b. 1992, Hawassa	Ceramics / Installation	This is what your link-in-bio could look like.	2018 — present	f	https://images.unsplash.com/photo-1494790108377-be9c29b29330?w=800&q=80	{}	f	\N	f	2026-07-12 21:55:36.688015+00	\N
44444444-4444-4444-4444-444444444444	33333333-3333-3333-3333-333333333333	selamawit-abebe	Selamawit Abebe	Addis Ababa–based painter exploring memory, ritual, and the textures of everyday life through oil and mixed media. Her work has been shown at the National Museum and regional galleries across Ethiopia.	selamawit@example.com		https://selamawit.example	Addis Ababa, Ethiopia	@selamawit.abebe		https://t.me/selamawitstudio	approved	2019-03-15 00:00:00+00	2026-07-16 15:15:16.188713+00	selamawit-abebe	b. 1990, Addis Ababa	Painting / Mixed media	Memory, ritual, and the textures of everyday life.	2015 — present	t	https://images.unsplash.com/photo-1544005313-94ddf0286df2?w=800&q=80	{}	f	\N	f	2026-07-16 15:15:16.188713+00	\N
\.


--
-- Data for Name: email_verification_tokens; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.email_verification_tokens (id, user_id, token_hash, expires_at, created_at) FROM stdin;
\.


--
-- Data for Name: event_locations; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.event_locations (id, name, pin_coords, created_at, updated_at) FROM stdin;
a9eff42d-f6a1-4b29-9ab6-a88d9420e54f	Alliance Ethio Francis	\N	2026-07-26 15:14:49.975279+00	2026-07-26 15:14:49.975279+00
72cabeb8-59d0-4ea7-8990-49939601d977	Hyatt Regency	\N	2026-07-26 15:14:49.983602+00	2026-07-26 15:14:49.983602+00
ae7f39e3-77cf-4755-b4de-dad69a0b96e3	timbuktoo Africa	\N	2026-07-26 15:14:49.986272+00	2026-07-26 15:14:49.986272+00
f4a9cc48-1337-4ea1-9e07-c02d1b42db0b	UNDP	\N	2026-07-26 15:14:49.989456+00	2026-07-26 15:14:49.989456+00
2aaa51e4-d608-49a6-8dfd-274a45ea475f	The Urban Center	\N	2026-07-26 15:14:49.992263+00	2026-07-26 15:14:49.992263+00
bbae7d72-f384-457f-9a2c-fe72d715cfd1	Friendship Park II (Infront of Abrehot Library Kids playing ground)	\N	2026-07-26 15:14:49.995022+00	2026-07-26 15:14:49.995022+00
0e807f38-fcc1-4ff7-a3cb-f33080ccdde6	Alliance Ethio-Française	\N	2026-07-26 15:14:49.997559+00	2026-07-26 15:14:49.997559+00
bce40310-d561-4ee1-b1e2-ac3494abdbca	To Be Announced	\N	2026-07-26 15:17:16.559738+00	2026-07-26 15:17:16.559738+00
8baee5ca-1567-4bb0-a6a0-fcb4f6a303a8	Golden tulip hotel, Bole	\N	2026-07-26 15:17:16.561802+00	2026-07-26 15:17:16.561802+00
a594d729-d530-4397-985b-c793047ccfa2	Alliance Ethio- Francis	\N	2026-07-26 15:17:16.564196+00	2026-07-26 15:17:16.564196+00
\.


--
-- Data for Name: events; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.events (id, title, description, source_url, image_url, location_id, starts_at, ends_at, scraped_at, status, review_notes, reviewed_by, reviewed_at, created_at, updated_at, slug, event_type, venue, city) FROM stdin;
ffffffff-ffff-ffff-ffff-ffffffff0011	After the Rains — Tewodros Hailu solo	Tewodros Hailu's first solo at Addis Fine Art in three years — nine new pigment-soaked linens hung unstretched, pinned like drying laundry. The work circles residue: walls, weather, and the memory of both.\n\nOpening night includes a walkthrough with the artist at 19:00. Works remain on view through August.	https://makdas.example/events/after-the-rains-tewodros-hailu	\N	\N	2026-06-27 15:00:00+00	\N	2026-07-12 21:55:36.688015+00	approved		\N	\N	2026-07-12 21:55:36.688015+00	2026-07-12 21:55:36.688015+00	after-the-rains-tewodros-hailu	Opening	Addis Fine Art	Addis Ababa
ffffffff-ffff-ffff-ffff-ffffffff0012	Tobiya Poetic Jazz Night	A monthly evening of Amharic poetry, live jazz, and open mic at Fendika. This month's theme: migration and return.\n\nDoors at 20:00. Entry 150 ETB at the door.	https://makdas.example/events/tobiya-poetic-jazz-night	\N	\N	2026-06-28 17:00:00+00	\N	2026-07-12 21:55:36.688015+00	approved		\N	\N	2026-07-12 21:55:36.688015+00	2026-07-12 21:55:36.688015+00	tobiya-poetic-jazz-night	Poetry	Fendika Cultural Center	Addis Ababa
ffffffff-ffff-ffff-ffff-ffffffff0013	Curator's Walkthrough — Skunder's Cosmos	A guided tour of the Skunder Boghossian retrospective with Alle School faculty. Focus on diasporic dream-spaces and the artist's chromatic grammar.\n\nFree with museum admission. Meet in the main hall.	https://makdas.example/events/curators-walkthrough-skunders-cosmos	\N	\N	2026-07-02 12:00:00+00	\N	2026-07-12 21:55:36.688015+00	approved		\N	\N	2026-07-12 21:55:36.688015+00	2026-07-12 21:55:36.688015+00	curators-walkthrough-skunders-cosmos	Talk	Modern Art Museum / Gebre Kristos Desta Center	Addis Ababa
ffffffff-ffff-ffff-ffff-ffffffff0014	Design Pop-up: Habesha Futures	A one-weekend pop-up at Zoma Museum featuring textile designers, furniture makers, and illustrators reimagining Ethiopian craft for contemporary interiors.\n\nSaturday and Sunday, 10:00–18:00.	https://makdas.example/events/design-pop-up-habesha-futures	\N	\N	2026-07-05 07:00:00+00	\N	2026-07-12 21:55:36.688015+00	approved		\N	\N	2026-07-12 21:55:36.688015+00	2026-07-12 21:55:36.688015+00	design-pop-up-habesha-futures	Pop-up	Zoma Museum	Addis Ababa
ffffffff-ffff-ffff-ffff-ffffffff0015	Yenegew Sew — Theatre revival	A revival of the classic Ethiopian play at the National Theatre — directed by a new generation of Alle School graduates. Subtitles in English for select performances.\n\nEvening shows Thursday through Sunday.	https://makdas.example/events/yenegew-sew-theatre-revival	\N	\N	2026-07-09 16:30:00+00	\N	2026-07-12 21:55:36.688015+00	approved		\N	\N	2026-07-12 21:55:36.688015+00	2026-07-12 21:55:36.688015+00	yenegew-sew-theatre-revival	Theatre	National Theatre	Addis Ababa
ffffffff-ffff-ffff-ffff-ffffffff0016	Goethe Film Series: Diaspora Editions	Monthly screening of contemporary African and diaspora cinema, followed by a moderated discussion. This edition features work from Ethiopian filmmakers in Berlin and London.\n\nSeating is limited — arrive early.	https://makdas.example/events/goethe-film-series-diaspora-editions	\N	\N	2026-07-11 15:30:00+00	\N	2026-07-12 21:55:36.688015+00	approved		\N	\N	2026-07-12 21:55:36.688015+00	2026-07-12 21:55:36.688015+00	goethe-film-series-diaspora-editions	Screening	Goethe-Institut	Addis Ababa
ffffffff-ffff-ffff-ffff-ffffffff0017	Bahir Dar Lake Sessions — Open studios	Open studios along Lake Tana with painters, weavers, and ceramicists working in a shared boathouse space. A chance to meet artists outside the Addis circuit.\n\nFree entry. Works available for direct purchase from studios.	https://makdas.example/events/bahir-dar-lake-sessions-open-studios	\N	\N	2026-07-13 08:00:00+00	\N	2026-07-12 21:55:36.688015+00	approved		\N	\N	2026-07-12 21:55:36.688015+00	2026-07-12 21:55:36.688015+00	bahir-dar-lake-sessions-open-studios	Pop-up	Lake Tana Boathouse	Bahir Dar
8a8abc03-5a37-4961-9df6-72530b86e638	QIBIBILOSH	QIBIBILOSH\n📅 July31st , 2026\n📍 Alliance Ethio Francis\n🙏 Netsanet\n\nQibibilosh Intergenerational, Interdeciplinary PanAfrican Platform invites you for  a short film by a visual artist Dr Hassan Aliyu (DProf), FRSA is a British-born Nigerian artist, researcher and President of the Nigeria Art Society UK (NASUK). The work has been screened at *Platform Projects, Athens (2023/24)*, *Casa de África during the 15th Havana Biennial (2024)*, and in Addis Ababa (2026). His work has been exhibited internationally, including at the 1-54 Contemporary African Art Fair, Iwalewahaus (Germany) and Rele (London). In 2026, he presented his solo exhibition Epic Journeys, simultaneously at *ArtSpace5-7*, Cambridge and Wolfson College, University of Cambridge, UK\n\n@eventsethiopia\nhttps://linktr.ee/eventsethiopia	https://t.me/EventsEthiopia/10587	\N	a9eff42d-f6a1-4b29-9ab6-a88d9420e54f	2026-07-25 18:53:34+00	\N	2026-07-26 15:17:15.377014+00	pending		\N	\N	2026-07-26 15:14:49.980098+00	2026-07-26 15:17:16.548153+00	qibibilosh	Opening	Alliance Ethio Francis	Addis Ababa
c621a49a-00b4-42c4-8d72-866f497cb281	QIBIBILOSH	QIBIBILOSH\n📅 Friday, July 24, 2026.\n📍 Alliance Ethio-Française \n🙏 Netsanet\n\nYou are warmly welcome to join us for our upcoming documentary screeing entitled: "As I Remember It: A Portrait of Dorothy West", by Prof Salem Mekuria, screeing this Friday 24th of July at Alliance Ethio-Française. \n\nDoor opens: 5:30 pm\nScreening starts: 6pm\n\nAll are warmly welcomed.\n\n@eventsethiopia\nhttps://linktr.ee/eventsethiopia	https://t.me/EventsEthiopia/10575	\N	0e807f38-fcc1-4ff7-a3cb-f33080ccdde6	2026-07-24 00:00:00+00	\N	2026-07-26 15:17:15.377691+00	pending		\N	\N	2026-07-26 15:17:16.557396+00	2026-07-26 15:17:16.557396+00	qibibilosh-10575	Opening	Alliance Ethio-Française	
dee93e70-dc2d-48e4-9774-3ee29fd15803	East African Comic Competition 2026	East African Comic Competition 2026\n📅 Deadline: July 31, 2026\n📍 Alliance Ethio-Française\n\nTheme: Inhabiting the World.\n\nEligibility: Open to residents of Ethiopia, Burundi, Djibouti, Kenya, Rwanda, Tanzania, and Uganda aged between 18 and 35 yrs.\n\nApply: https://allianceaddis.org/east-african-comic-competition-2026/\n\n@eventsethiopia\nhttps://linktr.ee/eventsethiopia	https://t.me/EventsEthiopia/10570	\N	0e807f38-fcc1-4ff7-a3cb-f33080ccdde6	2026-07-31 00:00:00+00	\N	2026-07-26 15:17:15.37774+00	pending		\N	\N	2026-07-26 15:17:16.558744+00	2026-07-26 15:17:16.558744+00	east-african-comic-competition-2026-10570	Opening	Alliance Ethio-Française	
a768c205-d03d-45ab-9b7f-471b35141a2c	Venture Meda Hackathon	Venture Meda Hackathon\n📅 August 21, 2026\n📍 To Be Announced\n🙏 @bethel_t23\n\nReady to solve real-world business challenges with innovative tech solutions?\nJoin the Venture Meda Hackathon 2026 and collaborate with fellow innovators to reimagine the future of e-commerce. Develop impactful solutions, gain hands-on experience, and turn your ideas into real ventures.\n\nWhy participate?\n🏆Win seed funding to accelerate your startup journey.\n🤝 Gain invaluable mentorship from industry leaders and successful entrepreneurs.\n🎓 Graduating (final-year) students are highly encouraged to apply, though all eligible students are welcome to participate.\n\nApplication Deadline: July 31, 2026\nEvent Date 📅:  August 21–22, 2026\nRegister here : https://forms.gle/aPizDkZagKboyFKC8\n\n#VentureMedaHackathon #ECommerce #StudentInnovation #Entrepreneurship #FutureBuilders\n\n@eventsethiopia\nhttps://linktr.ee/eventsethiopia	https://t.me/EventsEthiopia/10569	\N	bce40310-d561-4ee1-b1e2-ac3494abdbca	2026-08-21 00:00:00+00	\N	2026-07-26 15:17:15.377833+00	pending		\N	\N	2026-07-26 15:17:16.560581+00	2026-07-26 15:17:16.560581+00	venture-meda-hackathon-10569	Opening	To Be Announced	
9470f6ef-6a8d-4ad9-8ece-48df7a4b4b7b	She glows beauty and wellness expo 3	She glows beauty and wellness expo 3\n📅 July 25-26, 2026 at 10 am\n📍 Golden tulip hotel, Bole\n🙏 @Block_events2025\n\n✨ SHE Glows Beauty & Wellness Expo 2026 ✨\nCelebrate beauty, wellness, and women empowerment at one of Addis Ababa's most exciting beauty events! Discover top beauty and wellness brands, watch the Makeup Artist Competition, learn from industry experts during panel discussions, enjoy interactive activities, and stand a chance to win exciting beauty product giveaways.\n📅 Date: July 25–26, 2026\n📍 Venue: Golden Tulip Addis Ababa\n🕙 Time: 10:00 AM – 7:00 PM\n🎟️ Entrance Fee: FREE\nWhether you're a beauty enthusiast, makeup artist, entrepreneur, or simply looking for a fun weekend experience, there's something for everyone.\n📞 For inquiries & reservations:\n• Phone: +251974759595\n              : +251929187132\nBecause every woman deserves to glow. ✨\n\n@eventsethiopia\nhttps://linktr.ee/eventsethiopia	https://t.me/EventsEthiopia/10567	\N	8baee5ca-1567-4bb0-a6a0-fcb4f6a303a8	2026-07-25 00:00:00+00	\N	2026-07-26 15:17:15.377935+00	pending		\N	\N	2026-07-26 15:17:16.56276+00	2026-07-26 15:17:16.56276+00	she-glows-beauty-and-wellness-expo-3-10567	Opening	Golden tulip hotel, Bole	Addis Ababa
dc13ab7c-13b6-4d12-923f-7c6f0fdfaa57	📅 [Jul 26] Upcoming Events. Click to jump 👇	📅 [Jul 26] Upcoming Events. Click to jump 👇\n\nToday\n- Shop Local Ethiopia Shop & Play Bazaar\n- She glows beauty and wellness expo 3\n\n- ትምሮ (Timiro) - የማትሪክ መዘጋጃ\n- Africa AI Lab\n- Comic Competition \n\n@eventsethiopia	https://t.me/EventsEthiopia/10588	\N	\N	2026-07-26 00:00:00+00	\N	2026-07-26 15:17:15.376643+00	pending		\N	\N	2026-07-26 15:14:49.967525+00	2026-07-26 15:17:16.538665+00	jul-26-upcoming-events-click-to-jump	Opening		
9ab245f1-f384-4dd2-aa1b-c45f0ba6ea6d	Opening of "Echoes of Piassa"	Opening of "Echoes of Piassa"\n📅 July 29, 2026\n📍 Hyatt Regency\n\nSolo digital art exhibition by Dawit Kifle.\n\n@eventsethiopia\nhttps://linktr.ee/eventsethiopia	https://t.me/EventsEthiopia/10585	\N	72cabeb8-59d0-4ea7-8990-49939601d977	2026-07-29 00:00:00+00	\N	2026-07-26 15:17:15.377076+00	pending		\N	\N	2026-07-26 15:14:49.984786+00	2026-07-26 15:17:16.550631+00	opening-of-echoes-of-piassa	Opening	Hyatt Regency	
9d5cddca-0489-4ddb-b5ce-5b45ce45090f	AI for Public Good Hackathon	AI for Public Good Hackathon\n📅 Application deadline: July 30, 2026\n📍 timbuktoo Africa\n\nFor:\n- Students\n- Researchers\n- Professionals\n- Innovators\n...passionate about using AI for social impact?\n\n#AIforPublicGoodHackathon2026    \n\n✅ Build impactful AI solutions\n✅ Collaborate with experts and peers\n✅ Contribute to solving public good challenges\n✅ Expand your network and skills\n\n@eventsethiopia\nhttps://linktr.ee/eventsethiopia	https://t.me/EventsEthiopia/10584	\N	ae7f39e3-77cf-4755-b4de-dad69a0b96e3	2026-07-30 00:00:00+00	\N	2026-07-26 15:17:15.377173+00	pending		\N	\N	2026-07-26 15:14:49.987598+00	2026-07-26 15:17:16.551729+00	ai-for-public-good-hackathon	Opening	timbuktoo Africa	
b4a13b82-1628-4b14-9b2a-44d2664d41c4	Call for Applications: African Youth Co-Creators Council	Call for Applications: African Youth Co-Creators Council\n📅 Deadline: July 31, 2026\n📍 UNDP\n\nAfrican Youth Co-Creators Council is looking for 11 young leaders to help shape youth-centred solutions across the continent.\n\nAre you 18–34?\n- An African national living on the continent or in the diaspora?\n- Already driving change in your community?\n- Passionate about governance, peacebuilding, climate action, innovation, AI, mental health, gender or inclusive development?\n\n \nApply: https://www.undp.org/africa/blog/call-applications-african-youth-co-creators-council\n\n@eventsethiopia\nhttps://linktr.ee/eventsethiopia	https://t.me/EventsEthiopia/10581	\N	f4a9cc48-1337-4ea1-9e07-c02d1b42db0b	2026-07-31 00:00:00+00	\N	2026-07-26 15:17:15.377314+00	pending		\N	\N	2026-07-26 15:14:49.990576+00	2026-07-26 15:17:16.552985+00	call-for-applications-african-youth-co-creators-council	Opening	UNDP	
22a86117-80a0-4b76-816d-5c21afc167b5	Informed Choices 2.0	Informed Choices 2.0\n📅 Saturday, July 25, 2026\n9:00 AM - 12:00 PM\n2:00 PM - 5:00 PM\n📍 The Urban Center\n🙏 @Miheret_Taye\n\n🎓 Informed Choices 2.0\n\nThis half-day mentorship and career guidance program is designed for Grade 12 students who sat for the ESLCE this year and their parents/guardians.\n\n🏛 Meet practicing professionals\n💬 Ask your questions\n🎯 Explore career paths\n📚 Make an informed choice about your future\n\n📅 Saturday, July 25, 2026\n\nMorning Session\n🕘 9:00 AM – 12:00 PM\n(Registration opens at 8:30 AM)\n\nAfternoon Session\n🕑 2:00 PM – 5:00 PM\n(Registration opens at 1:30 PM)\n\n📍 The Urban Center\n\n⚠️ Attendance is by registration only. Spaces are limited.\n\n🔗 Register here: https://forms.office.com/r/15qJzg6SWC\n\n@eventsethiopia\nhttps://linktr.ee/eventsethiopia	https://t.me/EventsEthiopia/10578	\N	2aaa51e4-d608-49a6-8dfd-274a45ea475f	2026-07-25 00:00:00+00	\N	2026-07-26 15:17:15.377552+00	pending		\N	\N	2026-07-26 15:14:49.993444+00	2026-07-26 15:17:16.554279+00	informed-choices-2-0	Opening	The Urban Center	
392dfa24-a304-4ccd-8f97-0a98476f27e9	Shop Local Ethiopia Shop & Play Bazaar	Shop Local Ethiopia Shop & Play Bazaar\n📅 July 25 & 26 2026\n📍 Friendship Park II (Infront of Abrehot Library Kids playing ground)\n🙏 @dae2011\n\n🎪 Shop & Play Bazaar is here! 🎈\nTwo days of local shopping and non-stop kids' fun at Friendship Park Phase Two (right in front of Abrhot Library).\n🗓️ July 25–26 | Saturday & Sunday\n🕙 10:00 AM – 6:30 PM\n📍 Friendship Park Phase Two, opp. Abrhot Library\n💸 Free Entry\nBring the whole family — shop from local Ethiopian vendors while the kids play. See you there! 🇪🇹✨\n#ShopLocalEthiopia #MadeInEthiopia #AddisAbaba #FamilyFun #ShopAndPlayBazaar #FriendshipPark\n\n@eventsethiopia\nhttps://linktr.ee/eventsethiopia	https://t.me/EventsEthiopia/10577	\N	bbae7d72-f384-457f-9a2c-fe72d715cfd1	2026-07-25 00:00:00+00	\N	2026-07-26 15:17:15.377639+00	pending		\N	\N	2026-07-26 15:14:49.995937+00	2026-07-26 15:17:16.555662+00	shop-local-ethiopia-shop-play-bazaar	Opening	Friendship Park II (Infront of Abrehot Library Kids playing ground)	
5425649c-c8b3-4b98-8876-d959a8357d45	QIBIBILOSH	QIBIBILOSH\n📅 Friday July 17, 2026 6:00PM\n📍 Alliance Ethio- Francis\n🙏 Netsanet\n\nDear Qibibilosh family and Friends  you are warmly welcome  for a Screeing of SIDET: a documentary film by our very own Prof Salem Mekuria  at 6:00 PM -7:00 screeing  followed by  a post screeing discussion led by  Maya Misikir.don’t want to miss this !!! On upcoming Friday  17th of July 2026, Door opens5:30 and screeing starts 6:00.  at Alliance Ethio - Francis.\n\n@eventsethiopia\nhttps://linktr.ee/eventsethiopia	https://t.me/EventsEthiopia/10566	\N	a594d729-d530-4397-985b-c793047ccfa2	2026-07-17 00:00:00+00	\N	2026-07-26 15:17:15.377994+00	pending		\N	\N	2026-07-26 15:17:16.564993+00	2026-07-26 15:17:16.564993+00	qibibilosh-10566	Opening	Alliance Ethio- Francis	
80bf744b-03bd-4818-90f2-66984e52cfeb	Community Smoke Opening	Submitted via redesign smoke test	public://events/80bf744b-03bd-4818-90f2-66984e52cfeb	\N	\N	2026-08-01 18:00:00+00	\N	2026-07-26 15:26:55.495723+00	pending		\N	\N	2026-07-26 15:26:55.495723+00	2026-07-26 15:26:55.495744+00	community-smoke-opening-80bf744b-03bd-4818-90f2-66984e52cfeb	Opening	Test Gallery	Addis Ababa
\.


--
-- Data for Name: goose_db_version; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.goose_db_version (id, version_id, is_applied, tstamp) FROM stdin;
1	0	t	2026-07-06 12:43:20.628765
2	1	t	2026-07-06 12:43:20.640866
3	2	t	2026-07-06 12:43:20.652934
4	3	t	2026-07-07 08:52:37.680704
5	4	t	2026-07-07 08:52:37.705261
6	5	t	2026-07-07 08:52:37.72062
7	6	t	2026-07-07 08:52:37.740695
8	7	t	2026-07-07 08:52:37.748665
9	8	t	2026-07-07 08:52:37.762062
10	9	t	2026-07-07 08:52:37.773763
11	10	t	2026-07-07 11:59:42.396691
12	11	t	2026-07-12 21:28:38.404087
13	12	t	2026-07-12 21:28:38.447577
14	13	t	2026-07-12 21:28:38.465824
15	14	t	2026-07-12 21:55:36.688015
16	15	t	2026-07-16 12:38:20.346749
17	16	t	2026-07-16 15:14:18.629284
18	17	t	2026-07-18 08:57:30.943672
19	18	t	2026-07-18 08:57:30.952469
20	19	t	2026-07-18 11:31:34.552359
21	20	t	2026-07-18 13:16:29.597656
22	21	t	2026-07-18 15:30:51.210071
23	22	t	2026-07-18 15:51:44.603289
24	23	t	2026-07-26 11:42:13.780396
25	24	t	2026-07-26 11:42:13.816723
26	25	t	2026-07-26 11:42:13.821866
27	26	t	2026-07-26 11:42:13.828969
28	27	t	2026-07-26 11:42:13.842564
29	28	t	2026-07-26 11:42:13.85702
30	29	t	2026-07-26 11:42:13.869374
\.


--
-- Data for Name: institution_profiles; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.institution_profiles (id, user_id, slug, name, description, contact_email, contact_phone, contact_website, contact_location, status, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: media_assets; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.media_assets (id, owner_user_id, public_id, secure_url, resource_type, width, height, bytes, folder, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: oauth_accounts; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.oauth_accounts (id, user_id, provider, provider_user_id, email, created_at) FROM stdin;
\.


--
-- Data for Name: onboarding_applications; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.onboarding_applications (id, applicant_id, applicant_type, display_name, notes, status, reviewed_by, reviewed_at, created_at, updated_at, requested_handle) FROM stdin;
\.


--
-- Data for Name: page_view_daily; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.page_view_daily (entity_type, entity_id, day, count) FROM stdin;
article	ef2a5be9-c247-411c-972d-920f58504e45	2026-07-26	1
article	eeeeeeee-eeee-eeee-eeee-eeeeeeee0001	2026-07-26	1
article	eeeeeeee-eeee-eeee-eeee-eeeeeeee0013	2026-07-26	1
\.


--
-- Data for Name: page_view_dedupe; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.page_view_dedupe (hash, expires_at) FROM stdin;
f1b3c892db6533cc14bcf3b23cceded51f6df3f7262584c75baf7bcffbcd6a5b	2026-07-28 00:00:00+00
af5e876ffe908908efbd51d274ff2a0e5bac6e61a6a2b9c0a3304a5d68b66fc4	2026-07-28 00:00:00+00
4c3995de4eb15db3953bacfa85c95fae76b4458dce92ce8a045cc54f557c7c22	2026-07-28 00:00:00+00
\.


--
-- Data for Name: password_reset_tokens; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.password_reset_tokens (id, user_id, token_hash, expires_at, used_at, created_at) FROM stdin;
\.


--
-- Data for Name: scrape_settings; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.scrape_settings (id, scrape_enabled, scrape_sources, scrape_user_agent, scrape_timeout_seconds, scrape_interval_seconds, telegram_enabled, telegram_api_id, telegram_api_hash, telegram_channels, telegram_keywords, telegram_fetch_limit, updated_at, updated_by) FROM stdin;
1	f	{}	mq-scraper/1.0	30	21600	t	20930665	c56bed7348a1d60e1a09e448fe03411a	{EventsEthiopia,addisartevents,ethioculture}	{exhibition,opening,event,art}	50	2026-07-18 15:51:47.313949+00	\N
\.


--
-- Data for Name: user_notification_preferences; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.user_notification_preferences (user_id, email_on_new_application, email_on_event_sync_summary, newsletter_enabled, updated_at) FROM stdin;
00000000-0000-4000-8000-000000000001	t	f	f	2026-07-18 15:52:27.463527+00
\.


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.users (id, email, role, created_at, updated_at, password_hash, display_name, avatar_url, email_verified_at) FROM stdin;
aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaa0001	tewodros@example.com	artist	2026-07-12 21:28:38.465824+00	2026-07-12 21:28:38.465824+00	\N	\N	\N	\N
aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaa0002	meron@example.com	artist	2026-07-12 21:28:38.465824+00	2026-07-12 21:28:38.465824+00	\N	\N	\N	\N
aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaa0003	abel@example.com	artist	2026-07-12 21:28:38.465824+00	2026-07-12 21:28:38.465824+00	\N	\N	\N	\N
aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaa0004	hewan@example.com	artist	2026-07-12 21:28:38.465824+00	2026-07-12 21:28:38.465824+00	\N	\N	\N	\N
aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaa0005	demo@makdas.example	artist	2026-07-12 21:55:36.688015+00	2026-07-12 21:55:36.688015+00	\N	\N	\N	\N
33333333-3333-3333-3333-333333333333	selamawit@example.com	artist	2019-03-15 00:00:00+00	2026-07-16 15:15:16.169081+00	$2a$10$BRfeZ0LINJ06vDtvt6a8Cu9rJFS1DThf7YAe5xNFqDY4AbbJcCV42	\N	\N	\N
00000000-0000-4000-8000-000000000001	admin@mq.local	admin	2026-07-07 08:52:37.773763+00	2026-07-18 11:31:34.552359+00	$2a$10$Tv677q70S11wHwlos.tmMe1.75sqIF2bp7pe2T1Bb7baLf79CxRT.	\N	\N	\N
\.


--
-- Name: goose_db_version_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.goose_db_version_id_seq', 30, true);


--
-- Name: art_post_media art_post_media_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.art_post_media
    ADD CONSTRAINT art_post_media_pkey PRIMARY KEY (id);


--
-- Name: art_posts art_posts_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.art_posts
    ADD CONSTRAINT art_posts_pkey PRIMARY KEY (id);


--
-- Name: article_revisions article_revisions_article_version_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.article_revisions
    ADD CONSTRAINT article_revisions_article_version_unique UNIQUE (article_id, version);


--
-- Name: article_revisions article_revisions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.article_revisions
    ADD CONSTRAINT article_revisions_pkey PRIMARY KEY (id);


--
-- Name: article_submissions article_submissions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.article_submissions
    ADD CONSTRAINT article_submissions_pkey PRIMARY KEY (id);


--
-- Name: articles articles_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.articles
    ADD CONSTRAINT articles_pkey PRIMARY KEY (id);


--
-- Name: articles articles_slug_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.articles
    ADD CONSTRAINT articles_slug_key UNIQUE (slug);


--
-- Name: artist_profiles artist_profiles_handle_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.artist_profiles
    ADD CONSTRAINT artist_profiles_handle_key UNIQUE (handle);


--
-- Name: artist_profiles artist_profiles_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.artist_profiles
    ADD CONSTRAINT artist_profiles_pkey PRIMARY KEY (id);


--
-- Name: artist_profiles artist_profiles_slug_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.artist_profiles
    ADD CONSTRAINT artist_profiles_slug_key UNIQUE (slug);


--
-- Name: email_verification_tokens email_verification_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.email_verification_tokens
    ADD CONSTRAINT email_verification_tokens_pkey PRIMARY KEY (id);


--
-- Name: email_verification_tokens email_verification_tokens_token_hash_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.email_verification_tokens
    ADD CONSTRAINT email_verification_tokens_token_hash_key UNIQUE (token_hash);


--
-- Name: event_locations event_locations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.event_locations
    ADD CONSTRAINT event_locations_pkey PRIMARY KEY (id);


--
-- Name: events events_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.events
    ADD CONSTRAINT events_pkey PRIMARY KEY (id);


--
-- Name: events events_slug_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.events
    ADD CONSTRAINT events_slug_key UNIQUE (slug);


--
-- Name: events events_source_url_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.events
    ADD CONSTRAINT events_source_url_key UNIQUE (source_url);


--
-- Name: goose_db_version goose_db_version_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.goose_db_version
    ADD CONSTRAINT goose_db_version_pkey PRIMARY KEY (id);


--
-- Name: institution_profiles institution_profiles_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.institution_profiles
    ADD CONSTRAINT institution_profiles_pkey PRIMARY KEY (id);


--
-- Name: institution_profiles institution_profiles_slug_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.institution_profiles
    ADD CONSTRAINT institution_profiles_slug_key UNIQUE (slug);


--
-- Name: media_assets media_assets_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.media_assets
    ADD CONSTRAINT media_assets_pkey PRIMARY KEY (id);


--
-- Name: media_assets media_assets_public_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.media_assets
    ADD CONSTRAINT media_assets_public_id_key UNIQUE (public_id);


--
-- Name: oauth_accounts oauth_accounts_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.oauth_accounts
    ADD CONSTRAINT oauth_accounts_pkey PRIMARY KEY (id);


--
-- Name: oauth_accounts oauth_accounts_provider_provider_user_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.oauth_accounts
    ADD CONSTRAINT oauth_accounts_provider_provider_user_id_key UNIQUE (provider, provider_user_id);


--
-- Name: onboarding_applications onboarding_applications_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.onboarding_applications
    ADD CONSTRAINT onboarding_applications_pkey PRIMARY KEY (id);


--
-- Name: page_view_daily page_view_daily_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.page_view_daily
    ADD CONSTRAINT page_view_daily_pkey PRIMARY KEY (entity_type, entity_id, day);


--
-- Name: page_view_dedupe page_view_dedupe_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.page_view_dedupe
    ADD CONSTRAINT page_view_dedupe_pkey PRIMARY KEY (hash);


--
-- Name: password_reset_tokens password_reset_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.password_reset_tokens
    ADD CONSTRAINT password_reset_tokens_pkey PRIMARY KEY (id);


--
-- Name: password_reset_tokens password_reset_tokens_token_hash_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.password_reset_tokens
    ADD CONSTRAINT password_reset_tokens_token_hash_key UNIQUE (token_hash);


--
-- Name: scrape_settings scrape_settings_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.scrape_settings
    ADD CONSTRAINT scrape_settings_pkey PRIMARY KEY (id);


--
-- Name: user_notification_preferences user_notification_preferences_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_notification_preferences
    ADD CONSTRAINT user_notification_preferences_pkey PRIMARY KEY (user_id);


--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: idx_art_post_media_post_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_art_post_media_post_id ON public.art_post_media USING btree (art_post_id);


--
-- Name: idx_art_posts_artist_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_art_posts_artist_id ON public.art_posts USING btree (artist_id);


--
-- Name: idx_art_posts_search; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_art_posts_search ON public.art_posts USING gin (to_tsvector('simple'::regconfig, ((((((((COALESCE(title, ''::text) || ' '::text) || COALESCE(description, ''::text)) || ' '::text) || COALESCE(medium, ''::text)) || ' '::text) || COALESCE(city, ''::text)) || ' '::text) || COALESCE(style, ''::text))));


--
-- Name: idx_art_posts_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_art_posts_status ON public.art_posts USING btree (status);


--
-- Name: idx_article_revisions_article_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_article_revisions_article_id ON public.article_revisions USING btree (article_id, version DESC);


--
-- Name: idx_article_submissions_pending; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_article_submissions_pending ON public.article_submissions USING btree (created_at) WHERE (status = 'pending'::text);


--
-- Name: idx_article_submissions_submitter; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_article_submissions_submitter ON public.article_submissions USING btree (submitter_id, created_at DESC);


--
-- Name: idx_articles_author_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_articles_author_id ON public.articles USING btree (author_id);


--
-- Name: idx_articles_search_vector; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_articles_search_vector ON public.articles USING gin (search_vector);


--
-- Name: idx_articles_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_articles_status ON public.articles USING btree (status);


--
-- Name: idx_artist_profiles_search; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_artist_profiles_search ON public.artist_profiles USING gin (to_tsvector('simple'::regconfig, ((((((COALESCE(display_name, ''::text) || ' '::text) || COALESCE(bio, ''::text)) || ' '::text) || COALESCE(handle, ''::text)) || ' '::text) || COALESCE(discipline, ''::text))));


--
-- Name: idx_artist_profiles_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_artist_profiles_status ON public.artist_profiles USING btree (status);


--
-- Name: idx_artist_profiles_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_artist_profiles_user_id ON public.artist_profiles USING btree (user_id);


--
-- Name: idx_email_verification_tokens_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_email_verification_tokens_user_id ON public.email_verification_tokens USING btree (user_id);


--
-- Name: idx_event_locations_name; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_event_locations_name ON public.event_locations USING btree (lower(TRIM(BOTH FROM name)));


--
-- Name: idx_events_location_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_events_location_id ON public.events USING btree (location_id);


--
-- Name: idx_events_search_vector; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_events_search_vector ON public.events USING gin (search_vector);


--
-- Name: idx_events_starts_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_events_starts_at ON public.events USING btree (starts_at);


--
-- Name: idx_events_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_events_status ON public.events USING btree (status);


--
-- Name: idx_institution_profiles_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_institution_profiles_status ON public.institution_profiles USING btree (status);


--
-- Name: idx_institution_profiles_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_institution_profiles_user_id ON public.institution_profiles USING btree (user_id);


--
-- Name: idx_media_assets_owner; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_media_assets_owner ON public.media_assets USING btree (owner_user_id);


--
-- Name: idx_oauth_accounts_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_oauth_accounts_user_id ON public.oauth_accounts USING btree (user_id);


--
-- Name: idx_onboarding_applications_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_onboarding_applications_status ON public.onboarding_applications USING btree (status);


--
-- Name: idx_page_view_dedupe_expiry; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_page_view_dedupe_expiry ON public.page_view_dedupe USING btree (expires_at);


--
-- Name: idx_password_reset_tokens_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_password_reset_tokens_user_id ON public.password_reset_tokens USING btree (user_id);


--
-- Name: uq_onboarding_active_requested_handle; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uq_onboarding_active_requested_handle ON public.onboarding_applications USING btree (lower(requested_handle)) WHERE ((requested_handle IS NOT NULL) AND (status = ANY (ARRAY['pending'::text, 'approved'::text])));


--
-- Name: art_post_media art_post_media_art_post_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.art_post_media
    ADD CONSTRAINT art_post_media_art_post_id_fkey FOREIGN KEY (art_post_id) REFERENCES public.art_posts(id) ON DELETE CASCADE;


--
-- Name: art_post_media art_post_media_media_asset_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.art_post_media
    ADD CONSTRAINT art_post_media_media_asset_id_fkey FOREIGN KEY (media_asset_id) REFERENCES public.media_assets(id) ON DELETE SET NULL;


--
-- Name: art_posts art_posts_artist_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.art_posts
    ADD CONSTRAINT art_posts_artist_id_fkey FOREIGN KEY (artist_id) REFERENCES public.artist_profiles(id);


--
-- Name: article_revisions article_revisions_article_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.article_revisions
    ADD CONSTRAINT article_revisions_article_id_fkey FOREIGN KEY (article_id) REFERENCES public.articles(id) ON DELETE CASCADE;


--
-- Name: article_submissions article_submissions_article_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.article_submissions
    ADD CONSTRAINT article_submissions_article_id_fkey FOREIGN KEY (article_id) REFERENCES public.articles(id) ON DELETE SET NULL;


--
-- Name: article_submissions article_submissions_reviewed_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.article_submissions
    ADD CONSTRAINT article_submissions_reviewed_by_fkey FOREIGN KEY (reviewed_by) REFERENCES public.users(id);


--
-- Name: article_submissions article_submissions_submitter_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.article_submissions
    ADD CONSTRAINT article_submissions_submitter_id_fkey FOREIGN KEY (submitter_id) REFERENCES public.users(id);


--
-- Name: artist_profiles artist_profiles_portrait_media_asset_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.artist_profiles
    ADD CONSTRAINT artist_profiles_portrait_media_asset_id_fkey FOREIGN KEY (portrait_media_asset_id) REFERENCES public.media_assets(id) ON DELETE SET NULL;


--
-- Name: artist_profiles artist_profiles_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.artist_profiles
    ADD CONSTRAINT artist_profiles_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: email_verification_tokens email_verification_tokens_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.email_verification_tokens
    ADD CONSTRAINT email_verification_tokens_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: events events_location_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.events
    ADD CONSTRAINT events_location_id_fkey FOREIGN KEY (location_id) REFERENCES public.event_locations(id) ON DELETE SET NULL;


--
-- Name: institution_profiles institution_profiles_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.institution_profiles
    ADD CONSTRAINT institution_profiles_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: media_assets media_assets_owner_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.media_assets
    ADD CONSTRAINT media_assets_owner_user_id_fkey FOREIGN KEY (owner_user_id) REFERENCES public.users(id);


--
-- Name: oauth_accounts oauth_accounts_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.oauth_accounts
    ADD CONSTRAINT oauth_accounts_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: password_reset_tokens password_reset_tokens_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.password_reset_tokens
    ADD CONSTRAINT password_reset_tokens_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: scrape_settings scrape_settings_updated_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.scrape_settings
    ADD CONSTRAINT scrape_settings_updated_by_fkey FOREIGN KEY (updated_by) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: user_notification_preferences user_notification_preferences_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_notification_preferences
    ADD CONSTRAINT user_notification_preferences_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

\unrestrict azZsFcvp4tMvust95Hhi3lHsThfqdUjBDDdxqXlLGXLqquTFrUKEY4cKw6HQCAq

