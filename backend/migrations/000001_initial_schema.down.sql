-- Drop triggers
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP TRIGGER IF EXISTS update_processors_updated_at ON processors;
DROP TRIGGER IF EXISTS update_destinations_updated_at ON destinations;
DROP TRIGGER IF EXISTS update_sources_updated_at ON sources;
DROP TRIGGER IF EXISTS update_configurations_updated_at ON configurations;
DROP TRIGGER IF EXISTS update_agents_updated_at ON agents;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS processors;
DROP TABLE IF EXISTS destinations;
DROP TABLE IF EXISTS sources;
DROP TABLE IF EXISTS configurations;
DROP TABLE IF EXISTS agents;
