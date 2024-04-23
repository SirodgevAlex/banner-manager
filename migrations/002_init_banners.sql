CREATE TABLE IF NOT EXISTS public.banners (
    id SERIAL PRIMARY KEY,
    feature_id INT,
    tag_id INT,
    title TEXT,
    text TEXT,
    url TEXT,
    is_active BOOLEAN,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX public.idx_unique_feature_tag ON banners (feature_id, tag_id);
