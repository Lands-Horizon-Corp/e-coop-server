-- Enable the pgcrypto extension to use gen_random_uuid
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Enable the vector extension for vector operations
CREATE EXTENSION IF NOT EXISTS vector;