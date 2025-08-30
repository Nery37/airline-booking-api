-- Init script for MySQL docker container
-- This file is mounted to /docker-entrypoint-initdb.d/ in the MySQL container

-- Set timezone to UTC
SET time_zone = '+00:00';

-- Create database if not exists (already created by environment variables)
-- CREATE DATABASE IF NOT EXISTS airline_booking;

-- Grant privileges to airline_user
GRANT ALL PRIVILEGES ON airline_booking.* TO 'airline_user'@'%';
FLUSH PRIVILEGES;
