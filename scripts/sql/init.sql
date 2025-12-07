-- Initialize the database for the User Management Service

-- Create database if not exists (for development)
-- This is mainly for documentation as docker-entrypoint-initdb.d runs after db creation

-- Ensure proper encoding and collation
-- ALTER DATABASE recipe_manager SET timezone TO 'UTC';

-- Create schema if not exists
CREATE SCHEMA IF NOT EXISTS public;

-- Grant permissions to the user
GRANT ALL PRIVILEGES ON SCHEMA public TO user_management;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO user_management;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO user_management;

-- Set default privileges for future objects
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO user_management;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO user_management;

-- Enable UUID extension for generating UUIDs
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Enable pgcrypto for additional cryptographic functions
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create indexes for common queries (will be managed by SQLAlchemy migrations)
-- This is mainly for documentation

-- Example indexes that might be created:
-- CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_email ON users(email);
-- CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_username ON users(username);
-- CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_is_active ON users(is_active);
-- CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_user_follows_follower_id ON user_follows(follower_id);
-- CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_user_follows_followed_id ON user_follows(followed_id);
-- CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
-- CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_notifications_is_read ON notifications(is_read);

-- Set up basic performance monitoring
-- Enable query statistics collection
SELECT pg_stat_statements_reset(); -- Reset statistics if extension is available

-- Log slow queries (>100ms) for development
-- ALTER SYSTEM SET log_min_duration_statement = 100;
-- ALTER SYSTEM SET log_statement = 'all';
-- SELECT pg_reload_conf();

COMMIT;
